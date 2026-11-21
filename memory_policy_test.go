package main

import "testing"

func TestSafeServerMemoryLimit(t *testing.T) {
	tests := []struct{ total, free, want int }{
		{8192, 5734, 3686},
		{4096, 2500, 1476},
		{8192, 300, 0},
		{0, 4000, 0},
		{8192, 8192, 4915},
	}
	for _, tt := range tests {
		got := safeServerMemoryLimit(tt.total, tt.free)
		if got != tt.want {
			t.Errorf("total=%d livre=%d: obteve %d, esperava %d", tt.total, tt.free, got, tt.want)
		}
	}
}
func TestRecommendedMemoryWithinBudget(t *testing.T) {
	for _, tc := range []struct{ total, free int }{{8192, 5734}, {4096, 2500}, {8192, 400}, {32768, 30000}} {
		limit := safeServerMemoryLimit(tc.total, tc.free)
		recommended := recommendedServerMemory(tc.total, tc.free)
		if recommended > limit {
			t.Fatalf("recomendação %d excede limite %d", recommended, limit)
		}
	}
}

func TestServerMemoryTiersAreHalfGiBSteps(t *testing.T) {
	got := serverMemoryTiers(1704)
	want := []int{512, 1024, 1536}
	if len(got) != len(want) {
		t.Fatalf("tiers=%v, queria %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tiers=%v, queria %v", got, want)
		}
	}
}

func TestManualMemoryLimitAllowsThreeGiBOnIdleEightGiBPC(t *testing.T) {
	if got := manualServerMemoryLimit(8015, 3686); got < 3072 {
		t.Fatalf("limite manual %d não permite 3 GiB", got)
	}
}
func TestJavaHeapLeavesNativeOverhead(t *testing.T) {
	for _, limit := range []int{512, 1024, 2048, 4096} {
		heap := javaHeapLimit(limit)
		if heap >= limit || heap < 256 {
			t.Fatalf("limite de heap ruim: processo %d, heap %d", limit, heap)
		}
	}
}
