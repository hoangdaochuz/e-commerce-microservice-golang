package di

import (
	"fmt"
	"sync"

	"go.uber.org/dig"
)

var lock = &sync.Mutex{}

type DIContainer struct {
	Container *dig.Container
}

func newDIContainer() *DIContainer {
	return &DIContainer{
		Container: dig.New(),
	}
}

var DiContainer *DIContainer

func InitDIContainer() {
	if DiContainer == nil {
		lock.Lock()
		defer lock.Unlock()
		if DiContainer == nil {
			DiContainer = newDIContainer()
		} else {
			fmt.Println("DIContainer has already created")
		}
	} else {
		fmt.Println("DIContainer has already created")
	}
}

func Make(constructor interface{}) error {
	if DiContainer == nil {
		InitDIContainer()
	}
	return DiContainer.Container.Provide(constructor)
}

func Resolve(implement interface{}) error {
	if DiContainer == nil {
		InitDIContainer()
	}
	return DiContainer.Container.Invoke(implement)
}
