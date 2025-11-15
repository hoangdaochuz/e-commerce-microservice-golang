package circuitbreaker

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker/v2"
)

type Breaker[T any] struct {
	cb     *gobreaker.CircuitBreaker[T]
	config *Config
	//metrics MetricsCollector // Implement later
}

func NewBreaker[T any](cfg *Config) *Breaker[T] {
	breaker := &Breaker[T]{
		cb: (*gobreaker.CircuitBreaker[T])(gobreaker.NewCircuitBreaker[T](gobreaker.Settings{
			Name:        cfg.Name,
			MaxRequests: uint32(cfg.MaxRequests),
			Interval:    time.Second * time.Duration(cfg.Interval),
			Timeout:     time.Second * time.Duration(cfg.Timeout),
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > uint32(cfg.FailureThreshold)
			},
			// OnStateChange: cfg.OnStateChange,
			// IsSuccessful: func(err error) bool {
			// 	return err != nil
			// },
		})),
		config: cfg,
	}
	return breaker
}

func (b *Breaker[T]) Do(ctx context.Context, handler func() (T, error)) (*T, error) {
	var zeroValue T
	result, err := b.cb.Execute(func() (T, error) {
		select {
		case <-ctx.Done():
			return zeroValue, ctx.Err()
		default:
		}
		return handler()
	})

	if err != nil {
		// handle error
		return nil, err
	}
	return &result, nil
}

func (b *Breaker[T]) DoWithCallback(handler func() (T, error), fallback func() (T, error)) (*T, error) {
	res, err := b.cb.Execute(handler)
	if err != nil {
		if fallback != nil {
			resultFallback, errFallback := fallback()
			if errFallback != nil {
				return nil, fmt.Errorf("primary err: %w, fallback err: %w", err, errFallback)
			}
			return &resultFallback, nil
		}
		return nil, err
	}
	return &res, nil
}

func (b *Breaker[T]) GetCurrentState() gobreaker.State {
	return b.cb.State()
}

func (b *Breaker[T]) GetName() string {
	return b.cb.Name()
}

func (b *Breaker[T]) GetCountSuccessRequest() int {
	return int(b.cb.Counts().TotalSuccesses)
}

func (b *Breaker[T]) GetCountFailureRequest() int {
	return int(b.cb.Counts().TotalFailures)
}

func (b *Breaker[T]) GetCount() int {
	return int(b.cb.Counts().Requests)
}

func (b *Breaker[T]) IsOpen() bool {
	return b.cb.State() == gobreaker.StateOpen
}

func (b *Breaker[T]) IsClose() bool {
	return b.cb.State() == gobreaker.StateClosed
}

func (b *Breaker[T]) IsHalfOpen() bool {
	return b.cb.State() == gobreaker.StateHalfOpen
}
