package balancer

import (
	"sort"

	"github.com/sagernet/sing-box/common/healthcheck"
)

// Strategies
const (
	StrategyRandom         string = "random"
	StrategyRoundrobin     string = "roundrobin"
	StrategyFallback       string = "fallback"
	StrategyConsistentHash string = "consistenthash"
)

// Objectives
const (
	ObjectiveAlive     string = "alive"
	ObjectiveLeastPing string = "leastping"
	ObjectiveLeastLoad string = "leastload"
)

func sortNodesByIndex(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Index < nodes[j].Index
	})
}

func sortNodesByFn(nodes []*Node, rttFunc func(*Node) healthcheck.RTT) {
	sort.Slice(nodes, func(i, j int) bool {
		left := nodes[i]
		right := nodes[j]
		leftRTT, rightRTT := rttFunc(left), rttFunc(right)
		if leftRTT != rightRTT {
			return leftRTT < rightRTT
		}
		if left.Fail != right.Fail {
			return left.Fail < right.Fail
		}
		return left.All > right.All
	})
}
