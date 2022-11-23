package balancer_test

import (
	"testing"

	"github.com/sagernet/sing-box/common/balancer"
	"github.com/sagernet/sing-box/common/healthcheck"
)

func TestLeastNodes(t *testing.T) {
	nodes := []*balancer.Node{
		{Stats: healthcheck.Stats{Deviation: 50}},
		{Stats: healthcheck.Stats{Deviation: 70}},
		{Stats: healthcheck.Stats{Deviation: 100}},
		{Stats: healthcheck.Stats{Deviation: 110}},
		{Stats: healthcheck.Stats{Deviation: 120}},
		{Stats: healthcheck.Stats{Deviation: 150}},
	}
	tests := []struct {
		expected  uint
		baselines []healthcheck.RTT
		want      uint
	}{
		// typical cases
		{want: 1},
		{baselines: []healthcheck.RTT{100}, want: 2},
		{expected: 3, want: 3},
		{expected: 3, baselines: []healthcheck.RTT{50, 100, 150}, want: 5},

		// edge cases
		{expected: 0, baselines: nil, want: 1},
		{expected: 1, baselines: nil, want: 1},
		{expected: 0, baselines: []healthcheck.RTT{80, 100}, want: 2},
		{expected: 2, baselines: []healthcheck.RTT{50, 100}, want: 2},
		{expected: 9999, want: uint(len(nodes))},
		{expected: 9999, baselines: []healthcheck.RTT{50, 100, 150}, want: uint(len(nodes))},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := balancer.LeastNodes(
				nodes, tt.expected, tt.baselines,
				func(node *balancer.Node) healthcheck.RTT {
					return node.Deviation
				},
			); uint(len(got)) != tt.want {
				t.Errorf("selectNodes() = %v, want %v", len(got), tt.want)
			}
		})
	}
}
