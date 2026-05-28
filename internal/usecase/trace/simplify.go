package trace

import "math"

func Simplify(points []Point, epsilon float64) []Point {
	if len(points) < 3 || epsilon <= 0 {
		return append([]Point(nil), points...)
	}
	keep := make([]bool, len(points))
	keep[0] = true
	keep[len(points)-1] = true
	simplifyRange(points, 0, len(points)-1, epsilon*epsilon, keep)

	out := make([]Point, 0, len(points))
	for i, p := range points {
		if keep[i] {
			out = append(out, p)
		}
	}

	return out
}

func simplifyRange(points []Point, first, last int, epsilonSq float64, keep []bool) {
	if last <= first+1 {
		return
	}

	bestDist := -1.0
	bestIndex := -1
	for i := first + 1; i < last; i++ {
		d := pointSegmentDistanceSq(points[i], points[first], points[last])
		if d > bestDist {
			bestDist = d
			bestIndex = i
		}
	}

	if bestDist > epsilonSq {
		keep[bestIndex] = true
		simplifyRange(points, first, bestIndex, epsilonSq, keep)
		simplifyRange(points, bestIndex, last, epsilonSq, keep)
	}
}

func pointSegmentDistanceSq(p, a, b Point) float64 {
	vx := b.X - a.X
	vy := b.Y - a.Y
	wx := p.X - a.X
	wy := p.Y - a.Y
	denom := vx*vx + vy*vy
	if denom == 0 {
		dx := p.X - a.X
		dy := p.Y - a.Y
		return dx*dx + dy*dy
	}

	t := (wx*vx + wy*vy) / denom
	t = math.Max(0, math.Min(1, t))
	proj := Point{X: a.X + t*vx, Y: a.Y + t*vy}
	dx := p.X - proj.X
	dy := p.Y - proj.Y

	return dx*dx + dy*dy
}
