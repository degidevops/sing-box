package balancer

// Filter is the interface for nodes filter
type Filter interface {
	// Filter filters nodes from all.
	// Conventions：
	//  1. it keeps slice `all` unchanged
	//  2. it will only pick nodes from `all`, never create new elements, even a same-values-copy.
	//  3. it won't update the `qualified` field of nodes.
	Filter(all []*Node) []*Node
}
