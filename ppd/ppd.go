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

func New() (*PowerProfileDaemon, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
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
