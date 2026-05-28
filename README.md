# Centerline PNG to SVG

Standalone Go 1.24+ solution for the centerline extraction challenge. It uses only the Go standard library.

## Usage

Process one PNG:

```bash
go run ./cmd/centerline \
    -input ./input/<name>.png \
    -output ./out/<name>.svg
```

```bash
go run ./cmd/centerline \
    -input ./input/png1.png \
    -output ./out/png1.svg
```

Process a folder of PNG files:

```bash
go run ./cmd/centerline \
    -input ./input \
    -output ./out
```

Useful flags:

```text
-threshold 128 black/white cutoff for the binary mask
-stroke-width 45 SVG stroke width; use 0 to estimate from the image
-simplify 2 path simplification tolerance in pixels
-prune 8 remove tiny skeleton branches shorter than this
-smooth true emit cubic Beziers through simplified points
-workers N number of PNG files processed concurrently
```

For the challenge images, the output SVG keeps the input dimensions in the viewBox. With the provided 1024x1024 icons that gives `viewBox="0 0 1024 1024"`.

## Architecture

The code is organized as a small Clean Architecture CLI:

```text
cmd/centerline
    CLI flags and process exit behavior

internal/app/batch
    Job planning, worker pool, atomic output writes

internal/adapter/pngmask
    PNG decoding and black/white mask creation

internal/adapter/svgwriter
    SVG serialization

internal/usecase/trace
    Centerline extraction algorithms

internal/domain
    Plain data types: Mask, Point, Path, Result, Config

```

Dependency direction is inward. The centerline algorithms does not read files, write SVG, or know about CLI flags. It receives a binary `domain.Mask` and returns a `domain.Result`. This keeps the core algorithms testable and makes it easy to swap PNG/SVG adapters laters.

## Approach

The program turns each PNG into a binary foreground mask using alpha plus luminance: transparent pixels and light pixels are background, dark opaque pixels are foreground.

Then it applies Zhang-Suen thinning. This repeatedly removes boundary pixels that are safe to delete while preserving connectivity. The important checks are:

- the pixel must have 2 to 6 foreground neighbors, so endpoints are preserved.
- the 8-neighbor ring must have exactly one background-to-foreground transition, so deleting the pixels does not split a component;
- each half-step protects a different pair of directions, so the shape dhrink evenly toward the middle instead of drifting to one side.

After thinning, the remaining one-pixel-wide skeleton is converted into a graph. Pixels with degree other than 2 become graph nodes; chains of degree-2 becoms SVG paths. Small branches are pruned, paths are simplified with Ramer-Douglas-Peucker, and the SVG is emitted as stroked paths will `fill="none"`, round caps, and round joins.

Stroke width defaults to 45 to match the reference style. Passing `-stroke-with 0` estimates a width from a simple chamfer distance transform: skeleton pixels are roughly at the stroke radius, so twice their median distance to the background is a good first estimate.

## Differences from the Hint

The hint suggest tracing contours and probing perpendiculars across the stroke to find midpoints. This implementation instead uses topology-preserving thinning. Both approaches aim at the medial exis; thinning is simpler to implement without third-party libraries and gives good topology for solid icon strokes.

## Know Divergences

- Junction merging is geometric and local. It handles clean H/T/+ intersections, but ambiguous curved joins can still be emitted as several paths meeting at the same point.
- Thick blobs or decorative filled regions can create extra medial-axis branches. The `-prune` flag removes small artifacts, but shape-specific pruning would be better.
- Zhang-Suen thinning works on the pixel grid, so diagonal and curved strokes can have small bias before smoothing.
- The default cubic smoothing is conservative but can round sharp intended bends slightly. Use `-smooth=false` is exact polyline topology matters more than visual smoothness.

## Scaling Notes

Folder processing uses a bounded worker pool. Each worker processes one PNG at a time through decode, trace, render, and atomic write. This scales across CPU cores without loading all image pixels nto memory at once.
For larger datasets, tune `-workers` to match the machine. More workers can improve throughput, but skeletonization is CPU-heavy and each active 1024x1024 image owns several masks, so an excessive worker count wastes memory and cache. Output writes are atomic via temporary file plus renames, so interrupted runs do not leave half-written SVGs at the target path.

## What I would improve

- Add contour pairing for cleaner midpoint samples on constant-width strokes.
- Improve junction classification with local stroke radisus and contour evidence instead of only endpoint angles.,
- Add a better spur classifier using local stroke radius, not just path length.
- COmpare output against the provided reference SVGs with a small rastrization-based metrics.
