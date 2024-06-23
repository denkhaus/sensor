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
}

func (p *PulseTimer) Write(ctx *ScriptContext) error {
	return ctx.EmbeddedStore.Upsert(p.Name, p)
}

func (p *PulseTimer) pulse(ctx *ScriptContext, fnCondition func() bool, pinio gpio.PinIO) error {
	pin := io.NewPin(pinio)

	// defer func() {
	// 	if pin != nil {
	// 		pin.Close()
	// 	}
	// }()

	if fnCondition() {
		ctx.Logger.Infof("pulsetimer %s: pulse for %s", p.Name, p.PulseDuration)
		if pin != nil {
			if p.Inverted {
				pin.PulseLow(p.PulseDuration)
			} else {
				pin.PulseHigh(p.PulseDuration)
			}
		}
	} else {
		ctx.Logger.Infof("pulsetimer %s: condition not met. try again in %s", p.Name, p.WaitDuration)
	}

	//reset wait timer
	p.CurrentSpan = NewTimespan(time.Now(), p.WaitDuration)
	return p.Write(ctx)
}

func (p *PulseTimer) Process(ctx *ScriptContext, fnCondition func() bool, pinio gpio.PinIO) error {
	ctx.Logger.Debugf("process pulsetimer %s", p.Name)

	if p.CurrentSpan.IsZero() {
		ctx.Logger.Infof("initialize pulsetimer %s", p.Name)
		if p.PulseOnInitialize {
			if err := p.pulse(ctx, fnCondition, pinio); err != nil {
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

	if err := p.pulse(ctx, fnCondition, pinio); err != nil {
		return err
	}

	return nil
}
