package io

import (
	"time"

	"github.com/mrmorphic/hwio"
	"github.com/pkg/errors"
)

//go:generate stringer -type=PinState

type PinState int

// Pin states
const (
	PinStateLow PinState = iota
	PinStateHigh
)

type Operator struct {
}

func NewIOOperator() *Operator {
	return &Operator{}
}

func (p *Operator) Close() {
	hwio.CloseAll()
}

type Pin struct {
	hwio.Pin
	name string
}

func (p PinState) Negate() PinState {
	if p == PinStateLow {
		return PinStateHigh
	}

	return PinStateLow
}

func (p *Pin) Toggle() error {
	state, err := p.GetState()
	if err != nil {
		return err
	}

	if err := p.SetState(state.Negate()); err != nil {
		return err
	}

	return nil
}

func (p *Pin) SetState(state PinState) error {
	if err := hwio.DigitalWrite(p.Pin, int(state)); err != nil {
		return errors.Wrapf(err, "set state for pin %s to %s", p.name, state)
	}
	return nil
}

func (p *Pin) SetHigh() error {
	if err := hwio.DigitalWrite(p.Pin, int(PinStateHigh)); err != nil {
		return errors.Wrapf(err, "set state for pin %s to %s", p.name, PinStateLow)
	}
	return nil
}

func (p *Pin) SetLow() error {
	if err := hwio.DigitalWrite(p.Pin, int(PinStateLow)); err != nil {
		return errors.Wrapf(err, "set state for pin %s to %s", p.name, PinStateLow)
	}
	return nil
}

func (p *Pin) PulseHigh(dur time.Duration) error {
	return hwio.DigitalWrite(p.Pin, hwio.LOW)
}

func (p *Pin) Close() error {
	return hwio.ClosePin(p.Pin)
}

func (p *Pin) GetState() (PinState, error) {
	state, err := hwio.DigitalRead(p.Pin)
	if err != nil {
		return PinStateLow, errors.Wrapf(err, "get state for pin %s", p.name)
	}

	return PinState(state), nil
}

func (p *Operator) NewOutputPin(name string) (*Pin, error) {
	pin, err := hwio.GetPin(name)
	if err != nil {
		return nil, errors.Wrapf(err, "create pin %s", name)
	}

	if err := hwio.PinMode(pin, hwio.OUTPUT); err != nil {
		return nil, errors.Wrapf(err, "set pin mode for pin %s to %s", name, hwio.OUTPUT)
	}

	return &Pin{Pin: pin, name: name}, nil
}

func (p *Operator) NewInputPin(name string) (*Pin, error) {
	pin, err := hwio.GetPin(name)
	if err != nil {
		return nil, errors.Wrapf(err, "create pin %s", name)
	}

	if err := hwio.PinMode(pin, hwio.INPUT); err != nil {
		return nil, errors.Wrapf(err, "set pin mode for pin %s to %s", name, hwio.INPUT)
	}

	return &Pin{Pin: pin, name: name}, nil
}
