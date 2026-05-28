# Architecture

## Goal

The challenge is algorimic, but the code should still easy to review. The project is split so the medial-axis logic is independent from file formats and batch execution.

## Layers

````text
CLI
    cmd/centerline
    Parses flags, validates user input, prints progress

Application
    internal/app/batch
    Plans file jobs, runs a bounded worker pool, writes outputs atomically.

Adapters
    internal/adapter/pngmask
    Converts PNG pixels into a binary foreground mask.

    internal/adapter/svgwriter
    Serializes traced paths into SVG.

Use case
    internal/usecase/trace
    Converts a binary mask into centerline paths.

Domain
    internal/domain
    Shared data types with no I/O and no algorithms policy.

The dependency rule is: outer layers may import inner layers, but inner layers do not import outer layers. For example, `trace` does not know PNG exists; it only receives a `domain.Mask`

## Data Flow

```text
PNG file
    -> pngmask.DecodeFile
    -> domain.Mask
    -> trace.Service.Trace
    -> domain.Result
    -> svgwriter.Render
    -> atomic SVG write
````

## Scaling Behavior

Batch mode creates a list of PNG jobs, then processes them through a fixed-size worker pool. The default worker count is `runtime.NumCPU()`, and it can be overridden with `-workers`.

The design keeps memory bounded by job concurrency: each worker owns only the masks and paths for its current PNG. For a large folder, the program does not decode every image before processing starts.

Writes go to a temporary file in the destination directory and then rename into place. That makes output robust when a run is interrupted or a later file fails.

## Algorithms Boundary

The `trace.Service` is the core use case:

```go
Trace(mask domain.Mask, cfg domain,Config) domain.Result
```

That function is the place to improve for cleaner constant-width strokes;

- contour-pair midpoint extraction for cleaner constant-width strokes;
- radius-aware spur pruning;
- better junction classification;
- tile-based or active-set thinning for very large images.
