package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// HealthChecker ...
type HealthChecker interface {
	Check(ctx context.Context) error
}

// HealthCheckFunc ...
type HealthCheckFunc func(context.Context) error

func (h HealthCheckFunc) Check(ctx context.Context) error {
	return h(ctx)
}

// HealthCheckResponse ...
type HealthCheckResponse struct {
	Status string
	Errors []string
}

var healthChecks = make(map[string]HealthCheckFunc)

func NewHealthCheck(name string, f func(context.Context) error) HealthCheckFunc {
	healthChecks[name] = HealthCheckFunc(f)
	return healthChecks[name]
}

// HealthCheck ...
func HealthCheck(checkers []HealthChecker, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var wg sync.WaitGroup
		var errorsMap sync.Map

		for serviceName, ck := range healthChecks {
			wg.Add(1)

			go func(serviceName string, checker HealthChecker) {
				ctx, cancel := context.WithTimeout(r.Context(), time.Duration(10)*time.Second)
				defer cancel()
				defer wg.Done()

				checkDone := make(chan error, 1)
				go func() {
					checkDone <- checker.Check(ctx)
				}()

				select {
				case err := <-checkDone:
					errorsMap.Store(serviceName, err)
				case <-ctx.Done():
					errorsMap.Store(serviceName, fmt.Errorf("check(%s) timeout\n", serviceName))
				}
			}(serviceName, ck)
		}

		wg.Wait()

		var response HealthCheckResponse
		healthy := true

		w.Header().Set("Content-Type", "application/json")
		errorsMap.Range(func(_, value interface{}) bool {
			if value != nil {
				healthy = false
				response.Errors = append(response.Errors, value.(error).Error())
			}
			return true
		})

		if healthy {
			response.Status = "healthy"
		} else {
			response.Status = "unhealthy"
		}

		if err := json.NewEncoder(w).Encode(&response); err != nil {
			_, _ = w.Write([]byte("error"))
		}
	})

}
