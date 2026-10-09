package httpapi

import (
	"strings"
	"testing"
)

func TestAvailabilityPercent(t *testing.T) {
	tests := []struct {
		name       string
		successful int64
		total      int64
		want       *float64
	}{
		{name: "no samples is unknown", successful: 0, total: 0},
		{name: "all probes succeeded", successful: 12, total: 12, want: floatPointer(100)},
		{name: "rounds to two decimal places", successful: 6249, total: 6250, want: floatPointer(99.98)},
		{name: "all probes failed", successful: 0, total: 8, want: floatPointer(0)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := availabilityPercent(test.successful, test.total)
			if test.want == nil {
				if got != nil {
					t.Fatalf("availabilityPercent(%d, %d) = %v, want nil", test.successful, test.total, *got)
				}
				return
			}
			if got == nil || *got != *test.want {
				t.Fatalf("availabilityPercent(%d, %d) = %v, want %v", test.successful, test.total, got, *test.want)
			}
		})
	}
}

func floatPointer(value float64) *float64 {
	return &value
}

func TestDashboardSnapshotCacheKeyIsScopedAndDoesNotExposeUser(t *testing.T) {
	first := dashboardSnapshotKey("range=today&centerId=2", "alice")
	if strings.Contains(first, "alice") {
		t.Fatalf("cache key exposes username: %q", first)
	}
	if first == dashboardSnapshotKey("range=today&centerId=2", "bob") {
		t.Fatal("different users must not share dashboard snapshots")
	}
	if first != dashboardSnapshotKey("centerId=2&range=today", "alice") {
		t.Fatal("equivalent query parameters should use the same dashboard cache key")
	}
}

func TestDashboardInputKeyKeepsOnlyDashboardFilters(t *testing.T) {
	got := dashboardInputKey("ignored=x&range=today&centerId=2")
	if got != "range=today&centerId=2" {
		t.Fatalf("dashboardInputKey() = %q", got)
	}
}
