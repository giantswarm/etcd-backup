package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/coreos/go-semver/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
)

// Create temporary directory where all file related magic happens.
func CreateTMPDir() (string, error) {

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", microerror.Mask(err)
	}
	log.Print("Created temporary directory: ", tmpDir)

	return tmpDir, nil
}

//clear temporary directory
func ClearTMPDir(tmpDir string) {
	os.RemoveAll(tmpDir)
}

// create k8s client to access host k8s cluster
func CreateK8sClient(logger micrologger.Logger) (kubernetes.Interface, error) {
	guestClientConfig := k8sclient.Config{}
	{
		guestClientConfig.InCluster = true
		guestClientConfig.Logger = logger
	}

	k8sClient, err := k8sclient.New(guestClientConfig)
	if err != nil {
		return nil, microerror.Maskf(err, "error creating k8sclient for host cluster")
	}
	return k8sClient, nil
}

// create CRD client to access k8s crd resources cluster
func CreateCRDClient(logger micrologger.Logger) (*versioned.Clientset, error) {
	var err error
	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger:    logger,
			InCluster: true,
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if err != nil {
		return nil, microerror.Maskf(err, "error creating crd client for host cluster")
	}
	return g8sClient, nil
}

// check if cluster release version has guest cluster backup support
func CheckClusterVersionSupport(clusterID string, provider string, crdCLient *versioned.Clientset) (bool, error) {
	getOpts := metav1.GetOptions{}

	switch provider {
	case aws:
		{
			crd, err := crdCLient.ProviderV1alpha1().AWSConfigs(crdNamespace).Get(clusterID, getOpts)
			if err != nil {
				return false, microerror.Maskf(err, "failed to get aws crd"+clusterID)
			}
			crdVersionStr := crd.Spec.VersionBundle.Version
			if crdVersionStr == "" {
				crdVersionStr = "0.0.0"
			}
			crdVersion := semver.New(crdVersionStr)
			if crdVersion.Compare(*awsSupportFrom) >= 0 {
				// version has support
				return true, nil
			} else {
				// version doesnt have support
				return false, nil
			}
		}
	case azure:
		{
			crd, err := crdCLient.ProviderV1alpha1().AzureConfigs(crdNamespace).Get(clusterID, getOpts)
			if err != nil {
				return false, microerror.Maskf(err, "failed to get azure crd "+clusterID)
			}
			crdVersionStr := crd.Spec.VersionBundle.Version
			if crdVersionStr == "" {
				crdVersionStr = "0.0.0"
			}
			crdVersion := semver.New(crdVersionStr)
			if crdVersion.Compare(*azureSupportFrom) >= 0 {
				// version has support
				return true, nil
			} else {
				// version doesnt have support
				return false, nil
			}
		}
	case kvm:
		{
			// kvm backups are always supported
			return true, nil
		}
	}
	return false, nil
}

// fetch all guest clusters ids in host cluster
func GetAllGuestClusters(provider string, crdCLient *versioned.Clientset) ([]string, error) {
	var clusterList []string
	listOpt := metav1.ListOptions{}

	switch provider {
	case aws:
		{
			crdList, err := crdCLient.ProviderV1alpha1().AWSConfigs(crdNamespace).List(listOpt)
			if err != nil {
				return []string{}, microerror.Maskf(err, "failed to list aws crd")
			}
			for _, awsConfig := range crdList.Items {
				fmt.Sprintf("deletion timestamp check %#v", awsConfig.DeletionTimestamp)
				clusterList = append(clusterList, awsConfig.Name)
			}
			break
		}
	case azure:
		{
			crdList, err := crdCLient.ProviderV1alpha1().AzureConfigs(crdNamespace).List(listOpt)
			if err != nil {
				return []string{}, microerror.Maskf(err, "failed to list azure crd")
			}
			for _, azureConfig := range crdList.Items {
				clusterList = append(clusterList, azureConfig.Name)
			}
			break
		}
	case kvm:
		{
			crdList, err := crdCLient.ProviderV1alpha1().KVMConfigs(crdNamespace).List(listOpt)
			if err != nil {
				return []string{}, microerror.Maskf(err, "failed to list azure crd")
			}
			for _, kvmConfig := range crdList.Items {
				clusterList = append(clusterList, kvmConfig.Name)
			}
			break
		}
	default:
		{
			return []string{}, invalidProviderError
		}
	}

	return clusterList, nil
}

// fetch etcd client certs
func FetchCerts(clusterID string, k8sClient kubernetes.Interface) (*k8sclient.TLSClientConfig, error) {

	getOpts := metav1.GetOptions{}
	secret, err := k8sClient.CoreV1().Secrets(secretNamespace).Get(clusterID+"-etcd", getOpts)
	if err != nil {
		return nil, microerror.Maskf(err, "error getting etcd client certificates for guest cluster %s", clusterID)
	}

	certs := &k8sclient.TLSClientConfig{
		CAData:  secret.Data["ca"],
		KeyData: secret.Data["key"],
		CrtData: secret.Data["crt"],
	}

	return certs, nil
}

// fetch guest cluster etcd endpoint
func GetEtcdEndpoint(clusterID string, provider string, crdCLient *versioned.Clientset) (string, error) {
	getOpts := metav1.GetOptions{}
	var etcdEndpoint string

	switch provider {
	case aws:
		{
			crd, err := crdCLient.ProviderV1alpha1().AWSConfigs(crdNamespace).Get(clusterID, getOpts)
			if err != nil {
				fmt.Println()
				return "", microerror.Maskf(err, "error getting aws crd for guest cluster %s", clusterID)
			}
			etcdEndpoint = AwsEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	case azure:
		{
			crd, err := crdCLient.ProviderV1alpha1().AzureConfigs(crdNamespace).Get(clusterID, getOpts)
			if err != nil {
				fmt.Println()
				return "", microerror.Maskf(err, "error getting azure crd for guest cluster %s", clusterID)
			}
			etcdEndpoint = AzureEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	case kvm:
		{
			crd, err := crdCLient.ProviderV1alpha1().KVMConfigs(crdNamespace).Get(clusterID, getOpts)
			if err != nil {
				fmt.Println()
				return "", microerror.Maskf(err, "error getting kvm crd for guest cluster %s", clusterID)
			}
			etcdEndpoint = KVMEtcdEndpoint(crd.Spec.Cluster.Etcd.Domain)
			break
		}
	}

	// we already check for unknown provider at the start
	return etcdEndpoint, nil
}

// create cert files in tmp dir from certConfig and saves filenames back
func CreateCertFiles(clusterID string, certConfig *k8sclient.TLSClientConfig, tmpDir string) error {
	// cert
	err := ioutil.WriteFile(CertFile(clusterID, tmpDir), certConfig.CrtData, fileMode)
	if err != nil {
		return microerror.Maskf(err, "Failed to write crt file "+CertFile(clusterID, tmpDir))
	}
	certConfig.CrtFile = CertFile(clusterID, tmpDir)

	// key
	err = ioutil.WriteFile(KeyFile(clusterID, tmpDir), certConfig.KeyData, fileMode)
	if err != nil {
		return microerror.Maskf(err, "Failed to write key file "+KeyFile(clusterID, tmpDir))
	}
	certConfig.KeyFile = KeyFile(clusterID, tmpDir)

	// ca
	err = ioutil.WriteFile(CAFile(clusterID, tmpDir), certConfig.CAData, fileMode)
	if err != nil {
		return microerror.Maskf(err, "Failed to write ca file "+CAFile(clusterID, tmpDir))
	}
	certConfig.CAFile = CAFile(clusterID, tmpDir)

	return nil
}
