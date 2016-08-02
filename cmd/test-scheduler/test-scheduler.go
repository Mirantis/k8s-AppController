package main

import (
	"fmt"
	"time"

	"github.com/gluke77/k8s-ac/scheduler"
)

type FakeResource struct {
	key          string
	CreationTime time.Time
}

func (r FakeResource) Key() string {
	return r.key
}

func (r *FakeResource) Create() {
	r.CreationTime = time.Now()
	fmt.Println("Key: ", r.Key(), ", time:", r.CreationTime)
}

func (r FakeResource) Status() string {
	since := time.Since(r.CreationTime)

	if since.Seconds() < 1 {
		return "not ready"
	}

	return "ready"
}

func main() {
	resources := []scheduler.Resource{
		&FakeResource{key: "a"},
		&FakeResource{key: "b"},
		&FakeResource{key: "c"},
		&FakeResource{key: "d"},
		&FakeResource{key: "e"},
		&FakeResource{key: "f"},
	}

	deps := []scheduler.Dependency{
		{Dependent: "a", Depender: "d"},
		{Dependent: "b", Depender: "d"},
		{Dependent: "b", Depender: "e"},
		{Dependent: "c", Depender: "f"},
		{Dependent: "d", Depender: "f"},
		{Dependent: "e", Depender: "f"},
	}
	scheduler.Create(resources, deps)
}
