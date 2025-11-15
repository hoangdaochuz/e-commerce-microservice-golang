package circuitbreaker

import (
	"fmt"
	"reflect"
	"sync"
)

type CircuitBreakerRegistry[T any] struct {
	mu       *sync.RWMutex
	breakers map[string]*Breaker[T]
	configs  map[string]*Config
}

var (
	globalRegistries sync.Map
)

func newCircuitBreakerRegistry[T any]() *CircuitBreakerRegistry[T] {
	return &CircuitBreakerRegistry[T]{
		mu:       &sync.RWMutex{},
		breakers: make(map[string]*Breaker[T]),
		configs:  make(map[string]*Config),
	}
}

// Singleton pattern
func GetRegistry[T any]() *CircuitBreakerRegistry[T] {
	var targetType T

	targetTypeString := reflect.TypeOf(targetType).String()
	registry, _ := globalRegistries.LoadOrStore(targetTypeString, newCircuitBreakerRegistry[T]())
	return registry.(*CircuitBreakerRegistry[T])
}

func (c *CircuitBreakerRegistry[T]) register(name string, config *Config) error {
	if _, ok := c.breakers[name]; ok {
		return fmt.Errorf("breaker %s already exist", name)
	}
	breaker := NewBreaker[T](config)
	c.breakers[name] = breaker
	c.configs[name] = config
	return nil
}

func (c *CircuitBreakerRegistry[T]) Register(name string, config *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.register(name, config)
}

func (c *CircuitBreakerRegistry[T]) GetBreakerByName(name string) *Breaker[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	breaker, ok := c.breakers[name]
	if !ok {
		return nil
	}
	return breaker
}

func (c *CircuitBreakerRegistry[T]) GetOrCreateBreaker(name string, config *Config) (*Breaker[T], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var breaker *Breaker[T]
	var ok bool
	if breaker, ok = c.breakers[name]; !ok {
		err := c.register(name, config)
		if err != nil {
			return nil, err
		}
		return c.breakers[name], nil
	}
	return breaker, nil
}

func (c *CircuitBreakerRegistry[T]) Remove(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// clear breaker
	if _, ok := c.breakers[name]; !ok {
		return nil
	}
	delete(c.breakers, name)
	// clear config
	if _, ok := c.configs[name]; !ok {
		return nil
	}
	delete(c.configs, name)
	return nil
}

func (c *CircuitBreakerRegistry[T]) ListCircuitBreakerName() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := []string{}
	for key := range c.breakers {
		result = append(result, key)
	}
	return result
}

func (c *CircuitBreakerRegistry[T]) CountBreaker() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.breakers)
}

func (c *CircuitBreakerRegistry[T]) GetAllBreakers() []*Breaker[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := []*Breaker[T]{}
	for _, val := range c.breakers {
		result = append(result, val)
	}
	return result
}
