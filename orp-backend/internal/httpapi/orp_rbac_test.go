package httpapi

import "testing"

func TestMatchesRoleFilter(t *testing.T) {
	roleIDs := []int64{2, 7}
	tests := []struct {
		name   string
		filter string
		want   bool
	}{
		{name: "no filter", want: true},
		{name: "all sentinel", filter: "all", want: true},
		{name: "matching role", filter: "7", want: true},
		{name: "non matching role", filter: "9", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesRoleFilter(roleIDs, tt.filter); got != tt.want {
				t.Fatalf("matchesRoleFilter(%v, %q) = %v, want %v", roleIDs, tt.filter, got, tt.want)
			}
		})
	}
}
