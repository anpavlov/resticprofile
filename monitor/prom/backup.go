package prom

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Status uint

const (
	StatusFailed Status = iota
	StatusWarning
	StatusSuccess
)

type BackupMetrics struct {
	// duration         *prometheus.GaugeVec
	filesNew         prometheus.Gauge
	filesChanged     prometheus.Gauge
	filesUnmodified  prometheus.Gauge
	dirNew           prometheus.Gauge
	dirChanged       prometheus.Gauge
	dirUnmodified    prometheus.Gauge
	filesTotal       prometheus.Gauge
	bytesAdded       prometheus.Gauge
	bytesAddedPacked prometheus.Gauge
	bytesTotal       prometheus.Gauge
	// status           *prometheus.GaugeVec
	// time             *prometheus.GaugeVec
}

func newBackupMetrics() BackupMetrics {
	backupMetrics := BackupMetrics{
		filesNew: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "files_new",
			Help:      "Number of new files added to the backup.",
		}),
		filesChanged: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "files_changed",
			Help:      "Number of files with changes.",
		}),
		filesUnmodified: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "files_unmodified",
			Help:      "Number of files unmodified since last backup.",
		}),
		dirNew: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "dir_new",
			Help:      "Number of new directories added to the backup.",
		}),
		dirChanged: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "dir_changed",
			Help:      "Number of directories with changes.",
		}),
		dirUnmodified: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "dir_unmodified",
			Help:      "Number of directories unmodified since last backup.",
		}),
		filesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "files_processed",
			Help:      "Total number of files scanned by the backup for changes.",
		}),
		bytesAdded: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "added_bytes",
			Help:      "Total number of bytes added to the repository.",
		}),
		bytesAddedPacked: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "added_bytes_packed",
			Help:      "Total number of bytes added to the repository after compression.",
		}),
		bytesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: backup,
			Name:      "processed_bytes",
			Help:      "Total number of bytes scanned for changes.",
		}),
	}
	return backupMetrics
}
