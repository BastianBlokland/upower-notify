package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bastianblokland/upower-notify/notify"
	"github.com/bastianblokland/upower-notify/upower"
)

var (
	initialOnly        bool
	tick               time.Duration
	warn               time.Duration
	critical           time.Duration
	notificationExpiry time.Duration
	device             string

	notificationExpiryMilliseconds int32
)

func main() {

	flag.BoolVar(&initialOnly, "initialOnly", false, "Exit after sending the initial notification.")
	flag.DurationVar(&tick, "tick", 10*time.Second, "Update rate.")
	flag.DurationVar(&warn, "warn", 20*time.Minute, "Time to start warning (Warn).")
	flag.DurationVar(&critical, "critical", 10*time.Minute, "Time to start warning (Critical).")
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

	notifier, err := notify.New("Upower Agent")
	if err != nil {
		log.Fatal(err)
	}

	notifyState(update, notifier)

	if initialOnly {
		return
	}

	var old = update
	var lastNotifyPercentage = old.Percentage
	for range time.Tick(tick) {
		update, err = up.Get()
		if err != nil {
			notifier.Critical("Battery", fmt.Sprintf("Query failed: %s", err), notificationExpiryMilliseconds)
		}
		if update.Changed(old) {
			var notifyStep uint32
			switch {
			case update.TimeToEmpty < critical:
				notifyStep = 1
			case update.TimeToEmpty < warn:
				notifyStep = 5
			default:
				notifyStep = 20
			}
			if (old.State != update.State) || (uint32(update.Percentage)/notifyStep) != (uint32(lastNotifyPercentage)/notifyStep) {
				notifyState(update, notifier)
				lastNotifyPercentage = update.Percentage
			}
		}
		old = update
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

func notifyState(battery upower.Update, notifier *notify.Notifier) {
	var msg string
	var invalidState bool = false
	switch battery.State {
	case upower.Charging, upower.FullCharged, upower.PendingDischarge:
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
	case upower.Discharging, upower.PendingCharge:
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
	if invalidState || battery.TimeToEmpty < critical || battery.Percentage < 10 {
		notifier.Critical("Battery", msg, notificationExpiryMilliseconds)
	} else {
		notifier.Normal("Battery", msg, notificationExpiryMilliseconds)
	}
}
