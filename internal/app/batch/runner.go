package batch

import (
	"centerline-go/internal/adapter/pngmask"
	"centerline-go/internal/adapter/svgwriter"
	"centerline-go/internal/domain"
	"centerline-go/internal/usecase/trace"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type Runner struct {
	Config  domain.Config
	Workers int
	Tracer  trace.Service
}

type Processed struct {
	Job       Job
	PathCount int
}

func (r Runner) Run(ctx context.Context, jobs []Job, onDone func(Processed)) error {
	if len(jobs) == 0 {
		return nil
	}

	workers := r.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > len(jobs) {
		workers = len(jobs)
	}
	if r.Tracer == (trace.Service{}) {
		r.Tracer = trace.Service{}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobCh := make(chan Job)
	doneCh := make(chan Processed)
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for job := range jobCh {
				processed, err := r.process(job)
				if err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
					return
				}

				select {
				case doneCh <- processed:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobCh)
		for _, job := range jobs {
			select {
			case jobCh <- job:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for processed := range doneCh {
		if onDone != nil {
			onDone(processed)
		}
	}

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (r Runner) process(job Job) (Processed, error) {
	mask, err := pngmask.DecodeFile(job.Input, r.Config.Threshold)
	if err != nil {
		return Processed{}, fmt.Errorf("error decoding %s: %w", job.Input, err)
	}

	result := r.Tracer.Trace(mask, r.Config)
	if err := writeAtomic(job.Output, []byte(svgwriter.Render(result))); err != nil {
		return Processed{}, fmt.Errorf("error writing %s: %w", job.Output, err)
	}

	return Processed{
		Job:       job,
		PathCount: len(result.Paths),
	}, nil
}

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}

	cleanup = false
	return nil
}
