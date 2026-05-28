package svgwriter

import (
	"centerline-go/internal/domain"
	"strings"
	"testing"
)

func TestRenderUsesStrokedPaths(t *testing.T) {
	svg := Render(domain.Result{
		Width:       100,
		Height:      100,
		StrokeWidth: 5.0,
		Paths: []domain.Path{{
			Points: []domain.Point{{X: 10, Y: 10}, {X: 100, Y: 100}},
		}},
	})
	for _, want := range []string{
		`viewBox="0 0 1024 1024"`,
		`fill="none"`,
		`stroke-width="45"`,
		`stroke-linecap="round"`,
		`<path d="M 10 10 L 100 100"/>`,
	} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing %q: \n%s", want, svg)
		}
	}
}
