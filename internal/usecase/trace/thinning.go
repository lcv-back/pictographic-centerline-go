package trace

func ThinZhangSuen(mask Mask) Mask {
	validateMask(mask)
	out := Mask{
		W:    mask.W,
		H:    mask.H,
		Data: append([]bool(nil), mask.Data...),
	}

	if out.W < 3 || out.H < 3 {
		return out
	}

	candidates := initialThinningCandidates(out)
	changed := true
	for changed {
		changed = false
		var stepChanged bool
		candidates, stepChanged = thinStep(out, candidates, 0)
		if stepChanged {
			changed = true
		}
		candidates, stepChanged = thinStep(out, candidates, 1)
		if stepChanged {
			changed = true
		}
	}

	return out
}

func initialThinningCandidates(mask Mask) []int {
	var candidates []int
	for y := 1; y < mask.H-1; y++ {
		for x := 1; x < mask.W-1; x++ {
			idx := mask.Index(x, y)
			if mask.Data[idx] && hasBackgroundNeighbor(mask, x, y) {
				candidates = append(candidates, idx)
			}
		}
	}

	return candidates
}

func thinStep(mask Mask, candidates []int, phase int) ([]int, bool) {
	var remove []int
	removeFlag := make([]bool, len(mask.Data))

	for _, idx := range candidates {
		if !mask.Data[idx] {
			continue
		}

		x, y := xy(mask, idx)
		if x == 0 || y == 0 || x == mask.W-1 || y == mask.H-1 {
			continue
		}

		if shouldRemove(mask, x, y, phase) {
			remove = append(remove, idx)
			removeFlag[idx] = true
		}
	}

	for _, idx := range remove {
		mask.Data[idx] = false
	}

	next := make([]int, 0, len(candidates)+len(remove)*4)
	inNext := make([]bool, len(mask.Data))
	addCandidate := func(idx int) {
		if inNext[idx] || !mask.Data[idx] {
			return
		}

		x, y := xy(mask, idx)
		if x == 0 || y == 0 || x == mask.W-1 || y == mask.H-1 {
			return
		}

		if !hasBackgroundNeighbor(mask, x, y) {
			return
		}

		inNext[idx] = true
		next = append(next, idx)
	}

	for _, idx := range candidates {
		if !removeFlag[idx] {
			addCandidate(idx)
		}
	}

	for _, idx := range remove {
		x, y := xy(mask, idx)
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := x+dx, y+dy
				if mask.InBounds(nx, ny) {
					addCandidate(mask.Index(nx, ny))
				}
			}
		}
	}

	return next, len(remove) > 0
}

func shouldRemove(mask Mask, x, y int, phase int) bool {
	p2 := mask.At(x, y-1)
	p3 := mask.At(x+1, y-1)
	p4 := mask.At(x+1, y)
	p5 := mask.At(x+1, y+1)
	p6 := mask.At(x, y+1)
	p7 := mask.At(x-1, y+1)
	p8 := mask.At(x-1, y)
	p9 := mask.At(x-1, y-1)

	neighbors := countTrue(p2, p3, p4, p5, p6, p7, p8, p9)
	if neighbors < 2 || neighbors > 6 {
		return false
	}
	if transitions(p2, p3, p4, p5, p6, p7, p8, p9) != 1 {
		return false
	}

	if phase == 0 {
		return !(p2 && p4 && p6) && !(p4 && p6 && p8)
	}
	return !(p2 && p4 && p8) && !(p2 && p6 && p8)
}

func hasBackgroundNeighbor(mask Mask, x, y int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if !mask.At(x+dx, y+dy) {
				return true
			}
		}
	}
	return false
}

func countTrue(values ...bool) int {
	n := 0
	for _, v := range values {
		if v {
			n++
		}
	}
	return n
}

func transitions(values ...bool) int {
	n := 0
	for i := 0; i < len(values); i++ {
		if !values[i] && values[(i+1)%len(values)] {
			n++
		}
	}
	return n
}
