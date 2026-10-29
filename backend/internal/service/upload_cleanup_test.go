package service

import (
	"testing"
	"time"
)

func TestUploadRetention(t *testing.T) {
	for unit, duration := range map[string]time.Duration{
		"second": time.Second, "minute": time.Minute, "hour": time.Hour,
		"day": 24 * time.Hour, "week": 7 * 24 * time.Hour, "month": 30 * 24 * time.Hour,
	} {
		t.Run(unit, func(t *testing.T) {
			got, err := uploadRetention("7", unit)
			if err != nil || got != 7*duration {
				t.Fatalf("got %v, %v; want %v", got, err, 7*duration)
			}
		})
	}
	for _, value := range []string{"", "0", "-1", "1.5", "NaN", "9223372036854775807", "999999999999999999999999"} {
		if _, err := uploadRetention(value, "second"); err == nil {
			t.Errorf("accepted invalid value %q", value)
		}
	}
	if _, err := uploadRetention("1", "year"); err == nil {
		t.Fatal("accepted unknown unit")
	}
	if _, err := uploadRetention("106752", "day"); err == nil {
		t.Fatal("accepted duration overflow")
	}
}
