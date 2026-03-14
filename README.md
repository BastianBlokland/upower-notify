
# UPower-Notify

A lightweight daemon that sends desktop notifications about your battery status on Linux.
It monitors battery state changes via [UPower](https://upower.freedesktop.org/) over DBus and
sends notifications when the battery state or charge level changes significantly.

## Features

- Notifications on charging/discharging state changes
- Increasing notification frequency as the battery gets low (every 20% normally, every 5% below 20%, every 1% below 10%)
- Time-to-empty and time-to-full estimates in notifications
- Power usage in watts
- Optional [Power Profiles Daemon (PPD)](https://gitlab.freedesktop.org/hadess/power-profiles-daemon) integration to show the active power profile
- Reacts to DBus signals for low-latency updates, with a periodic poll as fallback

## Requirements

- [UPower](https://upower.freedesktop.org/)
- A desktop notification daemon (e.g. `dunst`, `mako`, `notify-osd`)
- Optional: [Power Profiles Daemon](https://gitlab.freedesktop.org/hadess/power-profiles-daemon)

## Usage

Launch `upower-notify` when starting your desktop environment.

```
upower-notify [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-tick` | `60s` | How often to poll for battery state as a fallback |
| `-notification-expiration` | `10s` | How long notifications are shown |
| `-device` | `DisplayDevice` | DBus device name for the battery |
| `-initialOnly` | `false` | Send one notification then exit |

## Credits

Based on [omeid/upower-notify](https://github.com/omeid/upower-notify) by [omeid](https://github.com/omeid).
