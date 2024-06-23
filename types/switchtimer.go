package types

import (
	"encoding/gob"
	"time"

	"github.com/denkhaus/sensor/io"
	"periph.io/x/conn/v3/gpio"
)

func init() {
	gob.Register(SwitchTimer{})
}

type SwitchTimerState int

const (
	SwitchTimerStateInitialized SwitchTimerState = iota
	SwitchTimerStateOff
	SwitchTimerStateOn
)

type SwitchTimer struct {
	Name           string
	Pin            gpio.PinIO
	Description    string
	CurrentSpan    Span
	CurrentState   SwitchTimerState
	OnDuration     time.Duration
	OffDuration    time.Duration
	WarnOnPinError bool
	pin            *io.Pin
}

func (p *SwitchTimer) Write(ctx *ScriptContext) error {
	return ctx.EmbeddedStore.Upsert(p.Name, p)
}

func (p *SwitchTimer) Process(ctx *ScriptContext) error {
	ctx.Logger.Debugf("process switchtimer %s", p.Name)

	if p.pin == nil {
		pin, err := io.NewPin(p.Pin)
		if err != nil && p.WarnOnPinError {
			ctx.Logger.Warnf("create digital output pin %s: %v", p.Pin, err)
		}

		p.pin = pin
	}

	if p.CurrentState == SwitchTimerStateInitialized {
		p.CurrentState = SwitchTimerStateOff
		p.CurrentSpan = NewTimespan(time.Now(), p.OffDuration)

		if p.pin != nil {
			p.pin.SetLow()
		}

		ctx.Logger.Infof("switchtimer %s turned off", p.Name)
		return p.Write(ctx)
	}

	if p.CurrentState == SwitchTimerStateOff {
		if p.CurrentSpan.ContainsTime(time.Now()) {
			return nil
		}
		p.CurrentState = SwitchTimerStateOn
		p.CurrentSpan = NewTimespan(time.Now(), p.OnDuration)

		if p.pin != nil {
			p.pin.SetHigh()
		}

		ctx.Logger.Infof("switchtimer %s turned on", p.Name)
		return p.Write(ctx)
	}

	if p.CurrentState == SwitchTimerStateOn {
		if p.CurrentSpan.ContainsTime(time.Now()) {
			return nil
		}
		p.CurrentState = SwitchTimerStateOff
		p.CurrentSpan = NewTimespan(time.Now(), p.OffDuration)

		if p.pin != nil {
			p.pin.SetLow()
		}

		ctx.Logger.Infof("switchtimer %s turned off", p.Name)
		return p.Write(ctx)
	}

	return nil
}
