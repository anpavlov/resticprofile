package prom

import (
	"maps"
	"runtime"
	"slices"
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
	labels   prometheus.Labels
	registry *prometheus.Registry

	info       *prometheus.GaugeVec
	resticInfo *prometheus.GaugeVec

	commandStatus   *prometheus.GaugeVec
	commandDuration *prometheus.GaugeVec
	commandTime     *prometheus.GaugeVec

	backup BackupMetrics
}

func NewMetrics(profile, group, version string, resticversion string, configLabels map[string]string) *Metrics {
	// default labels for all metrics
	labels := prometheus.Labels{profileLabel: profile}
	if group != "" {
		labels[groupLabel] = group
	}
	labels = mergeLabels(labels, configLabels)
	keys := slices.Collect(maps.Keys(labels))

	registry := prometheus.NewRegistry()
	p := &Metrics{
		labels:   labels,
		registry: registry,
	}
	p.info = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "resticprofile build information.",
	}, append(keys, goVersionLabel, versionLabel))
	// send the information about the build right away
	p.info.With(mergeLabels(cloneLabels(labels), map[string]string{goVersionLabel: runtime.Version(), versionLabel: version})).Set(1)

	p.resticInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_build_info",
		Help: "restic build information.",
	}, append(keys, versionLabel))
	// send the information about the build right away
	p.resticInfo.With(mergeLabels(cloneLabels(labels), map[string]string{versionLabel: resticversion})).Set(1)

	p.backup = newBackupMetrics(keys)

	p.commandDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "command",
		Name:      "duration_seconds",
		Help:      "Command execute duration (in seconds).",
	}, append(keys, commandLabel))

	p.commandStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "command",
		Name:      "status",
		Help:      "Command execute status: 0=fail, 1=warning, 2=success.",
	}, append(keys, commandLabel))

	p.commandTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "command",
		Name:      "time_seconds",
		Help:      "Last command run timestamp (unixtime).",
	}, append(keys, commandLabel))

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
	return p
}

func (p *Metrics) Results(command string, status Status, summary monitor.Summary) {
	p.commandDuration.MustCurryWith(prometheus.Labels{commandLabel: command}).With(p.labels).Set(summary.Duration.Seconds())
	p.commandStatus.MustCurryWith(prometheus.Labels{commandLabel: command}).With(p.labels).Set(float64(status))
	p.commandTime.MustCurryWith(prometheus.Labels{commandLabel: command}).With(p.labels).Set(float64(time.Now().Unix()))

	if command == constants.CommandBackup {
		p.backupResults(summary)
	}
}

func (p *Metrics) backupResults(summary monitor.Summary) {
	p.backup.filesNew.With(p.labels).Set(float64(summary.FilesNew))
	p.backup.filesChanged.With(p.labels).Set(float64(summary.FilesChanged))
	p.backup.filesUnmodified.With(p.labels).Set(float64(summary.FilesUnmodified))

	p.backup.dirNew.With(p.labels).Set(float64(summary.DirsNew))
	p.backup.dirChanged.With(p.labels).Set(float64(summary.DirsChanged))
	p.backup.dirUnmodified.With(p.labels).Set(float64(summary.DirsUnmodified))

	p.backup.filesTotal.With(p.labels).Set(float64(summary.FilesTotal))
	p.backup.bytesAdded.With(p.labels).Set(float64(summary.BytesAdded))
	p.backup.bytesAddedPacked.With(p.labels).Set(float64(summary.BytesAddedPacked))
	p.backup.bytesTotal.With(p.labels).Set(float64(summary.BytesTotal))
}

func (p *Metrics) SaveTo(filename string) error {
	return prometheus.WriteToTextfile(filename, p.registry)
}

func (p *Metrics) Push(url, format, jobName string) error {
	var expFmt expfmt.Format

	if format == "protobuf" {
		expFmt = expfmt.NewFormat(expfmt.TypeProtoDelim)
	} else {
		expFmt = expfmt.NewFormat(expfmt.TypeTextPlain)
	}

	return push.New(url, jobName).
		Format(expFmt).
		Gatherer(p.registry).
		Add()
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
