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

type DurationCallback func(onDuration time.Duration, offDuration time.Duration) (time.Duration, time.Duration)

type SwitchTimerState int

const (
	SwitchTimerStateInitialized SwitchTimerState = iota
	SwitchTimerStateOff
	SwitchTimerStateOn
)

type SwitchTimer struct {
	Name         string
	Description  string
	CurrentSpan  Span
	CurrentState SwitchTimerState
	OnDuration   time.Duration
	OffDuration  time.Duration
	Inverted     bool
	pin          *io.Pin
}

// Write writes the SwitchTimer to the embedded store in the given ScriptContext.
//
// Parameters:
// - ctx: A pointer to a ScriptContext object representing the context in which the write operation is being performed.
//
// Returns:
// - An error object if there was an error during the write operation, otherwise nil.
func (p *SwitchTimer) Write(ctx *ScriptContext) error {
	return ctx.EmbeddedStore.Upsert(p.Name, p)
}

// Process processes the SwitchTimer.
//
// It takes a ScriptContext and a gpio.PinIO as parameters.
// It returns an error.
func (p *SwitchTimer) Process(ctx *ScriptContext, dcb DurationCallback, pin gpio.PinIO) error {
	ctx.Logger.Debugf("process switchtimer %s", p.Name)

	if p.pin == nil {
		p.pin = io.NewPin(pin)
	}

	var onDuration, offDuration time.Duration

	if dcb != nil {
		onDuration, offDuration = dcb(p.OnDuration, p.OffDuration)
	} else {
		onDuration = p.OnDuration
		offDuration = p.OffDuration
	}

	if p.CurrentState == SwitchTimerStateInitialized {
		p.CurrentState = SwitchTimerStateOff
		p.CurrentSpan = NewTimespan(time.Now(), offDuration)

		if p.Inverted {
			p.pin.SetHigh()
		} else {
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
		p.CurrentSpan = NewTimespan(time.Now(), onDuration)

		if p.Inverted {
			p.pin.SetLow()
		} else {
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

		if p.Inverted {
			p.pin.SetHigh()
		} else {
			p.pin.SetLow()
		}

		ctx.Logger.Infof("switchtimer %s turned off", p.Name)
		return p.Write(ctx)
	}

	return nil
}
