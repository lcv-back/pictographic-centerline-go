package trace

import "sort"

const infDistance = 1 << 28

func EstimateStrokeWidth(mask Mask, skeleton Mask) float64 {
	validateMask(mask)
	validateMask(skeleton)
	if mask.W != skeleton.W || mask.H != skeleton.H {
		return 0
	}

	dist := make([]int, len(mask.Data))
	for i, on := range mask.Data {
		if on {
			dist[i] = infDistance
		}
	}

	for y := 0; y < mask.H; y++ {
		for x := 0; x < mask.W; x++ {
			idx := mask.Index(x, y)
			if dist[idx] == 0 {
				continue
			}
			relax(mask, dist, x, y, x-1, y, 10)
			relax(mask, dist, x, y, x, y-1, 10)
			relax(mask, dist, x, y, x-1, y-1, 14)
			relax(mask, dist, x, y, x+1, y-1, 14)
		}
	}

	for y := mask.H - 1; y >= 0; y-- {
		for x := mask.W - 1; x >= 0; x-- {
			idx := mask.Index(x, y)
			if dist[idx] == 0 {
				continue
			}
			relax(mask, dist, x, y, x+1, y, 10)
			relax(mask, dist, x, y, x, y+1, 10)
			relax(mask, dist, x, y, x+1, y+1, 14)
			relax(mask, dist, x, y, x-1, y+1, 14)
		}
	}

	var radii []int
	for idx, on := range skeleton.Data {
		if on && dist[idx] > 0 && dist[idx] < infDistance {
			radii = append(radii, dist[idx])
		}
	}
	if len(radii) == 0 {
		return 0
	}
	sort.Ints(radii)
	medianRadius := float64(radii[len(radii)/2]) / 10.0
	return 2 * medianRadius
}

func relax(mask Mask, dist []int, x, y, nx, ny, cost int) {
	if !mask.InBounds(nx, ny) {
		return
	}
	idx := mask.Index(x, y)
	nidx := mask.Index(nx, ny)
	if dist[nidx]+cost < dist[idx] {
		dist[idx] = dist[nidx] + cost
	}
}
