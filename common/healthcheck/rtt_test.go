package healthcheck_test

import (
	"testing"

	"github.com/sagernet/sing-box/common/healthcheck"
)

func TestDuration(t *testing.T) {
	tests := []struct {
		value healthcheck.RTT
		want  string
	}{
		{healthcheck.Failed, "0ms"},
		{healthcheck.RTT(1), "1ms"},
		{healthcheck.RTT(1000), "1000ms"},
		{healthcheck.RTT(1101), "1.10s"},
	}
	for _, tt := range tests {
		if got := tt.value.String(); got != tt.want {
			t.Errorf("Duration.String() = %v, want %v", got, tt.want)
		}
	}
}
