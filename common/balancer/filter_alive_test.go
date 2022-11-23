package balancer_test

import (
	"testing"
	"time"

	"github.com/sagernet/sing-box/common/balancer"
	"github.com/sagernet/sing-box/common/healthcheck"
	"github.com/sagernet/sing-box/option"
)

func TestFilterAlive(t *testing.T) {
	options := option.LoadBalancePickOptions{
		MaxFail: 2,
		MaxRTT:  option.Duration(time.Second),
	}
	tests := []struct {
		name  string
		alive bool
		stats healthcheck.Stats
	}{
		{
			"nil RTTStorage", true, healthcheck.Stats{
				All: 0, Fail: 0, Latest: 0, Average: 0,
			},
		},
		{
			"untested", true, healthcheck.Stats{
				All: 0, Fail: 0, Latest: 0, Average: 0,
			},
		},
		{
			"@max_rtt", true, healthcheck.Stats{
				All: 0, Fail: 0, Latest: healthcheck.Second, Average: healthcheck.Second,
			},
		},
		{
			"@max_fail", true, healthcheck.Stats{
				All: 10, Fail: 2, Latest: healthcheck.Second, Average: healthcheck.Second,
			},
		},
		{
			"@max_fail_2", true, healthcheck.Stats{
				All: 5, Fail: 1, Latest: healthcheck.Second, Average: healthcheck.Second,
			},
		},
		{
			"latest_fail", false, healthcheck.Stats{
				All: 10, Fail: 1, Latest: healthcheck.Failed, Average: healthcheck.Second,
			},
		},
		{
			"over max_fail", false, healthcheck.Stats{
				All: 5, Fail: 2, Latest: healthcheck.Second, Average: healthcheck.Second,
			},
		},
		{
			"over max_rtt", false, healthcheck.Stats{
				All: 10, Fail: 0, Latest: healthcheck.Second, Average: 2 * healthcheck.Second,
			},
		},
	}
	filter := balancer.NewAliveFilter(10, options)
	for i, tt := range tests {
		if filter.IsAlive(&tt.stats) != tt.alive {
			t.Errorf("IsAlive(#%d) = %v, want %v", i, !tt.alive, tt.alive)
		}
	}
}
