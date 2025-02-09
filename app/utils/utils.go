package utils

import (
	"fmt"
	"strings"
	"time"
)

func ExtractDurationInSecs(duration string) int64 {
	parsedDuration, err := time.ParseDuration(strings.ToLower(strings.ReplaceAll(duration, "PT", "")))
	if err != nil {
		fmt.Printf("Error parsing duration %s => %v\n", duration, err)
		return 0
	}
	return int64(parsedDuration.Seconds())
}
