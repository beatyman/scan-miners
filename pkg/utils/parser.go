package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseHashrate parses string like "306.23 TH/s" into value (306.23) and unit ("TH/s")
func ParseHashrate(hs string) (float64, string) {
	if hs == "" {
		return 0, ""
	}
	parts := strings.Split(strings.TrimSpace(hs), " ")
	if len(parts) < 2 {
		// Try to parse just the number if no unit
		val, err := strconv.ParseFloat(hs, 64)
		if err == nil {
			return val, ""
		}
		return 0, ""
	}
	
	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, parts[1]
	}
	return val, parts[1]
}

// GenerateIP generates IP from workerId like "30x182" -> "172.16.30.182"
func GenerateIP(workerID string) string {
	// Rule: 172.16.{x}.{y} where workerId is {x}x{y}
	parts := strings.Split(workerID, "x")
	if len(parts) != 2 {
		return "" // Invalid format
	}
	return fmt.Sprintf("172.16.%s.%s", parts[0], parts[1])
}
