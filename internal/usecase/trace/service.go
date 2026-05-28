package trace

import "fmt"

type Service struct{}

func (Service) Trace(fullMask Mask, cfg Config) Result {
	if cfg.Threshold == 0 {
		cfg.Threshold = 128
	}
	if cfg.Simplify < 0 {
		cfg.Simplify = 0
	}
	if cfg.Prune < 0 {
		cfg.Prune = 0
	}

	mask, offsetX, offsetT, ok := cropMask(fullMask, 2)
	if !ok {
		return Result{
			Width:       fullMask.W,
			Height:      fullMask.H,
			StrokeWidth: cfg.StrokeWidth,
			Smooth:      cfg.Smooth,
		}
	}

	skeleton := ThinZhangSuen(mask)
	paths := ExtractPaths(skeleton, cfg.Prune)
	translatePaths(paths, float64(offsetX), float64(offsetT))
	paths = MergeCollinear(paths)

	for i := range paths {
		if cfg.Simplify > 0 {
			paths[i].Points = Simplify(paths[i].Points, cfg.Simplify)
		}
	}

	strokeWidth := cfg.StrokeWidth
	if strokeWidth <= 0 {
		strokeWidth = EstimateStrokeWidth(mask, skeleton)
	}
	if strokeWidth <= 0 {
		strokeWidth = 45
	}

	return Result{
		Width:       fullMask.W,
		Height:      fullMask.H,
		StrokeWidth: strokeWidth,
		Smooth:      cfg.Smooth,
		Paths:       paths,
	}
}

func cropMask(mask Mask, margin int) (Mask, int, int, bool) {
	minX, minY := mask.W, mask.H
	maxX, maxY := -1, -1
	for y := 0; y < mask.H; y++ {
		for x := 0; x < mask.W; x++ {
			if !mask.Data[mask.Index(x, y)] {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	if maxX < minX || maxY < minY {
		return Mask{}, 0, 0, false
	}

	minX = maxInt(0, minX-margin)
	minY = maxInt(0, minY-margin)
	maxX = minInt(mask.W-1, maxX+margin)
	maxY = minInt(mask.H-1, maxY+margin)
	w := maxX - minX + 1
	h := maxY - minY + 1
	data := make([]bool, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			data[y*w+x] = mask.Data[mask.Index(minX+x, minY+y)]
		}
	}
	return Mask{
		W:    w,
		H:    h,
		Data: data,
	}, minX, minY, true
}

func translatePaths(paths []Path, dx, dy float64) {
	for i := range paths {
		for j := range paths[i].Points {
			paths[i].Points[j].X += dx
			paths[i].Points[j].Y += dy
		}
	}
}

func validateMask(m Mask) {
	if len(m.Data) != m.W*m.H {
		panic(fmt.Sprintf("mask size mismatch: expected %d, got %d", m.W*m.H, len(m.Data)))
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
