package mocks

import "sync"

//CounterWithMemo is a counter with atomic increment and decrememt and
//memoization of miminum and maximum values of the counter
type CounterWithMemo struct {
	counter int
	max     int
	min     int
	sync.Mutex
}

//Inc atomically increments the value of the counter
func (c *CounterWithMemo) Inc() {
	c.Lock()
	defer c.Unlock()

	c.counter++

	if c.max < c.counter {
		c.max = c.counter
	}
}

//Dec atomically decrements the value of the counter
func (c *CounterWithMemo) Dec() {
	c.Lock()
	defer c.Unlock()

	c.counter--
	if c.min < c.counter {
		c.min = c.counter
	}
}

//Min returns minimum value that counter reached
func (c *CounterWithMemo) Min() int {
	c.Lock()
	defer c.Unlock()

	return c.min
}

//Max returns maximum value that counter reached
func (c *CounterWithMemo) Max() int {
	c.Lock()
	defer c.Unlock()

	return c.max
}

//NewCounterWithMemo creates new instance of CounterWithMemo
func NewCounterWithMemo() *CounterWithMemo {
	return &CounterWithMemo{}
}
