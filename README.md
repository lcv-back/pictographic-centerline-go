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

## Why Go

I choose Go because the challenge forbids third-party libraries, and Go's standard library already includes reliable PNG through `image`, `image/png`, and `image/color`. That means the program can read pixels directly without depending on Pillow, OpenCV, skimage, potrace, or any other external package.

Go is also a good fit for this submission because the core algorithm is pixel/grid processing traversal. The code stays explicit, fast enough for 1024x1024 icons, easy to run with `go run`, and easy to review. Python would be pleasant for prototying, but without Pillow or Numpy its standard library is much less convenient for PNG pixel work. Rust would be a good systems choice, but PNG decoding normally requires an external crate; implementing PNG parsing from scarch would distract from centerline algorithm.

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

## Approach Trade-offs

The prompt hints at a contours-based approach:

```text
binary mask -> trace contour -> shoot perpendiculars across the stroke -> connect midpoints
```

This implementation uses a thinning-based approach instead:

```text
binary mask -> Zhang-Suen thinning -> skeleton graph -> simplified stroked SVG path
```

Both approaches target the same thing: the medial axis of the filled shape.

The contour/perpendicular approach has higher final-quality potential for clean constant width strokes. If the local direction is known, a perpendicular cut can find the two opposite boudaries and place the centerline exactly between them. That can produce smoother and more geometrically precise center points than pixel-grid thinning.

The hard part is ambiguity. At T, H, plus and curved junctions, a perpendicular line can intersect the shape boundary more than twice. Choosing wich boundary points from a pair becomes a topology problem. A wrong pairing can connect thw wrong branches or move the junction away from where a pen stroke would naturally meet.

Zhang-Suen thinning is more direct to implement under the no-library constraint. It removes boundary pixels only when doing so preserves connectivity, so it tends to keep the branch structure of the icon intact. That make it a practical first pass for preserving topology at intersections.

The trade-off is that thinning works on the pixel grid. Diagonal and curved strokes can becomes slightly stair-stepped, and small bumps on the contour can create short skeleton spurs. This implementation addresses that with graph tracing, short-branch pruning, collinear junction merging, Ramer-Douglas-Peuker simplification, and optional cubic smoothing.

In short, I chose thinning for this version because it gives a robust, explainable, dependency-free centerline extrator. With more time, I would combine both approaches: use thinning to recover topology, then use contour/ perpendicular midpoint sampling to refine each skeleton point to a clear geometric center.

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

```

```
