package outbound

import (
	"context"
	"net"
	"sync"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/balancer"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

var (
	_ adapter.Outbound           = (*LoadBalance)(nil)
	_ adapter.OutboundCheckGroup = (*LoadBalance)(nil)
	_ adapter.Service            = (*LoadBalance)(nil)
)

// LoadBalance is a load balance group
type LoadBalance struct {
	myOutboundGroupAdapter

	options     option.LoadBalanceOutboundOptions
	fallbackTag string

	balancer *balancer.Balancer
	fallback adapter.Outbound

	lastpicked syncString
}

type syncString struct {
	sync.RWMutex
	tag string
}

func (p *syncString) Get() string {
	p.RLock()
	defer p.RUnlock()
	return p.tag
}

func (p *syncString) Set(tag string) {
	p.Lock()
	defer p.Unlock()
	p.tag = tag
}

// NewLoadBalance creates a new load balance outbound
func NewLoadBalance(router adapter.Router, logger log.ContextLogger, tag string, options option.LoadBalanceOutboundOptions) (*LoadBalance, error) {
	return &LoadBalance{
		myOutboundGroupAdapter: myOutboundGroupAdapter{
			myOutboundAdapter: myOutboundAdapter{
				protocol: C.TypeLoadBalance,
				router:   router,
				logger:   logger,
				tag:      tag,
			},
			options: options.GroupCommonOption,
		},
		options:     options,
		fallbackTag: options.Fallback,
	}, nil
}

// Now implements adapter.OutboundGroup
func (s *LoadBalance) Now() string {
	return s.lastpicked.Get()
}

// All implements adapter.OutboundGroup
func (s *LoadBalance) All() []string {
	nodes := s.balancer.Nodes("")
	if len(nodes) == 0 {
		return []string{s.fallbackTag}
	}
	tags := make([]string, 0, len(nodes))
	s.logger.Debug(
		s.options.Pick.Objective, "/", s.options.Pick.Strategy,
		", ", len(nodes), " nodes available",
	)
	for _, n := range nodes {
		s.logger.Debug(
			"#", n.Index, " [", n.Tag(), "]",
			" STD=", n.Deviation,
			" AVG=", n.Average,
			" Fail=", n.Fail, "/", n.All,
		)
		tags = append(tags, n.Tag())
	}
	return tags
}

// Network implements adapter.Outbound
func (s *LoadBalance) Network() []string {
	fallbackNetworks := s.fallback.Network()
	fallbackTCP := common.Contains(fallbackNetworks, N.NetworkTCP)
	fallbackUDP := common.Contains(fallbackNetworks, N.NetworkUDP)
	if fallbackTCP && fallbackUDP {
		// fallback supports all network, we don't need to ask s.Balancer,
		// we know it can fallback to s.fallback for all networks even if
		// no outbound is available
		return fallbackNetworks
	}
	return s.availableNetworks(fallbackTCP, fallbackUDP)
}

// CheckAll implements adapter.OutboundCheckGroup
func (s *LoadBalance) CheckAll() {
	s.balancer.CheckAll()
}

// CheckProvider implements adapter.OutboundCheckGroup
func (s *LoadBalance) CheckProvider(tag string) {
	s.balancer.CheckProvider(tag)
}

// CheckOutbound implements adapter.OutboundCheckGroup
func (s *LoadBalance) CheckOutbound(tag string) (uint16, error) {
	return s.balancer.CheckOutbound(tag)
}

// DialContext implements adapter.Outbound
func (s *LoadBalance) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	outbound, err := s.pick(ctx, network, destination)
	if err != nil {
		return nil, err
	}
	conn, err := outbound.DialContext(ctx, network, destination)
	if err == nil {
		return conn, nil
	}
	s.logger.ErrorContext(ctx, err)
	if outbound.Tag() == s.fallbackTag {
		return nil, err
	}
	s.balancer.HealthCheck.ReportFailure(s.tag)
	nodes := s.balancer.Nodes(network)
	for _, fallback := range nodes {
		if fallback.Outbound.Tag() == outbound.Tag() {
			continue
		}
		conn, err = fallback.Outbound.DialContext(ctx, network, destination)
		if err == nil {
			return conn, nil
		}
	}
	return nil, err
}

// ListenPacket implements adapter.Outbound
func (s *LoadBalance) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	outbound, err := s.pick(ctx, N.NetworkUDP, destination)
	if err != nil {
		return nil, err
	}
	conn, err := outbound.ListenPacket(ctx, destination)
	if err == nil {
		return conn, nil
	}
	s.logger.ErrorContext(ctx, err)
	if outbound.Tag() == s.fallbackTag {
		return nil, err
	}
	s.balancer.HealthCheck.ReportFailure(s.tag)
	nodes := s.balancer.Nodes(N.NetworkUDP)
	for _, fallback := range nodes {
		if fallback.Outbound.Tag() == outbound.Tag() {
			continue
		}
		conn, err = fallback.Outbound.ListenPacket(ctx, destination)
		if err == nil {
			return conn, nil
		}
	}
	return nil, err
}

// NewConnection implements adapter.Outbound
func (s *LoadBalance) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	return NewConnection(ctx, s, conn, metadata)
}

// NewPacketConnection implements adapter.Outbound
func (s *LoadBalance) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	return NewPacketConnection(ctx, s, conn, metadata)
}

// Close implements adapter.Service
func (s *LoadBalance) Close() error {
	if s.balancer == nil {
		return nil
	}
	return s.balancer.Close()
}

// Start implements adapter.Service
func (s *LoadBalance) Start() error {
	// the fallback is required, in case that all outbounds are not available,
	// we can pick it instead of returning nil to avoid panic.
	if s.fallbackTag == "" {
		return E.New("fallback not set")
	}
	outbound, loaded := s.router.Outbound(s.fallbackTag)
	if !loaded {
		return E.New("fallback outbound not found: ", s.fallbackTag)
	}
	s.fallback = outbound
	if err := s.initProviders(); err != nil {
		return err
	}
	s.lastpicked.Set(s.fallbackTag)
	b, err := balancer.New(s.router, s.providers, s.providersByTag, s.options, s.logger)
	if err != nil {
		return err
	}
	s.balancer = b
	return s.balancer.Start()
}

func (s *LoadBalance) pick(ctx context.Context, network string, destination M.Socksaddr) (adapter.Outbound, error) {
	metadata := adapter.ContextFrom(ctx)
	if metadata == nil {
		metadata = &adapter.InboundContext{}
	}
	metadata.Destination = destination
	picked := s.balancer.Pick(network, metadata)
	var outbound adapter.Outbound
	switch {
	case picked != nil:
		outbound = picked.Outbound
	case s.fallbackTag != "":
		outbound = s.fallback
	case s.fallbackTag == "":
		outbounds := s.Outbounds()
		if len(outbounds) == 0 {
			return nil, E.New("no outbound available")
		}
		outbound = outbounds[0]
	}
	s.lastpicked.Set(outbound.Tag())
	return outbound, nil
}

// availableNetworks returns available networks of qualified nodes
func (s *LoadBalance) availableNetworks(fallbackTCP, fallbackUDP bool) []string {
	hasTCP, hasUDP := fallbackTCP, fallbackUDP
	nodes := s.balancer.Nodes("")
	for _, n := range nodes {
		if !hasTCP && common.Contains(n.Network(), N.NetworkTCP) {
			hasTCP = true
		}
		if !hasUDP && common.Contains(n.Network(), N.NetworkUDP) {
			hasUDP = true
		}
		if hasTCP && hasUDP {
			break
		}
	}
	switch {
	case hasTCP && hasUDP:
		return []string{N.NetworkTCP, N.NetworkUDP}
	case hasTCP:
		return []string{N.NetworkTCP}
	case hasUDP:
		return []string{N.NetworkUDP}
	default:
		return nil
	}
}
