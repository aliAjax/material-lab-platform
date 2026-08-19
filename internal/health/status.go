package health

import (
	"context"
	"fmt"
	"sort"
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
	if checks == nil {
		return []Result{}
	}
	results := make([]Result, 0, len(checks))
	channel := make(chan Result, len(checks))
	var wait sync.WaitGroup
	for name, check := range checks {
		wait.Add(1)
		go func(name string, check Check) {
			defer wait.Done()
			start := time.Now()
			err := runCheck(ctx, check)
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
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	return results
}

func runCheck(ctx context.Context, check Check) (err error) {
	if check == nil {
		return fmt.Errorf("health check is not configured")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("health check panicked: %v", recovered)
		}
	}()
	return check(ctx)
}
