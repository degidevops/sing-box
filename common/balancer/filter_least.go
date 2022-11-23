package balancer

import (
	"github.com/sagernet/sing-box/common/healthcheck"
	"github.com/sagernet/sing-box/option"
)

var _ Filter = (*LeastFilter)(nil)

// LeastFilter is the least load nodes filter
type LeastFilter struct {
	*AliveFilter
	expected  uint
	baselines []healthcheck.RTT
	rttFunc   func(node *Node) healthcheck.RTT
}

// NewLeastFilter returns a new LeastFilter
func NewLeastFilter(sampling uint, options option.LoadBalancePickOptions, rttFunc func(node *Node) healthcheck.RTT) *LeastFilter {
	return &LeastFilter{
		AliveFilter: NewAliveFilter(sampling, options),
		expected:    options.Expected,
		baselines:   healthcheck.RTTsOf(options.Baselines),
		rttFunc:     rttFunc,
	}
}

// Filter implements NodesFilter.
// NOTICE: be aware of the coding convention of this function
func (o *LeastFilter) Filter(all []*Node) []*Node {
	nodes := LeastNodes(
		o.AliveFilter.Filter(all),
		o.expected, o.baselines,
		o.rttFunc,
	)
	return nodes
}

// LeastNodes filters ordered nodes according to Baselines and Expected Count.
//
// The strategy always improves network response speed, not matter which mode below is configurated.
// But they can still have different priorities.
//
// 1. Bandwidth priority: no Baseline + Expected Count > 0.: selects `Expected Count` of nodes.
// (one if Expected Count <= 0)
//
// 2. Bandwidth priority advanced: Baselines + Expected Count > 0.
// Select `Expected Count` amount of nodes, and also those near them according to baselines.
// In other words, it selects according to different Baselines, until one of them matches
// the Expected Count, if no Baseline matches, Expected Count applied.
//
// 3. Speed priority: Baselines + `Expected Count <= 0`.
// go through all baselines until find selects, if not, select none. Used in combination
// with 'balancer.fallbackTag', it means: selects qualified nodes or use the fallback.
func LeastNodes(
	nodes []*Node, expected uint, baselines []healthcheck.RTT,
	rttFunc func(node *Node) healthcheck.RTT,
) []*Node {
	if len(nodes) == 0 {
		// s.logger.Debug("no qualified nodes")
		return nil
	}
	sortNodesByFn(nodes, rttFunc)
	expected2 := int(expected)
	availableCount := len(nodes)
	if expected2 > availableCount {
		return nodes
	}

	if expected2 <= 0 {
		expected2 = 1
	}
	if len(baselines) == 0 {
		return nodes[:expected2]
	}

	count := 0
	// go through all base line until find expected selects
	for _, baseline := range baselines {
		for i := count; i < availableCount; i++ {
			if rttFunc(nodes[i]) >= baseline {
				break
			}
			count = i + 1
		}
		// don't continue if find expected selects
		if count >= expected2 {
			break
		}
	}
	if expected > 0 && count < expected2 {
		count = expected2
	}
	return nodes[:count]
}
