package etcd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	namespace    = "etcd_backup"
	creationTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "creation_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup creation process.",
	})
	encryptionTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "encryption_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup encryption process.",
	})
	uploadTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "upload_time_ms"),
		Help: "Gauge about the time in ms spent by the ETCD backup upload process.",
	})
	backupSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, "", "size_bytes"),
		Help: "Gauge about the size of the backup file, as seen by S3.",
	})
)

func sendMetrics(prometheusConfig PrometheusConfig, creationTimeMeasurement int64, encryptionTimeMeasurement int64, uploadTimeMeasurement int64, backupSizeMeasurement int64) error {
	// prometheus URL might be empty, in that case we can't push any metric
	if prometheusConfig.Url != "" {
		registry := prometheus.NewRegistry()
		registry.MustRegister(creationTime, encryptionTime, uploadTime, backupSize)

		pusher := push.New(prometheusConfig.Url, prometheusConfig.Job).Gatherer(registry)

		creationTime.Set(float64(creationTimeMeasurement))
		encryptionTime.Set(float64(encryptionTimeMeasurement))
		uploadTime.Set(float64(uploadTimeMeasurement))
		backupSize.Set(float64(backupSizeMeasurement))

		if err := pusher.Add(); err != nil {
			return err
		}
	}

	return nil
}
