// A minimal binding for PPD (Power Profiles Daemon) to query the
// active power-profile over DBUS.

package ppd

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

var NoPowerProfileDaemon = errors.New("Couldn't get net.hadess.PowerProfiles")
var MalformedResponse = errors.New("Invalid response from PPD")

type State struct {
	ActiveProfile string
}

type PowerProfileDaemon struct {
	dbus dbus.BusObject
}

func New(conn *dbus.Conn) (*PowerProfileDaemon, error) {
	path := dbus.ObjectPath("/net/hadess/PowerProfiles")
	obj := conn.Object("net.hadess.PowerProfiles", path)
	if obj == nil {
		return nil, NoPowerProfileDaemon
	}
	return &PowerProfileDaemon{dbus: obj}, nil
}

func (s *State) Changed(old State) bool {
	return s.ActiveProfile != old.ActiveProfile
}

func (p *PowerProfileDaemon) AddMatchSignal(conn *dbus.Conn) error {
	return conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
		dbus.WithMatchObjectPath("/net/hadess/PowerProfiles"),
		dbus.WithMatchArg(0, "net.hadess.PowerProfiles"),
	)
}

func (ppd *PowerProfileDaemon) Get() (State, error) {
	state := State{}
	variant, err := ppd.dbus.GetProperty("net.hadess.PowerProfiles.ActiveProfile")
	if err != nil {
		return state, err
	}
	activeProfile, ok := variant.Value().(string)
	if !ok {
		return state, MalformedResponse
	}
	state.ActiveProfile = activeProfile
	return state, nil
}
