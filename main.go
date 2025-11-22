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
	for range time.Tick(tick) {
		update, err = up.Get()
		if err != nil {
			notifier.Critical("Battery", fmt.Sprintf("Failed to query status: %s", err), notificationExpiryMilliseconds)
			fmt.Printf("Failed to query status")
		}
		if update.Changed(old) {
			sendNotify(update, notifier, old.State != update.State)
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
	if battery.State == upower.Charging || battery.State == upower.FullCharged {
		notifier.Normal("Battery",
			fmt.Sprintf("%.0f%% %s\n%s until full\n%.1f W usage",
				battery.Percentage,
				battery.State,
				formatDuration(battery.TimeToFull),
				battery.EnergyRate),
			notificationExpiryMilliseconds)
	} else {
		notifier.Normal("Battery",
			fmt.Sprintf("%.0f%% %s\n%s until empty\n%.1f W usage",
				battery.Percentage,
				battery.State,
				formatDuration(battery.TimeToEmpty),
				battery.EnergyRate),
			notificationExpiryMilliseconds)
	}
}

func sendNotify(battery upower.Update, notifier *notify.Notifier, changed bool) {
	if changed {
		notifier.Normal("Battery", fmt.Sprintf("Battery state changed, %s.", battery.State), notificationExpiryMilliseconds)
	}

	switch battery.State {
	case upower.Charging:
	case upower.Discharging:
		switch {
		case battery.TimeToEmpty < critical:
			notifier.Critical("Battery", fmt.Sprintf("Battery level critical, %s remaining.", battery.TimeToEmpty), notificationExpiryMilliseconds)
			time.Sleep(critical / 10)
		case battery.TimeToEmpty < warn:
			notifier.Normal("Battery", fmt.Sprintf("Battery level low, %s remaining.", battery.TimeToEmpty), notificationExpiryMilliseconds)
			time.Sleep(warn / 10)
		default:
		}
	case upower.Empty:
		notifier.Critical("Battery", fmt.Sprintf("Battery exhausted, %s remaining.", battery.TimeToEmpty), notificationExpiryMilliseconds)
	case upower.FullCharged, upower.PendingCharge, upower.PendingDischarge:
	default:
		notifier.Critical("Battery", "Failed to query battery state.", notificationExpiryMilliseconds)
	}
}
