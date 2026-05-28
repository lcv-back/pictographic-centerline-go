package trace

import (
	"fmt"
	"math"
)

type endpoint struct {
	PathIndex int
	AtStart   bool
}

func MergeCollinear(paths []Path) []Path {
	paths = append([]Path(nil), paths...)
	for {
		i, j, endI, endJ, ok := findMergePair(paths)
		if !ok {
			return paths
		}
		merged := mergePair(paths[i], paths[j], endI, endJ)
		if j < i {
			i, j = j, i
		}
		paths[i] = merged
		paths = append(paths[:j], paths[j+1:]...)
	}
}

func findMergePair(paths []Path) (int, int, endpoint, endpoint, bool) {
	byPoint := make(map[string][]endpoint)
	for i, path := range paths {
		if path.Closed || len(path.Points) < 2 {
			continue
		}
		startKey := pointKey(path.Points[0])
		endKey := pointKey(path.Points[len(path.Points)-1])
		byPoint[startKey] = append(byPoint[startKey], endpoint{PathIndex: i, AtStart: true})
		byPoint[endKey] = append(byPoint[endKey], endpoint{PathIndex: i, AtStart: false})
	}

	bestDot := -0.92
	var bestA, bestB endpoint
	found := false
	for _, endpoints := range byPoint {
		for a := 0; a < len(endpoints); a++ {
			for b := a + 1; b < len(endpoints); b++ {
				if endpoints[a].PathIndex == endpoints[b].PathIndex {
					continue
				}
				dirA, okA := endpointDirection(paths[endpoints[a].PathIndex], endpoints[a].AtStart)
				dirB, okB := endpointDirection(paths[endpoints[b].PathIndex], endpoints[b].AtStart)
				if !okA || !okB {
					continue
				}

				dot := dirA.X*dirB.X + dirA.Y*dirB.Y
				if dot < bestDot {
					bestDot = dot
					bestA = endpoints[a]
					bestB = endpoints[b]
					found = true
				}
			}
		}
	}

	if !found {
		return 0, 0, endpoint{}, endpoint{}, false
	}

	return bestA.PathIndex, bestB.PathIndex, bestA, bestB, true
}

func mergePair(a, b Path, endA, endB endpoint) Path {
	pa := append([]Point(nil), a.Points...)
	pb := append([]Point(nil), b.Points...)

	if endA.AtStart {
		reversePoints(pa)
	}

	if !endB.AtStart {
		reversePoints(pb)
	}

	merged := make([]Point, 0, len(pa)+len(pb)-1)
	merged = append(merged, pa...)
	merged = append(merged, pb[1:]...)
	return Path{Points: merged}
}

func endpointDirection(path Path, atStart bool) (Point, bool) {
	if len(path.Points) < 2 {
		return Point{}, false
	}

	var from, to Point
	if atStart {
		from = path.Points[0]
		to = path.Points[1]
	} else {
		from = path.Points[len(path.Points)-1]
		to = path.Points[len(path.Points)-2]
	}

	dx := to.X - from.X
	dy := to.Y - from.Y

	length := math.Hypot(dx, dy)
	if length == 0 {
		return Point{}, false
	}

	return Point{X: dx / length, Y: dy / length}, true
}

func reversePoints(points []Point) {
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}
}

func pointKey(p Point) string {
	return fmt.Sprintf("%.2f, %.2f", p.X, p.Y)
}
