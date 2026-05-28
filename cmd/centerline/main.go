package main

import (
	"centerline-go/internal/app/batch"
	"centerline-go/internal/domain"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
)

func main() {
	input := flag.String("input", "", "PNG file or directory to process")
	output := flag.String("output", "", "SVG file or output directory; defaults next to the input")
	threshold := flag.Int("threshold", 128, "foreground luminance threshold, 0..255")
	strokeWidth := flag.Float64("stroke-width", 45, "SVG stroke width; use <=0 to estimate from the mask")
	simplify := flag.Float64("simplify", 2.0, "Ramer-Douglas-Peucker simplification tolerance in pixels")
	prune := flag.Float64("prune", 8.0, "drop skeleton branches shorter than this many pixels")
	smooth := flag.Bool("smooth", true, "emit cubic Beziers through simplified points")
	workers := flag.Int("workers", runtime.NumCPU(), "number of PNG files to process concurrently")
	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "missing -input")
		flag.Usage()
		os.Exit(2)
	}

	if *threshold < 0 || *threshold > 255 {
		fmt.Fprintln(os.Stderr, "-threshold must be between 0 and 255")
		os.Exit(2)
	}

	jobs, err := batch.PlanJobs(*input, *output)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error planning jobs:", err)
		os.Exit(1)
	}

	cfg := domain.Config{
		Threshold:   uint8(*threshold),
		StrokeWidth: *strokeWidth,
		Simplify:    *simplify,
		Prune:       *prune,
		Smooth:      *smooth,
	}

	runner := batch.Runner{
		Config:  cfg,
		Workers: *workers,
	}

	err = runner.Run(context.Background(), jobs, func(processed batch.Processed) {
		fmt.Printf("%s -> %s (%d paths)\n", processed.Job.Input, processed.Job.Output, processed.PathCount)
	})

	if err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}
