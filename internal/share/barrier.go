package share

import "sync"

type Gate interface {
	Wait()
}

type NoopGate struct{}

func (NoopGate) Wait() {}

type ControlledGate struct {
	arrived chan struct{}
	release chan struct{}
	once    sync.Once
}

func NewControlledGate(participants int) *ControlledGate {
	if participants < 1 {
		participants = 1
	}
	return &ControlledGate{
		arrived: make(chan struct{}, participants),
		release: make(chan struct{}),
	}
}

func (g *ControlledGate) Wait() {
	g.arrived <- struct{}{}
	<-g.release
}

func (g *ControlledGate) AwaitArrivals(participants int) {
	for index := 0; index < participants; index++ {
		<-g.arrived
	}
	g.once.Do(func() { close(g.release) })
}
