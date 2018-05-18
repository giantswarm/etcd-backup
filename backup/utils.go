package backup

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

// fetch all guest clusters ids in host cluster
func GetAllGuestClusters(provider string, crdCLient *versioned.Clientset) ([]string, error) {
	var clusterList []string
	listOpt := metav1.ListOptions{}

	if provider == aws {
		crdList, err := crdCLient.CoreV1alpha1().AWSClusterConfigs(crdNamespace).List(listOpt)
		if err != nil {
			return []string{}, microerror.Maskf(err, "failed to list aws crd")
		}
		for _, awsConfig := range crdList.Items {
			clusterList = append(clusterList, awsConfig.Name)
		}

	} else if provider == azure {
		crdList, err := crdCLient.CoreV1alpha1().AzureClusterConfigs(crdNamespace).List(listOpt)
		if err != nil {
			return []string{}, microerror.Maskf(err, "failed to list azure crd")
		}
		for _, azureConfig := range crdList.Items {
			clusterList = append(clusterList, azureConfig.Name)
		}

	} else if provider == kvm {
		crdList, err := crdCLient.CoreV1alpha1().KVMClusterConfigs(crdNamespace).List(listOpt)
		if err != nil {
			return []string{}, microerror.Maskf(err, "failed to list azure crd")
		}
		for _, kvmConfig := range crdList.Items {
			clusterList = append(clusterList, kvmConfig.Name)
		}
	} else {
		return []string{}, microerror.Mask(invalidProviderError)
	}

	return clusterList, nil
}

// fetch etcd client certs
func FetchCerts(clusterID string, k8sClient kubernetes.Interface) (*k8sclient.TLSClientConfig, error) {

	certs := &k8sclient.TLSClientConfig{}
	{
		getOpts := metav1.GetOptions{}
		secret, err := k8sClient.CoreV1().Secrets(secretNamespace).Get(clusterID+"-etcd", getOpts)
		if err != nil {
			return nil, microerror.Maskf(err, "error getting etcd client certificates for guest cluster %s", clusterID)
		}
		certs.CAData = secret.Data["ca"]
		certs.KeyData = secret.Data["key"]
		certs.KeyData = secret.Data["crt"]
	}

	return certs, nil
}

// fetch guest cluster etcd endpoint
func GetEtcdEndpoint(clusterID string, k8sClient kubernetes.Interface) (string, error) {
	getOpts := metav1.GetOptions{}
	ingress, err := k8sClient.ExtensionsV1beta1().Ingresses(clusterID).Get("etcd", getOpts)
	if err != nil {
		fmt.Println()
		return "", microerror.Maskf(err, "error getting etcd endpoint for guest cluster %s", clusterID)
	}
	return ingress.Spec.Rules[0].Host, nil
}
