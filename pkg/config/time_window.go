package config

import (
	"fmt"
	"log"
	"time"
)

// ParseTimeWindow parses the time strings from environment variables and returns start and stop times.
func ParseTimeWindow(startStr, stopStr string) (time.Time, time.Time, error) {
	layout := "1504" // 24-hour format (HHMM)
	start, err := time.Parse(layout, startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid START_TIME: %v", err)
	}

	stop, err := time.Parse(layout, stopStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid STOP_TIME: %v", err)
	}

	return start, stop, nil
}

// WaitForNextActivePeriod waits until the next active time window starts.
func WaitForNextActivePeriod(start, stop time.Time) {
	now := time.Now().UTC()
	nextStart := time.Date(now.Year(), now.Month(), now.Day(), start.Hour(), start.Minute(), 0, 0, time.UTC)

	if now.After(nextStart) {
		nextStart = nextStart.Add(24 * time.Hour)
	}

	duration := nextStart.Sub(now)
	log.Printf("Waiting for %v until next active period at %v", duration, nextStart)
	time.Sleep(duration)
}

// WaitForNextStalePeriod waits until the time window ends.
func WaitForNextStalePeriod(start, stop time.Time) {
	now := time.Now().UTC()
	nextStop := time.Date(now.Year(), now.Month(), now.Day(), stop.Hour(), stop.Minute(), 0, 0, time.UTC)

	if now.After(nextStop) {
		nextStop = nextStop.Add(24 * time.Hour)
	}

	duration := nextStop.Sub(now)
	log.Printf("Waiting for %v until next stale period at %v", duration, nextStop)
	time.Sleep(duration)
}
