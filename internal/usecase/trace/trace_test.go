package trace

import (
	"testing"

	"centerline-go/internal/domain"
)

func TestHorizontalStrokeTracesNearCenter(t *testing.T) {
	mask := domain.Mask{W: 31, H: 15, Data: make([]bool, 31*15)}
	for y := 5; y <= 9; y++ {
		for x := 4; x <= 26; x++ {
			mask.Data[mask.Index(x, y)] = true
		}
	}

	result := Service{}.Trace(mask, Config{
		Threshold:   128,
		StrokeWidth: 5,
		Simplify:    0,
		Prune:       0,
	})
	if len(result.Paths) == 0 {
		t.Fatal("expected at least one centerline path")
	}

	longest := result.Paths[0]
	for _, p := range result.Paths[1:] {
		if pathLength(p.Points, p.Closed) > pathLength(longest.Points, longest.Closed) {
			longest = p
		}
	}

	totalY := 0.0
	for _, p := range longest.Points {
		totalY += p.Y
	}
	avgY := totalY / float64(len(longest.Points))
	if avgY < 6.5 || avgY > 8.5 {
		t.Fatalf("centerline average y = %.2f, want near 7.5", avgY)
	}
}

func TestMergeCollinearBranchesThroughJunction(t *testing.T) {
	paths := []Path{
		{Points: []Point{{X: 10, Y: 0}, {X: 10, Y: 10}}},
		{Points: []Point{{X: 10, Y: 10}, {X: 10, Y: 20}}},
		{Points: []Point{{X: 10, Y: 10}, {X: 20, Y: 10}}},
	}
	merged := MergeCollinear(paths)
	if len(merged) != 2 {
		t.Fatalf("merged path count = %d, want 2", len(merged))
	}

	foundVertical := false
	for _, path := range merged {
		a := path.Points[0]
		b := path.Points[len(path.Points)-1]
		if a.X == 10 && b.X == 10 && ((a.Y == 0 && b.Y == 20) || (a.Y == 20 && b.Y == 0)) {
			foundVertical = true
		}
	}
	if !foundVertical {
		t.Fatalf("expected one vertical path merged through the T junction: %#v", merged)
	}
}
