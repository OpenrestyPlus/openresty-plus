package httpapi

import "testing"

func TestPublishBatchIndexesLowestWeightFirst(t *testing.T) {
	nodes := map[int64]orpDocument{
		11: {"weight": float64(10)},
		13: {"weight": float64(1)},
		14: {"weight": float64(5)},
	}
	indexes := publishBatchIndexes([]int64{11, 13, 14}, nodes, 3)
	if indexes[13] != 0 || indexes[14] != 1 || indexes[11] != 2 {
		t.Fatalf("unexpected batches: %+v", indexes)
	}
}
