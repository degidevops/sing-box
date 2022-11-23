package balancer

import (
	"github.com/sagernet/sing-box/common/healthcheck"
	"github.com/sagernet/sing-box/option"
)

var _ Filter = (*AliveFilter)(nil)

// AliveFilter is the alive nodes filter
type AliveFilter struct {
	maxRTT      healthcheck.RTT
	maxFailRate float32
}

// NewAliveFilter returns a new AliveFilter
func NewAliveFilter(sampling uint, options option.LoadBalancePickOptions) *AliveFilter {
	return &AliveFilter{
		maxRTT:      healthcheck.RTTOf(options.MaxRTT),
		maxFailRate: float32(options.MaxFail) / float32(sampling),
	}
}

// Filter implements NodesFilter.
// NOTICE: be aware of the coding convention of this function
func (o *AliveFilter) Filter(all []*Node) []*Node {
	alive := make([]*Node, 0, len(all))
	for _, node := range all {
		if o.IsAlive(&node.Stats) {
			alive = append(alive, node)
		}
	}
	return alive
}

// IsAlive tells if a node is alive according to the s
func (o *AliveFilter) IsAlive(s *healthcheck.Stats) bool {
	if s.All == 0 {
		// untetsted
		return true
	}
	if s.Latest == healthcheck.Failed {
		return false
	}
	if s.Fail > 0 && float32(s.Fail)/float32(s.All) > o.maxFailRate {
		return false
	}
	if o.maxRTT > 0 && s.Average > o.maxRTT {
		return false
	}
	return true
}
