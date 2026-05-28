package batch

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Job struct {
	Input  string
	Output string
}

func PlanJobs(input, output string) ([]Job, error) {
	info, err := os.Stat(input)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		out := output
		if out == "" {
			out = replaceExt(input, ".svg")
		} else if outInfo, err := os.Stat(out); err == nil && outInfo.IsDir() {
			out = filepath.Join(out, replaceExt(filepath.Base(input), ".svg"))
		}
		return []Job{{Input: input, Output: out}}, nil
	}

	entries, err := os.ReadDir(input)
	if err != nil {
		return nil, err
	}

	outDir := output
	if outDir == "" {
		outDir = input
	} else if outInfo, err := os.Stat(outDir); err == nil && !outInfo.IsDir() {
		return nil, fmt.Errorf("when -input is a directory, -output must be a directory")
	}

	var jobs []Job
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.EqualFold(filepath.Ext(name), ".png") {
			jobs = append(jobs, Job{
				Input:  filepath.Join(input, name),
				Output: filepath.Join(outDir, replaceExt(name, ".svg")),
			})
		}
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].Input < jobs[j].Input })
	if len(jobs) == 0 {
		return nil, fmt.Errorf("no PNG files found in %s", input)
	}
	return jobs, nil
}

func replaceExt(path, ext string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + ext
}
