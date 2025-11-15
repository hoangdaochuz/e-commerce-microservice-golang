package circuitbreaker_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/require"
)

var circuitBreakerConfig *circuitbreaker.Config = &circuitbreaker.Config{
	Name:             "circuit-breaker-test",
	MaxRequests:      2,
	Interval:         30,
	Timeout:          10,
	FailureThreshold: 3,
}

func Test_CircuitBreaker(t *testing.T) {
	t.Run("Test_Handle_func_successfully_in_Closed", func(t *testing.T) {
		breaker := circuitbreaker.NewBreaker[string](circuitBreakerConfig)
		result, err := breaker.Do(context.Background(), func() (string, error) {
			return "success", nil
		})
		require.NoError(t, err)
		require.Equal(t, "success", *result)
	})

	t.Run("Test_Change_Circuit_Breaker_2_Open", func(t *testing.T) {
		breaker := circuitbreaker.NewBreaker[string](circuitBreakerConfig)
		for range 4 {
			_, err := breaker.Do(context.Background(), func() (string, error) {
				return "", fmt.Errorf("test fail")
			})
			require.NotNil(t, err)
		}

		currentState := breaker.GetCurrentState()
		require.Equal(t, gobreaker.StateOpen.String(), currentState.String())
		// after circuit breaker change state to open. The next request will be rejected immediately regardless it succeed
		_, err := breaker.Do(context.Background(), func() (string, error) {
			return "success", nil
		})
		require.Error(t, err)
	})

	t.Run("Test_Change_Circuit_Breaker_From_Open_To_HalfOpen", func(t *testing.T) {
		breaker := circuitbreaker.NewBreaker[string](circuitBreakerConfig)
		var err error
		for range 4 {
			_, err = breaker.Do(context.Background(), func() (string, error) {
				return "", fmt.Errorf("fail")
			})
		}
		require.NotNil(t, err)
		require.Equal(t, gobreaker.StateOpen.String(), breaker.GetCurrentState().String())
		// after 10s
		time.Sleep(10 * time.Second)
		require.Equal(t, gobreaker.StateHalfOpen.String(), breaker.GetCurrentState().String())

		t.Run("Test_Change_To_Open_When_A_Request_Fail_In_Half-Open", func(t *testing.T) {
			_, err := breaker.Do(context.Background(), func() (string, error) {
				return "", fmt.Errorf("Error")
			})
			require.NotNil(t, err)
			require.Equal(t, gobreaker.StateOpen.String(), breaker.GetCurrentState().String())
		})
		time.Sleep(10 * time.Second) // delay to breaker change from open -> half-open

		t.Run("Test_Change_To_Close_When_Next_Request_Success_In_Half_Open", func(t *testing.T) {
			// if next request success -> change state to close.
			// at least have {{MaxRequests}} success then circuit breaker change state from half open to close
			for range 3 {
				result, err := breaker.Do(context.Background(), func() (string, error) {
					return "success", nil
				})
				require.Equal(t, "success", *result)
				require.NoError(t, err)
			}
			require.Equal(t, gobreaker.StateClosed.String(), breaker.GetCurrentState().String())
		})
	})
}
