package domain

type Config struct {
	Threshold   uint8
	StrokeWidth float64
	Simplify    float64
	Prune       float64
	Smooth      bool
}

type Result struct {
	Width       int
	Height      int
	StrokeWidth float64
	Smooth      bool
	Paths       []Path
}

type Path struct {
	Points []Point
	Closed bool
}

type Point struct {
	X float64
	Y float64
}

type Mask struct {
	W    int
	H    int
	Data []bool
}

func (m Mask) Index(x, y int) int {
	return y*m.W + x
}

func (m Mask) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < m.W && y < m.H
}

func (m Mask) At(x, y int) bool {
	if !m.InBounds(x, y) {
		return false
	}
	return m.Data[m.Index(x, y)]
}
