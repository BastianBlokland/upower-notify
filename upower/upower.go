//go:generate stringer -type=State

// A minimal binding for UPower over DBUS.
// it is designed to be as simple as possible.

package upower

import (
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

var NoUpower = errors.New("Couldn't get org.freedesktop.UPower")

type State int

const (
	Unknown State = iota
	Charging
	Discharging
	Empty
	FullCharged
	PendingCharge
	PendingDischarge
)

type Update struct {
	EnergyRate  float64
	Percentage  float64
	State       State
	TimeToEmpty time.Duration
	TimeToFull  time.Duration
}

func (s *Update) Changed(old Update) bool {
	return *s != old
}

func New(device string) (*UPower, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	path := dbus.ObjectPath("/org/freedesktop/UPower/devices/" + device)
	up := conn.Object("org.freedesktop.UPower", path)
	if up == nil {
		return nil, NoUpower
	}

	return &UPower{dbus: up}, nil
}

type UPower struct {
	dbus dbus.BusObject
}

func getProp[T any](props map[string]dbus.Variant, key string) (T, error) {
	v, ok := props[key].Value().(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("unexpected type for %s", key)
	}
	return v, nil
}

func (u *UPower) Get() (Update, error) {
	props := map[string]dbus.Variant{}
	update := Update{}
	err := u.dbus.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.UPower.Device").Store(&props)
	if err != nil {
		return update, err
	}

	var stateRaw uint32
	var timeToEmpty, timeToFull int64

	if update.EnergyRate, err = getProp[float64](props, "EnergyRate"); err != nil {
		return update, err
	}
	if update.Percentage, err = getProp[float64](props, "Percentage"); err != nil {
		return update, err
	}
	if stateRaw, err = getProp[uint32](props, "State"); err != nil {
		return update, err
	}
	update.State = State(stateRaw)
	if timeToEmpty, err = getProp[int64](props, "TimeToEmpty"); err != nil {
		return update, err
	}
	update.TimeToEmpty = time.Duration(timeToEmpty) * time.Second
	if timeToFull, err = getProp[int64](props, "TimeToFull"); err != nil {
		return update, err
	}
	update.TimeToFull = time.Duration(timeToFull) * time.Second

	return update, nil
}
