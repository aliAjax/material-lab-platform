package health

import (
	"context"
	"sync"
	"time"
)

type Check func(context.Context) error

type Result struct {
	Name     string        `json:"name"`
	Healthy  bool          `json:"healthy"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
}

func Run(ctx context.Context, checks map[string]Check) []Result {
	results := make([]Result, 0, len(checks))
	channel := make(chan Result, len(checks))
	var wait sync.WaitGroup
	for name, check := range checks {
		wait.Add(1)
		go func(name string, check Check) {
			defer wait.Done()
			start := time.Now()
			err := check(ctx)
			result := Result{Name: name, Healthy: err == nil, Duration: time.Since(start)}
			if err != nil {
				result.Error = err.Error()
			}
			channel <- result
		}(name, check)
	}
	wait.Wait()
	close(channel)
	for result := range channel {
		results = append(results, result)
	}
	return results
}
