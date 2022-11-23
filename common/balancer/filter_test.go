package balancer_test

import (
	"testing"

	"github.com/sagernet/sing-box/common/balancer"
)

func BenchmarkAlive32(b *testing.B) {
	benchmarkFilter(b, benchmarkAliveFilter, 32)
}

func BenchmarkAlive128(b *testing.B) {
	benchmarkFilter(b, benchmarkAliveFilter, 128)
}

func BenchmarkLeastLoad32(b *testing.B) {
	benchmarkFilter(b, benchmarkLeastLoadFilter, 32)
}

func BenchmarkLeastLoad128(b *testing.B) {
	benchmarkFilter(b, benchmarkLeastLoadFilter, 128)
}

func benchmarkFilter(b *testing.B, f balancer.Filter, count int) {
	outbounds, store := genStorage(count)
	alive := allNodes(outbounds, store)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Filter(alive)
	}
}
