package io

import (
	"time"

	"github.com/pkg/errors"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
)

func init() {
	host.Init()
}

type Pin struct {
	gpio.PinIO
	name string
}

func (p *Pin) Toggle() error {

	state := p.GetState()

	if state == gpio.Low {
		if err := p.SetHigh(); err != nil {
			return err
		}
	} else {
		if err := p.SetLow(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Pin) SetState(state gpio.Level) error {
	if err := p.PinIO.Out(state); err != nil {
		return errors.Wrapf(err, "set state for pin %s to %s", p.name, state)
	}
	return nil
}

func (p *Pin) SetHigh() error {
	return p.SetState(gpio.High)
}

func (p *Pin) SetLow() error {
	return p.SetState(gpio.Low)
}

func (p *Pin) PulseHigh(dur time.Duration) error {
	if err := p.SetHigh(); err != nil {
		return err
	}

	time.Sleep(dur)

	if err := p.SetLow(); err != nil {
		return err
	}

	return nil
}
func (p *Pin) PulseLow(dur time.Duration) error {
	if err := p.SetLow(); err != nil {
		return err
	}

	time.Sleep(dur)

	if err := p.SetHigh(); err != nil {
		return err
	}

	return nil
}

func (p *Pin) Close() error {
	return p.PinIO.Halt()
}

func (p *Pin) GetState() gpio.Level {
	return p.PinIO.Read()
}

func NewPin(pin gpio.PinIO) *Pin {
	return &Pin{PinIO: pin}
}
