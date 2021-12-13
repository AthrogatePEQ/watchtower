package notifications

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	t "github.com/containrrr/watchtower/pkg/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type InfluxdbStat struct {
	hostname         string
	measurement      string
	scanned          int
	fresh            int
	updated          int
	failed           int
	skipped          int
	stale            int
	stale_containers string
}

var (
	influxdbHost            string
	influxdbAPI             string
	influxdbAuth            string
	influxdbDatabase        string
	influxdbRetentionPolicy string
	influxdbMeasurement     string
	influxdbHostnameTag     string
	influxdbClient          *http.Client
)

func NewInfluxdbStat(report t.Report) *InfluxdbStat {
	staleContainers := ""
	for _, container := range report.Stale() {
		if staleContainers != "" {
			staleContainers += ", "
		}
		staleContainers += container.Name()
	}
	return &InfluxdbStat{
		hostname:         influxdbHostnameTag,
		measurement:      influxdbMeasurement,
		scanned:          len(report.Scanned()),
		updated:          len(report.Updated()),
		failed:           len(report.Failed()),
		skipped:          len(report.Skipped()),
		stale:            len(report.Stale()),
		fresh:            len(report.Fresh()),
		stale_containers: staleContainers,
	}

}

func InitInfluxdbNotifier(cmd *cobra.Command) error {
	f := cmd.PersistentFlags()
	influxdbHostnameTag, _ = f.GetString("notifications-hostname")

	if influxdbHostnameTag == "" {
		influxdbHostnameTag, _ = os.Hostname()
	}

	influxdbHost, _ = f.GetString("influxdb-host")
	if influxdbHost == "" {
		return fmt.Errorf("%v is a required when --influxdb is supplied", "--influxdb-host")
	}
	influxdbAuth, _ = f.GetString("influxdb-auth")
	influxdbDatabase, _ = f.GetString("influxdb-database")
	if influxdbDatabase == "" {
		return fmt.Errorf("%v is a required when --influxdb is supplied", "--influxdb-database")
	}

	influxdbRetentionPolicy, _ = f.GetString("influxdb-retention-policy")
	if influxdbRetentionPolicy == "" {
		influxdbRetentionPolicy = "autogen"
	}

	influxdbMeasurement, _ = f.GetString("influxdb-measurement")
	if influxdbMeasurement == "" {
		return fmt.Errorf("%v is a required when --influxdb is supplied", "--influxdb-measurement")
	}

	influxdbClient = &http.Client{}

	return nil
}

func UpdateInfluxdbStats(report t.Report) {

	stat := NewInfluxdbStat(report)

	log.Debug(log.Fields{
		"host":             stat.hostname,
		"scanned":          stat.scanned,
		"updated":          stat.updated,
		"failed":           stat.failed,
		"skipped":          stat.skipped,
		"stale":            stat.stale,
		"fresh":            stat.fresh,
		"stale-containers": stat.stale_containers,
	})

	// write point immediately

	body := stat.measurement + ",host=" + stat.hostname + " scanned=" + strconv.Itoa(stat.scanned) + ",updated=" + strconv.Itoa(stat.updated) + ",failed=" + strconv.Itoa(stat.failed) + ",skipped=" + strconv.Itoa(stat.skipped) + ",fresh=" + strconv.Itoa(stat.fresh) + ",stale=" + strconv.Itoa(stat.stale) + ",stale_containers=\"" + stat.stale_containers + "\""
	reader := strings.NewReader(body)

	// v2.0
	//url := proto + hostname + ":" + port + path + "?db=" + org + "/" + bucket

	// v1.8
	url := influxdbHost + influxdbAPI + "?db=" + influxdbDatabase + "&rp=" + influxdbRetentionPolicy

	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := influxdbClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	if resp != nil {
		log.Debug(resp.Status)
		log.Debug(resp.Body)
	}
}
