package metrics

import (
	"github.com/giantswarm/etcd-backup/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const labelTenantClusterId = "tenant_cluster_id"

var (
	labels = []string{
		labelTenantClusterId,
	}
	namespace    = "etcd_backup"
	creationTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "creation_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup creation process.",
	}, labels)
	encryptionTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "encryption_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup encryption process.",
	}, labels)
	uploadTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "upload_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup upload process.",
	}, labels)
	backupSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "size_bytes"),
		Help: "Gauge about the size of the backup file, as seen by S3.",
	}, labels)
	successCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: prometheus.BuildFQName(namespace, "", "success_count"),
		Help: "Count of successful backups",
	}, labels)
	failureCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: prometheus.BuildFQName(namespace, "", "failure_count"),
		Help: "Count of failed backups",
	}, labels)
)

func Send(prometheusConfig *config.PrometheusConfig, metrics *BackupMetrics, tenantClusterName string) error {
	// prometheus URL might be empty, in that case we can't push any metric
	if prometheusConfig.Url != "" {
		registry := prometheus.NewRegistry()

		labels := prometheus.Labels{
			labelTenantClusterId: tenantClusterName,
		}

		if metrics.Successful {
			// successful backup
			registry.MustRegister(creationTime, encryptionTime, uploadTime, backupSize, successCounter)
			pusher := push.New(prometheusConfig.Url, prometheusConfig.Job).Gatherer(registry)

			creationTime.With(labels).Set(float64(metrics.CreationTimeMeasurement))
			encryptionTime.With(labels).Set(float64(metrics.EncryptionTimeMeasurement))
			uploadTime.With(labels).Set(float64(metrics.UploadTimeMeasurement))
			backupSize.With(labels).Set(float64(metrics.BackupSizeMeasurement))
			successCounter.With(labels).Inc()

			if err := pusher.Add(); err != nil {
				return err
			}
		} else {
			// failed backup
			registry.MustRegister(failureCounter)
			pusher := push.New(prometheusConfig.Url, prometheusConfig.Job).Gatherer(registry)

			failureCounter.With(labels).Inc()

			if err := pusher.Add(); err != nil {
				return err
			}
		}
	}

	return nil
}
