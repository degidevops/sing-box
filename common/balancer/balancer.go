package balancer

import (
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/healthcheck"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
)

// Balancer is the load balancer
type Balancer struct {
	*healthcheck.HealthCheck

	router    adapter.Router
	providers []adapter.Provider
	logger    log.ContextLogger
	options   *option.LoadBalanceOutboundOptions

	filter   Filter
	strategy Strategy
}

// Node is a banalcer Node with health check result
type Node struct {
	adapter.Outbound
	healthcheck.Stats

	Index     int
	qualified bool
}

// New creates a new load balancer
//
// The globalHistory is optional and is only used to sync latency history
// between different health checkers. Each HealthCheck will maintain its own
// history storage since different ones can have different check destinations,
// sampling numbers, etc.
func New(
	router adapter.Router,
	providers []adapter.Provider, providersByTag map[string]adapter.Provider,
	options option.LoadBalanceOutboundOptions, logger log.ContextLogger,
) (*Balancer, error) {
	if options.Pick.Strategy == "" {
		options.Pick.Strategy = StrategyRandom
	}
	if options.Pick.Objective == "" {
		options.Pick.Objective = ObjectiveAlive
	}

	var (
		filter   Filter
		strategy Strategy
	)
	switch options.Pick.Objective {
	case ObjectiveAlive:
		filter = NewAliveFilter(options.Check.Sampling, options.Pick)
	case ObjectiveLeastLoad:
		filter = NewLeastFilter(
			options.Check.Sampling, options.Pick,
			func(node *Node) healthcheck.RTT {
				return node.Deviation
			},
		)
	case ObjectiveLeastPing:
		filter = NewLeastFilter(
			options.Check.Sampling, options.Pick,
			func(node *Node) healthcheck.RTT {
				return node.Average
			},
		)
	default:
		return nil, E.New("unknown objective: ", options.Pick.Objective)
	}
	switch options.Pick.Strategy {
	case StrategyRandom:
		strategy = NewRandomStrategy()
	case StrategyRoundrobin:
		strategy = NewRoundRobinStrategy()
	case StrategyFallback:
		strategy = NewFallbackStrategy()
	case StrategyConsistentHash:
		if options.Pick.Objective != ObjectiveAlive {
			// nodes selected by objectives other than "alive", are few and always
			// changing, though the rest codes can run without error, but it's
			// meaningless, because most of requests will failed to pick a qulified
			// node, and be routed to 'fallback' node
			return nil, E.New("consistenthash strategy works only with 'alive' objective")
		}
		strategy = NewConsistentHashStrategy()
	default:
		return nil, E.New("unknown strategy: ", options.Pick.Strategy)
	}
	return &Balancer{
		router:      router,
		options:     &options,
		logger:      logger,
		providers:   providers,
		HealthCheck: healthcheck.New(router, providers, providersByTag, options.Check, logger),
		filter:      filter,
		strategy:    strategy,
	}, nil
}

// Pick picks a node
func (b *Balancer) Pick(network string, metadata *adapter.InboundContext) *Node {
	all := b.nodes(network)
	filtered := b.filter.Filter(all)
	return b.strategy.Pick(all, filtered, metadata)
}

// Nodes returns qualified nodes according to pick options
func (b *Balancer) Nodes(network string) []*Node {
	return b.filter.Filter(b.nodes(network))
}

// nodes returns all nodes for the network
func (b *Balancer) nodes(network string) []*Node {
	all := make([]*Node, 0)
	idx := 0
	for _, provider := range b.providers {
		for _, outbound := range provider.Outbounds() {
			idx++
			node := &Node{
				Outbound: outbound,
				Index:    idx,
			}
			networks := node.Network()
			if network != "" && !common.Contains(networks, network) {
				continue
			}
			if group, ok := outbound.(adapter.OutboundGroup); ok {
				real, err := adapter.RealOutbound(b.router, group)
				if err != nil {
					continue
				}
				outbound = real
			}
			node.Stats = b.HealthCheck.Storage.Stats(outbound.Tag())
			all = append(all, node)
		}
	}
	return all
}
