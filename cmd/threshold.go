package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// compareThresholds: Compares the scan results to the allowed thresholds
func compareThresholds(results map[string]*int64) (bool, []string) {
	var failedScan bool
	var failedLevels []string
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
			failedLevels = append(failedLevels, level)
		}
	}
	return failedScan, failedLevels
}
