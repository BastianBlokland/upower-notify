//go:generate stringer -type=Urgency

// A minimal binding for DBus Desktop Notifications,
// it is designed to be as simple as send-notify.

package notify

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

var NoNotifications = errors.New("Couldn't get org.freedesktop.Notifications")

type Urgency byte

const (
	Low Urgency = iota
	Normal
	Critical
)

type Notifier struct {
	dbus dbus.BusObject
	app  string
}

func New(conn *dbus.Conn, app string) (*Notifier, error) {
	notification := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	if notification == nil {
		return nil, NoNotifications
	}

	return &Notifier{dbus: notification, app: app}, nil
}

func (n *Notifier) Low(Summary string, Body string, ExpireTimeout int32) error {
	return n.Send(Summary, Body, Low, ExpireTimeout)
}

func (n *Notifier) Normal(Summary string, Body string, ExpireTimeout int32) error {
	return n.Send(Summary, Body, Normal, ExpireTimeout)
}

func (n *Notifier) Critical(Summary string, Body string, ExpireTimeout int32) error {
	return n.Send(Summary, Body, Critical, ExpireTimeout)
}

func (n *Notifier) Send(summary string, body string, urgency Urgency, expireTimeout int32) error {
	return n.dbus.Call("org.freedesktop.Notifications.Notify", 0,
		n.app, // app_name
		uint32(0), // replaces_id
		"", // app_icon
		summary, // summary
		body, // body
		[]string{}, // actions
		map[string]dbus.Variant{"urgency": dbus.MakeVariant(urgency)}, // hints
		expireTimeout,
	).Err
}
