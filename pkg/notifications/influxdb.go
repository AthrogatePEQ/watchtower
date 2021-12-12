package notifications

import (
	"context"
	"fmt"
	"os"
	"time"

	t "github.com/containrrr/watchtower/pkg/types"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/spf13/cobra"
        log "github.com/sirupsen/logrus"
)

var (
	influxdbHost            string
	influxdbAuth            string
	influxdbDatabase        string
	influxdbRetentionPolicy string
	influxdbMeasurement     string
	influxdbHostnameTag     string
	influxdbClient          influxdb2.Client
	influxdbWiteAPI         api.WriteAPIBlocking
)

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

	influxdbClient = influxdb2.NewClient(influxdbHost, influxdbAuth)
	influxdbWiteAPI = influxdbClient.WriteAPIBlocking("", influxdbDatabase+"/"+influxdbRetentionPolicy)

	return nil
}

func UpdateInfluxdbStats(report t.Report) {

	staleContainers := ""

	for _, container := range report.Stale() {
		if staleContainers != "" {
			staleContainers += ", "
		}
		staleContainers += container.Name()
	}

	// Create point using fluent style
	p := influxdb2.NewPointWithMeasurement(influxdbMeasurement).
		AddTag("host", influxdbHostnameTag).
		AddField("scanned", len(report.Scanned())).
		AddField("updated", len(report.Updated())).
		AddField("failed", len(report.Failed())).
		AddField("skipped", len(report.Skipped())).
		AddField("stale", len(report.Stale())).
		AddField("fresh", len(report.Fresh())).
		AddField("stale-containers", staleContainers).
		SetTime(time.Now())

        log.Debug(log.Fields{
           "host": influxdbHostnameTag,
           "scanned": len(report.Scanned()),
	   "updated": len(report.Updated()),
	   "failed": len(report.Failed()),
	   "skipped": len(report.Skipped()),
	   "stale": len(report.Stale()),
	   "fresh": len(report.Fresh()),
	   "stale-containers": staleContainers,
        })

	// write point immediately
	influxdbWiteAPI.WritePoint(context.Background(), p)

}
