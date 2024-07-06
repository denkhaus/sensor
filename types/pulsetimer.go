package types

import (
	"encoding/gob"
	"time"

	"github.com/denkhaus/sensor/io"
	"periph.io/x/conn/v3/gpio"
)

func init() {
	gob.Register(PulseTimer{})
}

type PulseTimer struct {
	Name              string
	Description       string
	CurrentSpan       Span
	PulseDuration     time.Duration
	WaitDuration      time.Duration
	PulseOnInitialize bool
	Inverted          bool
	pin               *io.Pin
}

// Write writes the PulseTimer to the embedded store in the given ScriptContext.
//
// Parameters:
// - ctx: A pointer to a ScriptContext object representing the context in which the write operation is being performed.
// Return type(s).
func (p *PulseTimer) Write(ctx *ScriptContext) error {
	return ctx.EmbeddedStore.Upsert(p.Name, p)
}

// pulse performs a pulse operation based on the condition and updates the timer.
//
// It takes a ScriptContext and a function returning a boolean as parameters.
// It returns an error.
func (p *PulseTimer) pulse(ctx *ScriptContext, fnCondition func() bool) error {

	if fnCondition() {
		ctx.Logger.Infof("pulsetimer %s: pulse for %s", p.Name, p.PulseDuration)

		// pulse high
		if p.Inverted {
			p.pin.PulseLow(p.PulseDuration)
		} else {
			p.pin.PulseHigh(p.PulseDuration)
		}

	} else {
		ctx.Logger.Infof("pulsetimer %s: condition not met. try again in %s", p.Name, p.WaitDuration)

		// set low
		if p.Inverted {
			p.pin.SetHigh()
		} else {
			p.pin.SetLow()
		}
	}

	//reset wait timer
	p.CurrentSpan = NewTimespan(time.Now(), p.WaitDuration)
	return p.Write(ctx)
}

// Process processes the PulseTimer.
//
// It takes a ScriptContext, a function fnCondition that returns a boolean,
// and a gpio.PinIO as parameters.
// It returns an error.
func (p *PulseTimer) Process(ctx *ScriptContext, fnCondition func() bool, pin gpio.PinIO) error {
	ctx.Logger.Debugf("process pulsetimer %s", p.Name)

	if p.pin == nil {
		p.pin = io.NewPin(pin)
	}

	if p.CurrentSpan.IsZero() {
		ctx.Logger.Infof("initialize pulsetimer %s", p.Name)
		if p.PulseOnInitialize {
			if err := p.pulse(ctx, fnCondition); err != nil {
				return err
			}
		} else {
			p.CurrentSpan = NewTimespan(time.Now(), p.WaitDuration)
		}

		return p.Write(ctx)
	}

	if p.CurrentSpan.ContainsTime(time.Now()) {
		return nil
	}

	if err := p.pulse(ctx, fnCondition); err != nil {
		return err
	}

	return nil
}
