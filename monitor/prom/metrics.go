package prom

import (
	"runtime"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
)

const (
	namespace      = "resticprofile"
	backup         = "backup"
	groupLabel     = "group"
	commandLabel   = "command"
	profileLabel   = "profile"
	goVersionLabel = "goversion"
	versionLabel   = "version"
)

type Metrics struct {
	// labels   prometheus.Labels
	registry *prometheus.Registry

	info       *prometheus.GaugeVec
	resticInfo *prometheus.GaugeVec

	commandStatus   *prometheus.GaugeVec
	commandDuration *prometheus.GaugeVec
	commandTime     *prometheus.GaugeVec

	backup BackupMetrics
	pusher *push.Pusher
}

func NewMetrics(profile, group, version string, resticversion string,
	pushgateway string, pushformat string, configLabels map[string]string) *Metrics {
	// default labels for all metrics
	labels := prometheus.Labels{profileLabel: profile}
	if group != "" {
		labels[groupLabel] = group
	}
	labels = mergeLabels(labels, configLabels)
	// keys := slices.Collect(maps.Keys(labels))

	registry := prometheus.NewRegistry()
	p := &Metrics{
		// labels:   labels,
		registry: registry,
	}
	p.info = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "resticprofile build information.",
	}, []string{goVersionLabel, versionLabel})
	// send the information about the build right away
	p.info.With(map[string]string{goVersionLabel: runtime.Version(), versionLabel: version}).Set(1)

	p.resticInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_build_info",
		Help: "restic build information.",
	}, []string{versionLabel})
	// send the information about the build right away
	p.resticInfo.With(map[string]string{versionLabel: resticversion}).Set(1)

	p.backup = newBackupMetrics()

	p.commandDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "command",
		Name:      "duration_seconds",
		Help:      "Command execute duration (in seconds).",
	}, []string{commandLabel})

	p.commandStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "command",
		Name:      "status",
		Help:      "Command execute status: 0=fail, 1=warning, 2=success.",
	}, []string{commandLabel})

	p.commandTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "command",
		Name:      "time_seconds",
		Help:      "Last command run timestamp (unixtime).",
	}, []string{commandLabel})

	registry.MustRegister(
		p.info,
		p.resticInfo,
		p.backup.filesNew,
		p.backup.filesChanged,
		p.backup.filesUnmodified,
		p.backup.dirNew,
		p.backup.dirChanged,
		p.backup.dirUnmodified,
		p.backup.filesTotal,
		p.backup.bytesAdded,
		p.backup.bytesAddedPacked,
		p.backup.bytesTotal,
		p.commandDuration,
		p.commandStatus,
		p.commandTime,
	)

	var expFmt expfmt.Format
	if pushformat == "protobuf" {
		expFmt = expfmt.NewFormat(expfmt.TypeProtoDelim)
	} else {
		expFmt = expfmt.NewFormat(expfmt.TypeTextPlain)
	}

	p.pusher = push.New(pushgateway, "resticprofile").
		Format(expFmt).
		Gatherer(p.registry)

	for k, v := range labels {
		p.pusher = p.pusher.Grouping(k, v)
	}

	return p
}

func (p *Metrics) Results(command string, status Status, summary monitor.Summary) {
	p.commandDuration.With(prometheus.Labels{commandLabel: command}).Set(summary.Duration.Seconds())
	p.commandStatus.With(prometheus.Labels{commandLabel: command}).Set(float64(status))
	p.commandTime.With(prometheus.Labels{commandLabel: command}).Set(float64(time.Now().Unix()))

	if command == constants.CommandBackup {
		p.backupResults(summary)
	}
}

func (p *Metrics) backupResults(summary monitor.Summary) {
	p.backup.filesNew.Set(float64(summary.FilesNew))
	p.backup.filesChanged.Set(float64(summary.FilesChanged))
	p.backup.filesUnmodified.Set(float64(summary.FilesUnmodified))

	p.backup.dirNew.Set(float64(summary.DirsNew))
	p.backup.dirChanged.Set(float64(summary.DirsChanged))
	p.backup.dirUnmodified.Set(float64(summary.DirsUnmodified))

	p.backup.filesTotal.Set(float64(summary.FilesTotal))
	p.backup.bytesAdded.Set(float64(summary.BytesAdded))
	p.backup.bytesAddedPacked.Set(float64(summary.BytesAddedPacked))
	p.backup.bytesTotal.Set(float64(summary.BytesTotal))
}

func (p *Metrics) SaveTo(filename string) error {
	return prometheus.WriteToTextfile(filename, p.registry)
}

func (p *Metrics) Push( /*url, format, jobName string*/ ) error {
	// var expFmt expfmt.Format

	// if format == "protobuf" {
	// 	expFmt = expfmt.NewFormat(expfmt.TypeProtoDelim)
	// } else {
	// 	expFmt = expfmt.NewFormat(expfmt.TypeTextPlain)
	// }

	return p.pusher.Add()
}

func mergeLabels(labels prometheus.Labels, add map[string]string) prometheus.Labels {
	for key, value := range add {
		labels[key] = value
	}
	return labels
}

func cloneLabels(labels prometheus.Labels) prometheus.Labels {
	clone := make(prometheus.Labels, len(labels))
	for key, value := range labels {
		clone[key] = value
	}
	return clone
}
