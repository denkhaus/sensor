package types

import (
	"encoding/gob"
	"time"

	"github.com/denkhaus/sensor/io"
)

func init() {
	gob.Register(PulseTimer{})
}

type PulseTimer struct {
	Name              string
	PinName           string
	Description       string
	CurrentSpan       Span
	PulseDuration     time.Duration
	WaitDuration      time.Duration
	PulseOnInitialize bool
	WarnOnPinError    bool
}

func (p *PulseTimer) Write(ctx *ScriptContext) error {
	return ctx.EmbeddedStore.Upsert(p.Name, p)
}

func (p *PulseTimer) pulse(ctx *ScriptContext, fnCondition func() bool) error {
	pin, err := io.NewOutputPin(p.PinName)
	if err != nil && p.WarnOnPinError {
		ctx.Logger.Warnf("open digital output pin %s: %v", p.PinName, err)
	}

	defer func() {
		if pin != nil {
			pin.Close()
		}
	}()

	if fnCondition() {
		ctx.Logger.Infof("pulsetimer %s: pulse for %s", p.Name, p.PulseDuration)
		if pin != nil {
			pin.Pulse(p.PulseDuration)
		}
	} else {
		ctx.Logger.Infof("pulsetimer %s: condition not met. try again in %s", p.Name, p.WaitDuration)
	}

	//reset wait timer
	p.CurrentSpan = NewTimespan(time.Now(), p.WaitDuration)
	return p.Write(ctx)
}

func (p *PulseTimer) Process(ctx *ScriptContext, fnCondition func() bool) error {
	ctx.Logger.Debugf("process pulsetimer %s", p.Name)

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
