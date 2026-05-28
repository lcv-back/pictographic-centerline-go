package svgwriter

import (
	"centerline-go/internal/domain"
	"fmt"
	"html"
	"strings"
)

type Result = domain.Result
type Path = domain.Path
type Point = domain.Point

func Render(result Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d">`+"\n", result.Width, result.Height)
	fmt.Fprintf(&b, `<g fill="none" stroke="black" stroke-width="%s" stroke-linecap="round" stroke-linejoin="round">`+"\n", formatFloat(result.StrokeWidth))
	for _, path := range result.Paths {
		if len(path.Points) < 2 {
			continue
		}
		fmt.Fprintf(&b, `     <path d="%s"/>`+"\n", html.EscapeString(pathData(path, result.Smooth)))
	}
	b.WriteString("</g>\n</svg>\n")
	return b.String()
}

func pathData(path Path, smooth bool) string {
	points := path.Points
	var b strings.Builder
	fmt.Fprintf(&b, "M %s %s", formatFloat(points[0].X), formatFloat(points[0].Y))

	if smooth && len(points) >= 3 && !path.Closed {
		for i := 0; i < len(points)-1; i++ {
			p0 := points[max(0, i-1)]
			p1 := points[i]
			p2 := points[i+1]
			p3 := points[min(len(points)-1, i+2)]
			c1 := Point{
				X: p1.X + (p2.X-p0.X)/6.0,
				Y: p1.Y + (p2.Y-p0.Y)/6.0,
			}
			c2 := Point{
				X: p2.X - (p3.X-p1.X)/6.0,
				Y: p2.Y - (p3.Y-p1.Y)/6.0,
			}
			fmt.Fprintf(&b, " C %s %s, %s %s, %s %s", formatFloat(c1.X), formatFloat(c1.Y), formatFloat(c2.X), formatFloat(c2.Y), formatFloat(p2.X), formatFloat(p2.Y))
		}
	} else {
		for i := 1; i < len(points); i++ {
			fmt.Fprintf(&b, " L %s %s", formatFloat(points[i].X), formatFloat(points[i].Y))
		}
	}

	if path.Closed {
		b.WriteString(" Z")
	}
	return b.String()
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", f), "0"), ".")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
