package cmd

import (
	log "github.com/sirupsen/logrus"
)

// compareThresholds: Compares the scan results to the allowed thresholds
func compareThresholds(allowedThresholds map[string]int64, results map[string]*int64) (bool, []string) {
	var failedScan bool
	var failedLevels []string
	for level, value := range results {
		log.Infof("%v: Allowed number of vulnerabilities: %v, Amount of vulnerabilities: %v", level, allowedThresholds[level], *value)
		if *value >= allowedThresholds[level] {
			failedScan = true
			failedLevels = append(failedLevels, level)
		}
	}
	return failedScan, failedLevels
}
