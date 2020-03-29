package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func compareThresholds(results map[string]*int64) {
	var failedScan bool
	allowedThresholds := map[string]int64{
		"INFORMATIONAL": viper.GetInt64("info"),
		"LOW":           viper.GetInt64("low"),
		"MEDIUM":        viper.GetInt64("medium"),
		"HIGH":          viper.GetInt64("high"),
		"CRITICAL":      viper.GetInt64("critical"),
	}
	for level, value := range results {
		log.Infof("%v: Allowed number of vulnerabilities: %v, Amount of vulnerabilities: %v", level, allowedThresholds[level], *value)
		if *value > allowedThresholds[level] {
			failedScan = true
		}
	}
	if failedScan {
		log.Fatalf("Scan failed due to exceeding threshold levels.")
	} else {
		log.Info("Scan passed!")
	}
}
