package cmd

import (
	"testing"
)

func TestCompareThresholdsFail(t *testing.T) {
	undefined := int64(2)
	info := int64(2)
	low := int64(2)
	medium := int64(2)
	high := int64(2)
	critical := int64(2)
	scanThresholds := map[string]*int64{
		"UNDEFINED":     &undefined,
		"INFORMATIONAL": &info,
		"LOW":           &low,
		"MEDIUM":        &medium,
		"HIGH":          &high,
		"CRITICAL":      &critical,
	}
	allowedThresholds := map[string]int64{
		"UNDEFINED":     2,
		"INFORMATIONAL": 1,
		"LOW":           1,
		"MEDIUM":        1,
		"HIGH":          1,
		"CRITICAL":      1,
	}
	failedScan, _ := compareThresholds(allowedThresholds, scanThresholds)
	if failedScan {
		t.Logf("PASSED: compareThresholds returned true indicating error when comparing results")
	} else {
		t.Errorf("FAILED: compareThresholds should have returned true but returned false instead.")
	}
}
