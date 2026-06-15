package repository

import (
	"fmt"
	"testing"
)

func TestHashIndex_Deterministic(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	a := hashIndex(id)
	b := hashIndex(id)
	if a != b {
		t.Fatalf("hashIndex not deterministic: %d vs %d", a, b)
	}
}

func TestHashIndex_NonNegative(t *testing.T) {
	ids := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"00000000-0000-0000-0000-000000000001",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
		"",
		"ana-silva",
	}
	for _, id := range ids {
		h := hashIndex(id)
		if h < 0 {
			t.Fatalf("hashIndex(%q) = %d, want non-negative", id, h)
		}
	}
}

func TestHashIndex_VariantDistribution(t *testing.T) {
	variants := []string{"chronological", "ranked"}
	counts := map[string]int{"chronological": 0, "ranked": 0}
	for i := 0; i < 200; i++ {
		id := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", i, i*3, i*7, i*11, i*13)
		idx := hashIndex(id) % len(variants)
		counts[variants[idx]]++
	}
	if counts["chronological"] == 0 || counts["ranked"] == 0 {
		t.Fatalf("expected both variants to be assigned, got %v", counts)
	}
}
