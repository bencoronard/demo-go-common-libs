package actuator

import "go.uber.org/fx"

type ActuatorParams struct {
	fx.In
}

type Actuator interface {
	Liveness() bool
	Readiness() bool
}

type actuatorImpl struct {
}

func NewActuator(p ActuatorParams) Actuator {
	return &actuatorImpl{}
}

func (a *actuatorImpl) Liveness() bool {
	return true
}

func (a *actuatorImpl) Readiness() bool {
	return true
}
