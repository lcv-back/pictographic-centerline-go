package trace

import (
	"centerline-go/internal/domain"
	"math"
)

type component struct {
	ID     int
	Pixels []int
	Center Point
}

func ExtractPaths(skeleton Mask, prune float64) []Path {
	validateMask(skeleton)
	compID, comps := nodeComponents(skeleton)
	visited := make(map[uint64]bool)

	var paths []Path
	for _, comp := range comps {
		for _, pixel := range comp.Pixels {
			for _, nb := range neighborIndices(skeleton, pixel) {
				if compID[nb] == comp.ID {
					continue
				}
				if isVisited(visited, pixel, nb) {
					continue
				}
				path := traceFromComponent(skeleton, compID, comps, comp.ID, pixel, nb, visited)
				if keepPath(path, prune) {
					paths = append(paths, path)
				}
			}
		}
	}

	for idx, on := range skeleton.Data {
		if !on {
			continue
		}
		for _, nb := range neighborIndices(skeleton, idx) {
			if isVisited(visited, idx, nb) {
				continue
			}
			path := traceLoop(skeleton, idx, nb, visited)
			if keepPath(path, prune) {
				paths = append(paths, path)
			}
		}
	}

	return paths
}

func nodeComponents(skeleton Mask) ([]int, []component) {
	n := len(skeleton.Data)
	nodeMask := make([]bool, n)
	for idx, on := range skeleton.Data {
		if !on {
			continue
		}
		deg := len(neighborIndices(skeleton, idx))
		if deg != 2 {
			nodeMask[idx] = true
		}
	}

	compID := make([]int, n)
	for i := range compID {
		compID[i] = -1
	}

	var comps []component
	queue := make([]int, 0, 32)
	for idx, isNode := range nodeMask {
		if !isNode || compID[idx] >= 0 {
			continue
		}
		id := len(comps)
		queue = append(queue[:0], idx)
		compID[idx] = id
		var pixels []int
		sumX, sumY := 0, 0

		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			pixels = append(pixels, cur)
			x, y := xy(skeleton, cur)
			sumX += x
			sumY += y

			for _, nb := range neighborIndices(skeleton, cur) {
				if nodeMask[nb] && compID[nb] < 0 {
					compID[nb] = id
					queue = append(queue, nb)
				}
			}

			count := float64(len(pixels))
			comps = append(comps, component{
				ID:     id,
				Pixels: pixels,
				Center: domain.Point{
					X: float64(sumX)/count + 0.5,
					Y: float64(sumY)/count + 0.5,
				},
			})
		}

	}

	return compID, comps
}

func traceFromComponent(skeleton Mask, compID []int, comps []component, startComp int, startPixel int, next int, visited map[uint64]bool) Path {
	points := []Point{
		comps[startComp].Center,
	}
	prev := startPixel
	cur := next
	markVisited(visited, prev, cur)

	for steps := 0; steps <= len(skeleton.Data); steps++ {
		if compID[cur] >= 0 {
			points = append(points, comps[compID[cur]].Center)
			return Path{
				Points: points,
			}
		}

		points = append(points, pointFor(skeleton, cur))
		candidates := unvisitedForwardNeighbors(skeleton, cur, prev, visited)
		if len(candidates) == 0 {
			return Path{Points: points}
		}

		nxt := candidates[0]
		markVisited(visited, cur, nxt)
		prev, cur = cur, nxt
	}

	return Path{Points: points}
}

func traceLoop(skeleton Mask, start int, next int, visited map[uint64]bool) Path {
	points := []Point{pointFor(skeleton, start)}
	prev := start
	cur := next
	markVisited(visited, prev, cur)

	for steps := 0; steps <= len(skeleton.Data); steps++ {
		if cur == start {
			return Path{Points: points, Closed: true}
		}
		points = append(points, pointFor(skeleton, cur))
		candidates := unvisitedForwardNeighbors(skeleton, cur, prev, visited)
		if len(candidates) == 0 {
			return Path{Points: points}
		}
		nxt := candidates[0]
		markVisited(visited, cur, nxt)
		prev, cur = cur, nxt
	}

	return Path{Points: points}
}

func unvisitedForwardNeighbors(skeleton Mask, cur int, prev int, visited map[uint64]bool) []int {
	var out []int
	for _, nb := range neighborIndices(skeleton, cur) {
		if nb == prev {
			continue
		}
		if !isVisited(visited, cur, nb) {
			out = append(out, nb)
		}
	}

	return out
}

func keepPath(path Path, prune float64) bool {
	if len(path.Points) < 2 {
		return false
	}
	if prune <= 0 {
		return true
	}
	return pathLength(path.Points, path.Closed) >= prune
}

func pathLength(points []Point, closed bool) float64 {
	if len(points) < 2 {
		return 0
	}
	total := 0.0
	for i := 1; i < len(points); i++ {
		total += distance(points[i-1], points[i])
	}
	if closed {
		total += distance(points[len(points)-1], points[0])
	}
	return total
}

func neighborIndices(mask Mask, idx int) []int {
	x, y := xy(mask, idx)
	out := make([]int, 0, 8)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if mask.InBounds(nx, ny) {
				nidx := mask.Index(nx, ny)
				if mask.Data[nidx] {
					out = append(out, nidx)
				}
			}
		}
	}
	return out
}

func pointFor(mask Mask, idx int) Point {
	x, y := xy(mask, idx)
	return Point{
		X: float64(x) + 0.5,
		Y: float64(y) + 0.5,
	}
}

func xy(mask Mask, idx int) (int, int) {
	return idx % mask.W, idx / mask.W
}

func edgeKey(a, b int) uint64 {
	if a > b {
		a, b = b, a
	}
	return uint64(uint32(a))<<32 | uint64(uint32(b))
}

func isVisited(visited map[uint64]bool, a, b int) bool {
	return visited[edgeKey(a, b)]
}

func markVisited(visited map[uint64]bool, a, b int) {
	visited[edgeKey(a, b)] = true
}

func distance(a, b Point) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}
