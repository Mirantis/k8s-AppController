package mocks

import "time"

//CountingResource is a fake resource that becomes ready after given timeout.
//It also increases the counter when started and decreases it when becomes ready
type CountingResource struct {
	key       string
	status    string
	counter   *CounterWithMemo
	timeout   time.Duration
	startTime time.Time
}

//Key returns a key of the CountingResource
func (c CountingResource) Key() string {
	return c.key
}

//Status returns a status of the CountingResource. It also updates the status
//after provided timeout and decrements counter
func (c *CountingResource) Status() (string, error) {
	if time.Since(c.startTime) >= c.timeout && c.status != "ready" {
		c.counter.Dec()
		c.status = "ready"
	}

	return c.status, nil
}

//Create increments counter and sets creation time
func (c *CountingResource) Create() error {
	c.counter.Inc()
	c.startTime = time.Now()
	return nil
}

//NewCountingResource creates new instance of CountingResource
func NewCountingResource(key string, counter *CounterWithMemo, timeout time.Duration) *CountingResource {
	return &CountingResource{
		key:     key,
		status:  "not ready",
		counter: counter,
		timeout: timeout,
	}
}
