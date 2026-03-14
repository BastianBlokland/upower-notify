package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/godbus/dbus/v5"

	"upower-notify/ppd"
	"upower-notify/notify"
	"upower-notify/upower"
)

const (
	lowBatteryThreshold    = 10.0
	mediumBatteryThreshold = 20.0
	notifyStepLow          = 1
	notifyStepMedium       = 5
	notifyStepDefault      = 20
)

var (
	initialOnly        bool
	tick               time.Duration
	notificationExpiry time.Duration
	device             string

	notificationExpiryMilliseconds int32
)

func main() {

	flag.BoolVar(&initialOnly, "initialOnly", false, "Exit after sending the initial notification.")
	flag.DurationVar(&tick, "tick", 60*time.Second, "Update rate.")
	flag.DurationVar(&notificationExpiry, "notification-expiration", 10*time.Second, "Notifications expiry duration.")
	flag.StringVar(&device, "device", "DisplayDevice", "DBus device name for the battery.")
	flag.Parse()

	notificationExpiryMilliseconds = int32(notificationExpiry / time.Millisecond)

	dbusConnSystem, err := dbus.SystemBus()
	if err != nil {
		log.Fatal(err)
	}

	dbusConnSession, err := dbus.SessionBus()
	if err != nil {
		log.Fatal(err)
	}

	up := upower.New(dbusConnSystem, device)

	update, err := up.Get()
	if err != nil {
		log.Fatal(err)
	}

	pd := ppd.New(dbusConnSystem)
	var profileState ppd.State
	profileState, err = pd.Get()
	if err != nil {
		// PPD is optional; if unavailable treat it as absent.
		pd = nil
	}

	notifier := notify.New(dbusConnSession, "Upower Notify")

	notifyState(update, profileState, notifier)

	if initialOnly {
		return
	}

	dbusSignals := make(chan *dbus.Signal, 16)
	dbusConnSystem.Signal(dbusSignals)
	if err := up.AddMatchSignal(dbusConnSystem); err != nil {
		log.Fatal(err)
	}
	if pd != nil {
		if err := pd.AddMatchSignal(dbusConnSystem); err != nil {
			log.Fatal(err)
		}
	}

	var old = update
	var oldProfileState = profileState
	var lastNotifyPercentage = old.Percentage
	poll := time.NewTicker(tick)
	for {
		select {
		case <-dbusSignals:
		case <-poll.C:
		}

		newUpdate, err := up.Get()
		if err != nil {
			notifier.Critical("Battery", fmt.Sprintf("Query failed: %s", err), notificationExpiryMilliseconds)
			continue
		}
		update = newUpdate

		if pd != nil {
			newProfileState, err := pd.Get()
			if err != nil {
				notifier.Critical("Battery", fmt.Sprintf("Profile (PPD) query failed: %s", err), notificationExpiryMilliseconds)
				continue
			}
			profileState = newProfileState
		}

		if update.Changed(old) || profileState.Changed(oldProfileState) {
			var charging = update.State == upower.Charging || update.State == upower.FullCharged || update.State == upower.PendingCharge

			var notifyStep uint32
			switch {
			case !charging && update.Percentage < lowBatteryThreshold:
				notifyStep = notifyStepLow
			case !charging && update.Percentage < mediumBatteryThreshold:
				notifyStep = notifyStepMedium
			default:
				notifyStep = notifyStepDefault
			}

			stateChanged := update.State != old.State && update.State != upower.PendingCharge && update.State != upower.PendingDischarge
			profileChanged := profileState.ActiveProfile != oldProfileState.ActiveProfile
			percentageChanged := (uint32(update.Percentage)/notifyStep) != (uint32(lastNotifyPercentage)/notifyStep)

			if stateChanged || profileChanged || percentageChanged {
				notifyState(update, profileState, notifier)
				lastNotifyPercentage = update.Percentage
			}
		}
		old = update
		oldProfileState = profileState
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	var result string
	if hours > 0 {
		result += fmt.Sprintf("%d hour", hours)
		if hours > 1 {
			result += "s"
		}
	}
	if minutes > 0 {
		if result != "" {
			result += " and "
		}
		result += fmt.Sprintf("%d minute", minutes)
		if minutes > 1 {
			result += "s"
		}
	}
	return result
}

func notifyState(battery upower.Update, profileState ppd.State, notifier *notify.Notifier) {
	var msg string
	var invalidState bool
	switch battery.State {
	case upower.Charging, upower.FullCharged, upower.PendingCharge, upower.PendingDischarge:
		msg = fmt.Sprintf("%.0f%% %s", battery.Percentage, battery.State)
		var strTillFull = formatDuration(battery.TimeToFull)
		if strTillFull != "" {
			msg += fmt.Sprintf("\n%s until full", strTillFull)
		}
		if battery.EnergyRate != 0.0 {
			msg += fmt.Sprintf("\n%.1f W usage", battery.EnergyRate)
		}
	case upower.Empty:
		msg = fmt.Sprintf("%.0f%%", battery.Percentage)
		if battery.EnergyRate != 0.0 {
			msg += fmt.Sprintf("\n%.1f W usage", battery.EnergyRate)
		}
	case upower.Discharging:
		msg = fmt.Sprintf("%.0f%% %s", battery.Percentage, battery.State)
		var strTillEmpty = formatDuration(battery.TimeToEmpty)
		if strTillEmpty != "" {
			msg += fmt.Sprintf("\n%s until empty", strTillEmpty)
		}
		if battery.EnergyRate != 0.0 {
			msg += fmt.Sprintf("\n%.1f W usage", battery.EnergyRate)
		}
	default:
		msg = fmt.Sprintf("%.0f%% Invalid State", battery.Percentage)
		invalidState = true
	}
	if profileState.ActiveProfile != "" {
		msg += fmt.Sprintf(" (%s profile)", profileState.ActiveProfile)
	}
	if invalidState || battery.Percentage < lowBatteryThreshold {
		notifier.Critical("Battery", msg, notificationExpiryMilliseconds)
	} else {
		notifier.Normal("Battery", msg, notificationExpiryMilliseconds)
	}
}
