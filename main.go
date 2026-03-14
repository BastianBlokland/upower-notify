package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"upower-notify/notify"
	"upower-notify/upower"
	"upower-notify/ppd"
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
	flag.DurationVar(&tick, "tick", 10*time.Second, "Update rate.")
	flag.DurationVar(&notificationExpiry, "notification-expiration", 10*time.Second, "Notifications expiry duration.")
	flag.StringVar(&device, "device", "DisplayDevice", "DBus device name for the battery.")
	flag.Parse()

	notificationExpiryMilliseconds = int32(notificationExpiry / time.Millisecond)
	up, err := upower.New(device)
	if err != nil {
		log.Fatal(err)
	}

	update, err := up.Get()
	if err != nil {
		log.Fatal(err)
	}

	powerProfileDaemon, err := ppd.New()
	var profileState ppd.State
	if powerProfileDaemon != nil {
		profileState, err = powerProfileDaemon.Get();
		if err != nil {
			log.Fatal(err)
		}
	}

	notifier, err := notify.New("Upower Agent")
	if err != nil {
		log.Fatal(err)
	}

	notifyState(update, profileState, notifier)

	if initialOnly {
		return
	}

	var old = update
	var oldProfileState = profileState
	var lastNotifyPercentage = old.Percentage
	for range time.Tick(tick) {
		newUpdate, err := up.Get()
		if err != nil {
			notifier.Critical("Battery", fmt.Sprintf("Query failed: %s", err), notificationExpiryMilliseconds)
			continue
		}
		update = newUpdate

		if powerProfileDaemon != nil {
			newProfileState, err := powerProfileDaemon.Get()
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
			case !charging && update.Percentage < 10:
				notifyStep = 1
			case !charging && update.Percentage < 20:
				notifyStep = 5
			default:
				notifyStep = 20
			}

			stateChanged := update.State != old.State
			profileChanged := profileState.ActiveProfile != oldProfileState.ActiveProfile
			percentageChanged := (uint32(update.Percentage)/notifyStep) != (uint32(lastNotifyPercentage)/notifyStep);

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
		if len(result) > 0 {
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
	var invalidState bool = false
	switch battery.State {
	case upower.Charging, upower.FullCharged, upower.PendingCharge, upower.PendingDischarge:
		msg = fmt.Sprintf("%.0f%% %s", battery.Percentage, battery.State)
		var strTillFull = formatDuration(battery.TimeToFull)
		if len(strTillFull) != 0 {
			msg += fmt.Sprintf("\n%s until full", strTillFull)
		}
		if battery.EnergyRate != 0.0 {
			msg += fmt.Sprintf("\n%.1f W usage", battery.EnergyRate)
		}
		break
	case upower.Empty:
		msg = fmt.Sprintf("%.0f%%", battery.Percentage)
		if battery.EnergyRate != 0.0 {
			msg += fmt.Sprintf("\n%.1f W usage", battery.EnergyRate)
		}
		break
	case upower.Discharging:
		msg = fmt.Sprintf("%.0f%% %s", battery.Percentage, battery.State)
		var strTillEmpty = formatDuration(battery.TimeToEmpty)
		if len(strTillEmpty) != 0 {
			msg += fmt.Sprintf("\n%s until empty", strTillEmpty)
		}
		if battery.EnergyRate != 0.0 {
			msg += fmt.Sprintf("\n%.1f W usage", battery.EnergyRate)
		}
		break
	default:
		msg = fmt.Sprintf("%.0f%% Invalid State", battery.Percentage)
		invalidState = true
		break
	}
	if len(profileState.ActiveProfile) > 0 {
		msg += fmt.Sprintf(" (%s profile)", profileState.ActiveProfile)
	}
	if invalidState || battery.Percentage < 10 {
		notifier.Critical("Battery", msg, notificationExpiryMilliseconds)
	} else {
		notifier.Normal("Battery", msg, notificationExpiryMilliseconds)
	}
}
