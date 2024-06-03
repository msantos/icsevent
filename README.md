icsevent - sleep(1) using ICS calendar events

# SYNOPSIS

icsevent [*options*] *URL*

# DESCRIPTION

icsevent is a simple, minimal command line utility for polling ICS files
for calendar events. icsevent runs as part of a shell pipeline to send
notifications or kick off other jobs.

The ICS format is parsed using [gocal](https://github.com/apognu/gocal).

By default, events are converted to a percent-encoded, line oriented
format that can be piped to other utilities.

Using the `--wait` option will cause icsevent to sleep until the next
event.

## BUILDING

```
go build
```

## POLLING

When the `--wait` option is used, some options can be used to control the
polling interval. icsevent should be run under a supervisor process
like [daemontools](https://cr.yp.to/daemontools.html).

* `--wait-max=15m`

  The `--wait-max` option controls how long icsevent will sleep before
  polling for new events.

  For example, if the next event happens in 1 month, icsevent will:

  * exit after 15 minutes
  * be restarted by the supervisor process
  * poll for events

* `--wait-min=30m`

  The `--wait-min` option controls the polling rate when no events
  are found.

  If no events are found, icsevent will exit immediately by default and
  be restarted by the supervisor, polling the calendar service every second.

  Setting the value of the `--wait-min` option to "30m" limits the rate
  of polling to 1 connection every 30 minutes.

# EXAMPLES

## Dump events for the next 3 months

```
icsevent --duration="$((3*30*24))h" https://www.calendarlabs.com/ical-calendar/ics/39/Canada_Holidays.ics
```

## Modify formatting

```
FORMAT='{{if eq .State "start"}}
{{- .Date}}

{{.Summary}}
{{- if .Location }}
Location: {{.Location}}
{{- end }}
{{- if .Description}}
Description: {{.Description}}
{{- end }}
{{else}}
---
{{end}}'

icsevent --duration="$((3*30*24))h" \
  --output-format="$FORMAT" \
  https://www.calendarlabs.com/ical-calendar/ics/39/Canada_Holidays.ics
```

## Wait for next event

```
icsevent --wait --duration="$((3*30*24))h" \
  --output-format="$FORMAT" \
  https://www.calendarlabs.com/ical-calendar/ics/39/Canada_Holidays.ics
```

## Wait with poll intervals

```
icsevent --wait --wait-max=60m --wait-min=30m --duration="$((3*30*24))h" \
  https://www.calendarlabs.com/ical-calendar/ics/39/Canada_Holidays.ics
```

## daemontools: run scripts: sending to an XMPP groupchat

Uses [xmppipe](https://github.com/msantos/xmppipe).

To run:

```
svscan service
```

* service/20-icsevent/run

```bash
#!/bin/sh

URL="https://www.calendarlabs.com/ical-calendar/ics/39/Canada_Holidays.ics"

NOTIFYDIR="$TMPDIR/xmpp-notify"

# Use XEP-0393: Message Styling
FORMAT='{{.Date}}: {{.State}}: *{{.Summary}}*
{{- if .Location }}
_Location_: {{.Location}}
{{- end }}
{{- if .Description}}
_Description_: {{.Description}}
{{- end }}
{{- if .Attendees}}
_Attendees_:
{{ range .Attendees }}* {{ if .Cn }}{{.Cn}}{{else}}{{ .Value}}{{end}}
{{ end }}
{{- end }}
{{- if .Organizer}}
_Organizer_: {{.Organizer.Cn}}
{{- end }}
'

exec > "$NOTIFYDIR/pipe" 2>&1
exec icsevent \
    --output-format="$FORMAT" \
    --wait \
    --wait-min=900 \
    --wait-max=901 \
    --duration=24h \
    "$URL"
```

* service/10-xmppipe/run

```bash
#!/bin/sh

umask 0077

NOTIFYDIR="$TMPDIR/xmpp-notify"

export XMPPIPE_USERNAME="bot@example.com"
export XMPPIPE_PASSWORD="bot-password"

mkdir -p "$NOTIFYDIR"
rm -f "$NOTIFYDIR/pipe"
mkfifo "$NOTIFYDIR/pipe"
exec <> "$NOTIFYDIR/pipe"
exec xmppipe -o groupchat
```

# OPTIONS

--dryrun
: When running with `--wait`, do not actually sleep

--duration *duration*
: Window beginning at start time to check for events (default 15m0s)

--output-format *string*
: Template for formatting output using the [Go template
syntax](https://golang.org/pkg/text/template/)

--start *int*
: Start time in epoch seconds (default: now)

--verbose *int*
: Enable debug messages

--wait
: Wait for first event

--wait-max *int*
: Maximum time to wait for next event

--wait-min *int*
: Minimum amount of time to poll for new events

# TEMPLATE FUNCTIONS

## text

Converts HTML to [styled plain text](https://xmpp.org/extensions/xep-0393.html).

```
{{.Description | text}}
```

## match

Boolean regular expression match:

```
{{- if not (match "(?i)^Cancelled" .Summary) -}}
...
{{ end }}
```
