package main

import (
	"time"

	"github.com/denkhaus/containers"
	"github.com/denkhaus/sensor/store"
	"github.com/denkhaus/sensor/types"
	"github.com/pkg/errors"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3/rpi"
)

const (
	ECMinThreshold            = 0.4
	ECMaxThreshold            = 2.0
	AquaPumpStateIDGreenhouse = "AquaPumpGreenhouse"
	AquaPumpStateIDHydroRack  = "AquaPumpHydroRack"
	DosePumpStateIDDefault    = "DosePumpDefault"
)

func readAquaPumpState(ctx *types.ScriptContext, aquaPumpStateID string) (*types.SwitchTimer, error) {
	var pumpState types.SwitchTimer
	err := ctx.EmbeddedStore.Get(aquaPumpStateID, &pumpState)
	return &pumpState, err
}

func readDosePumpState(ctx *types.ScriptContext, dosePumpStateID string) (*types.PulseTimer, error) {
	var pumpState types.PulseTimer
	err := ctx.EmbeddedStore.Get(dosePumpStateID, &pumpState)
	return &pumpState, err
}

func durationCallback(onDuration time.Duration, offDuration time.Duration) (time.Duration, time.Duration) {
	hour := time.Now().Hour()

	// Increase the offDuration if the hour is between 0 and 8
	// This is to enable a night mode where the pump runs less often
	if containers.BetweenInclusive(0, 8, hour) {
		offDuration *= 2
	}

	return onDuration, offDuration
}

func processAquaPump(ctx *types.ScriptContext, pumpStateID string, pinio gpio.PinIO) error {
	ps, err := readAquaPumpState(ctx, pumpStateID)
	if err != nil {
		return errors.Wrapf(err, "readAquaPumpState %s", pumpStateID)
	}

	if err := ps.Process(ctx, durationCallback, pinio); err != nil {
		return errors.Wrapf(err, "process aqua pump %s", pumpStateID)
	}

	return nil
}

func processDosePump(ctx *types.ScriptContext, pumpStateID string, pinio gpio.PinIO) error {
	ps, err := readDosePumpState(ctx, pumpStateID)
	if err != nil {
		return errors.Wrapf(err, "readDosePumpState %s", pumpStateID)
	}

	fnCondition := func() bool {
		hum := ctx.SensorStore.Get(store.Humidity)

		if hum >= 50.0 {
			cond := ctx.SensorStore.Get(store.ConductivityWeighted)
			return cond >= ECMinThreshold && cond < ECMaxThreshold
		} else {
			ctx.Logger.Warnf("humidity %f is too low", hum)
		}

		return false
	}

	if err := ps.Process(ctx, fnCondition, pinio); err != nil {
		return errors.Wrapf(err, "process dose pump %s", pumpStateID)
	}

	return nil
}

func Setup(ctx *types.ScriptContext) error {

	ctx.Logger.Infof("setup pumpstatus %s", DosePumpStateIDDefault)

	status1 := types.PulseTimer{
		Name:              DosePumpStateIDDefault,
		PulseOnInitialize: true,
		Inverted:          true,
		Description:       "The dose pump status",
		PulseDuration:     time.Second * 3,
		WaitDuration:      time.Minute * 5,
	}

	if err := ctx.EmbeddedStore.Upsert(status1.Name, status1); err != nil {
		return errors.Wrapf(err, "upsert pumpstatus %s", status1.Name)
	}

	ctx.Logger.Infof("setup pumpstatus %s", AquaPumpStateIDGreenhouse)

	status2 := types.SwitchTimer{
		Name:         AquaPumpStateIDGreenhouse,
		Description:  "The greenhouse pump status",
		Inverted:     true,
		CurrentState: types.SwitchTimerStateInitialized,
		OnDuration:   time.Second * 20,
		OffDuration:  time.Minute * 5,
	}

	if err := ctx.EmbeddedStore.Upsert(status2.Name, status2); err != nil {
		return errors.Wrapf(err, "upsert pumpstatus %s", status2.Name)
	}

	ctx.Logger.Infof("setup pumpstatus %s", AquaPumpStateIDHydroRack)

	status3 := types.SwitchTimer{
		Name:         AquaPumpStateIDHydroRack,
		Description:  "The hydrorack pump status",
		Inverted:     true,
		CurrentState: types.SwitchTimerStateInitialized,
		OnDuration:   time.Second * 10,
		OffDuration:  time.Second * 40,
	}

	if err := ctx.EmbeddedStore.Upsert(status3.Name, status3); err != nil {
		return errors.Wrapf(err, "upsert pumpstatus %s", status3.Name)
	}

	return nil
}

func Script(ctx *types.ScriptContext) error {
	condWeighted := ctx.SensorStore.Get(store.ConductivityWeighted)
	cond := ctx.SensorStore.Get(store.Conductivity)
	temp := ctx.SensorStore.Get(store.Temperature)
	hum := ctx.SensorStore.Get(store.Humidity)
	tds := ctx.SensorStore.Get(store.TDS)

	ctx.Logger.Infof("EC:[w: %f|r: %f], Humidity: %f, TDS: %f, Temp: %f", condWeighted, cond, hum, tds, temp)

	if err := processAquaPump(ctx, AquaPumpStateIDGreenhouse, rpi.P1_38); err != nil {
		return err
	}

	if err := processAquaPump(ctx, AquaPumpStateIDHydroRack, rpi.P1_32); err != nil {
		return err
	}

	if err := processDosePump(ctx, DosePumpStateIDDefault, rpi.P1_35); err != nil {
		return err
	}

	return nil
}
