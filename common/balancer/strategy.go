package balancer

import "github.com/sagernet/sing-box/adapter"

// Strategy is the interface for balancer strategies
type Strategy interface {
	Pick(all, filtered []*Node, metadata *adapter.InboundContext) *Node
}
