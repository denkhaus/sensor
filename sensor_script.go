package main

import (
	"time"

	"github.com/denkhaus/sensor/store"
	"github.com/denkhaus/sensor/types"
	"github.com/pkg/errors"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3/rpi"
)

const (
	ECMinThreshold            = 0.4
	ECMaxThreshold            = 1.7
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

func processAquaPump(ctx *types.ScriptContext, pumpStateID string, pinio gpio.PinIO) error {
	ps, err := readAquaPumpState(ctx, pumpStateID)
	if err != nil {
		return errors.Wrapf(err, "readAquaPumpState %s", pumpStateID)
	}

	if err := ps.Process(ctx, pinio); err != nil {
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
			cond := ctx.SensorStore.Get(store.Conductivity)
			return cond >= ECMinThreshold && cond < ECMaxThreshold
		}
		return false
	}

	if err := ps.Process(ctx, fnCondition, pinio); err != nil {
		return errors.Wrapf(err, "process dose pump %s", pumpStateID)
	}

	return nil
}

func setup(ctx *types.ScriptContext) error {
	_, err := readDosePumpState(ctx, DosePumpStateIDDefault)
	if store.IsDocumentNotFoundError(err) {
		ctx.Logger.Infof("setup pumpstatus %s", DosePumpStateIDDefault)

		status1 := types.PulseTimer{
			Name:              DosePumpStateIDDefault,
			PulseOnInitialize: true,
			WarnOnPinError:    false,
			Description:       "The dose pump status",
			PulseDuration:     time.Second * 2,
			WaitDuration:      time.Minute * 2,
		}

		if err := ctx.EmbeddedStore.Upsert(status1.Name, status1); err != nil {
			return errors.Wrapf(err, "upsert pumpstatus %s", status1.Name)
		}
	}

	_, err = readAquaPumpState(ctx, AquaPumpStateIDGreenhouse)
	if store.IsDocumentNotFoundError(err) {
		ctx.Logger.Infof("setup pumpstatus %s", AquaPumpStateIDGreenhouse)

		status1 := types.SwitchTimer{
			Name:           AquaPumpStateIDGreenhouse,
			Description:    "The greenhouse pump status",
			WarnOnPinError: false,
			CurrentState:   types.SwitchTimerStateInitialized,
			OnDuration:     time.Second * 3,
			OffDuration:    time.Second * 60,
		}

		if err := ctx.EmbeddedStore.Upsert(status1.Name, status1); err != nil {
			return errors.Wrapf(err, "upsert pumpstatus %s", status1.Name)
		}
	}

	_, err = readAquaPumpState(ctx, AquaPumpStateIDHydroRack)
	if store.IsDocumentNotFoundError(err) {
		ctx.Logger.Infof("setup pumpstatus %s", AquaPumpStateIDHydroRack)

		status2 := types.SwitchTimer{
			Name:           AquaPumpStateIDHydroRack,
			Description:    "The hydrorack pump status",
			WarnOnPinError: false,
			CurrentState:   types.SwitchTimerStateInitialized,
			OnDuration:     time.Second * 10,
			OffDuration:    time.Second * 20,
		}

		if err := ctx.EmbeddedStore.Upsert(status2.Name, status2); err != nil {
			return errors.Wrapf(err, "upsert pumpstatus %s", status2.Name)
		}
	}

	return nil
}

func Script(ctx *types.ScriptContext) error {
	if err := setup(ctx); err != nil {
		return errors.Wrap(err, "setup")
	}

	cond := ctx.SensorStore.Get(store.Conductivity)
	hum := ctx.SensorStore.Get(store.Humidity)
	ctx.Logger.Infof("EC: %f, Humidity: %f, ", cond, hum)

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
