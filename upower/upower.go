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
	Capacity         float64
	Energy           float64
	EnergyEmpty      float64
	EnergyFull       float64
	EnergyFullDesign float64
	EnergyRate       float64
	HasHistory       bool
	HasStatistics    bool
	IconName         string
	IsPresent        bool
	IsRechargeable   bool
	Luminosity       float64
	Model            string
	NativePath       string
	Online           bool
	Percentage       float64
	PowerSupply      bool
	Serial           string
	State            State
	Technology       uint32
	Temperature      float64
	TimeToEmpty      time.Duration
	TimeToFull       time.Duration
	Type             uint32
	UpdateTime       uint64
	Vendor           string
	Voltage          float64
	WarningLevel     uint32
}

func (s *Update) Changed(old Update) bool {
	if s.Capacity != old.Capacity {
		return true
	}

	if s.Energy != old.Energy {
		return true
	}

	if s.EnergyEmpty != old.EnergyEmpty {
		return true
	}

	if s.EnergyFull != old.EnergyFull {
		return true
	}

	if s.EnergyFullDesign != old.EnergyFullDesign {
		return true
	}

	if s.EnergyRate != old.EnergyRate {
		return true
	}

	if s.HasHistory != old.HasHistory {
		return true
	}

	if s.HasStatistics != old.HasStatistics {
		return true
	}

	if s.IconName != old.IconName {
		return true
	}

	if s.IsPresent != old.IsPresent {
		return true
	}

	if s.IsRechargeable != old.IsRechargeable {
		return true
	}

	if s.Luminosity != old.Luminosity {
		return true
	}

	if s.Model != old.Model {
		return true
	}

	if s.NativePath != old.NativePath {
		return true
	}

	if s.Online != old.Online {
		return true
	}

	if s.Percentage != old.Percentage {
		return true
	}

	if s.PowerSupply != old.PowerSupply {
		return true
	}

	if s.Serial != old.Serial {
		return true
	}

	if s.State != old.State {
		return true
	}

	if s.Technology != old.Technology {
		return true
	}

	if s.Temperature != old.Temperature {
		return true
	}

	if s.TimeToEmpty != old.TimeToEmpty {
		return true
	}

	if s.TimeToFull != old.TimeToFull {
		return true
	}

	if s.Type != old.Type {
		return true
	}

	if s.UpdateTime != old.UpdateTime {
		return true
	}

	if s.Vendor != old.Vendor {
		return true
	}

	if s.Voltage != old.Voltage {
		return true
	}

	if s.WarningLevel != old.WarningLevel {
		return true
	}

	return false

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

	if update.Capacity, err = getProp[float64](props, "Capacity"); err != nil {
		return update, err
	}
	if update.Energy, err = getProp[float64](props, "Energy"); err != nil {
		return update, err
	}
	if update.EnergyEmpty, err = getProp[float64](props, "EnergyEmpty"); err != nil {
		return update, err
	}
	if update.EnergyFull, err = getProp[float64](props, "EnergyFull"); err != nil {
		return update, err
	}
	if update.EnergyFullDesign, err = getProp[float64](props, "EnergyFullDesign"); err != nil {
		return update, err
	}
	if update.EnergyRate, err = getProp[float64](props, "EnergyRate"); err != nil {
		return update, err
	}
	if update.HasHistory, err = getProp[bool](props, "HasHistory"); err != nil {
		return update, err
	}
	if update.HasStatistics, err = getProp[bool](props, "HasStatistics"); err != nil {
		return update, err
	}
	if update.IconName, err = getProp[string](props, "IconName"); err != nil {
		return update, err
	}
	if update.IsPresent, err = getProp[bool](props, "IsPresent"); err != nil {
		return update, err
	}
	if update.IsRechargeable, err = getProp[bool](props, "IsRechargeable"); err != nil {
		return update, err
	}
	if update.Luminosity, err = getProp[float64](props, "Luminosity"); err != nil {
		return update, err
	}
	if update.Model, err = getProp[string](props, "Model"); err != nil {
		return update, err
	}
	if update.NativePath, err = getProp[string](props, "NativePath"); err != nil {
		return update, err
	}
	if update.Online, err = getProp[bool](props, "Online"); err != nil {
		return update, err
	}
	if update.Percentage, err = getProp[float64](props, "Percentage"); err != nil {
		return update, err
	}
	if update.PowerSupply, err = getProp[bool](props, "PowerSupply"); err != nil {
		return update, err
	}
	if update.Serial, err = getProp[string](props, "Serial"); err != nil {
		return update, err
	}
	if stateRaw, err = getProp[uint32](props, "State"); err != nil {
		return update, err
	}
	update.State = State(stateRaw)
	if update.Technology, err = getProp[uint32](props, "Technology"); err != nil {
		return update, err
	}
	if update.Temperature, err = getProp[float64](props, "Temperature"); err != nil {
		return update, err
	}
	if timeToEmpty, err = getProp[int64](props, "TimeToEmpty"); err != nil {
		return update, err
	}
	update.TimeToEmpty = time.Duration(timeToEmpty) * time.Second
	if timeToFull, err = getProp[int64](props, "TimeToFull"); err != nil {
		return update, err
	}
	update.TimeToFull = time.Duration(timeToFull) * time.Second
	if update.Type, err = getProp[uint32](props, "Type"); err != nil {
		return update, err
	}
	if update.UpdateTime, err = getProp[uint64](props, "UpdateTime"); err != nil {
		return update, err
	}
	if update.Vendor, err = getProp[string](props, "Vendor"); err != nil {
		return update, err
	}
	if update.Voltage, err = getProp[float64](props, "Voltage"); err != nil {
		return update, err
	}
	if update.WarningLevel, err = getProp[uint32](props, "WarningLevel"); err != nil {
		return update, err
	}

	return update, nil
}
