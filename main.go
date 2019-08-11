package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"text/template"
	"time"

	"github.com/apognu/gocal"
)

// argvT : command line arguments
type argvT struct {
	url          string
	outputFormat string
	dateFormat   string
	dryrun       bool
	wait         bool
	waitMax      int64
	waitMin      int64
	verbose      int
	start        time.Time
	end          time.Time
}

type eventT struct {
	gocal.Event
	Epoch  int64
	Diff   int64
	Date   string
	Status string
}

const (
	version      = "0.3.0"
	formatStdout = `{{.Epoch}} {{.Diff}} {{.Status}} {{ .Summary | urlquery -}}
{{- if .Description }} {{ .Description | urlquery }}
{{- else }} -
{{- end -}}
{{- if .Location }} {{ .Location | urlquery }}
{{- else }} -
{{- end }}
`
	formatMessage = `{{.Date}}: {{.Status}}: {{.Summary}}
{{- if .Location }}
Location: {{.Location}}
{{- end }}
{{- if .Description}}
Description: {{.Description}}
{{- end }}
`
)

func args() *argvT {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, `%s v%s
  Usage: %s [<option>] <url>

`, path.Base(os.Args[0]), version, os.Args[0])
		flag.PrintDefaults()
	}

	start := flag.Int64("start", 0,
		"Start time in epoch seconds (default: now)")
	dryrun := flag.Bool("dryrun", false,
		"Do not sleep")
	duration := flag.Duration("duration", 15*time.Minute,
		"Duration to check for events")
	outputFormat := flag.String("output-format", "",
		"Template for formatting output")
	dateFormat := flag.String("date-format",
		"Mon Jan _2 15:04:05 MST 2006",
		"Format for date string")
	wait := flag.Bool("wait", false,
		"Wait for first event")
	waitMax := flag.Int64("wait-max", 0,
		"Maximum time to wait for next event")
	waitMin := flag.Int64("wait-min", 0,
		"Minimum amount of time to poll for new events")
	verbose := flag.Int("verbose", 0,
		"Enable debug messages")

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	startTime := time.Now()
	if *start != 0 {
		startTime = time.Unix(*start, 0)
	}

	return &argvT{
		url:          flag.Arg(0),
		outputFormat: *outputFormat,
		dateFormat:   *dateFormat,
		dryrun:       *dryrun,
		wait:         *wait,
		waitMax:      *waitMax,
		waitMin:      *waitMin,
		start:        startTime,
		end:          startTime.Add(*duration),
		verbose:      *verbose,
	}
}

func main() {
	argv := args()

	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	h := &http.Client{Transport: t}
	resp, err := h.Get(argv.url)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	c := gocal.NewParser(resp.Body)
	c.Start = &argv.start
	c.End = &argv.end

	err = c.Parse()

	if err != nil {
		log.Fatalln(err)
	}

	m := make(map[int64]bool)
	event := make(map[int64]eventT)

	for _, e := range c.Events {
		start := e.Start.Unix()
		end := e.End.Unix()

		if argv.verbose > 1 {
			fmt.Printf("%+v\n", e)
		}

		if e.Start.UnixNano() >= argv.start.UnixNano() {
			event[start] = eventT{
				Event:  e,
				Epoch:  e.Start.Unix(),
				Diff:   e.Start.Unix() - argv.start.Unix(),
				Date:   e.Start.Local().Format(argv.dateFormat),
				Status: "start",
			}
			m[start] = true
		}

		event[end] = eventT{
			Event:  e,
			Epoch:  e.End.Unix(),
			Diff:   e.End.Unix() - argv.start.Unix(),
			Date:   e.End.Local().Format(argv.dateFormat),
			Status: "end",
		}
		m[end] = true
	}

	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	ev := waitEvent
	if !argv.wait {
		ev = writeEvent
	}

	err = ev(argv, keys, event)

	if err != nil {
		log.Fatalln(err)
	}
}

func waitfor(argv *argvT, seconds int64) {
	if argv.dryrun {
		fmt.Printf("sleep: %d\n", seconds)
		return
	}
	time.Sleep(time.Duration(seconds) * time.Second)
}

func waitEvent(argv *argvT, keys []int64, event map[int64]eventT) error {
	if len(keys) == 0 {
		waitfor(argv, argv.waitMin)
		return nil
	}
	e := event[keys[0]]
	output := true
	seconds := e.Diff
	if argv.waitMax > 0 && seconds > argv.waitMax {
		output = false
		seconds = argv.waitMax
	}
	waitfor(argv, seconds)
	if !output {
		return nil
	}
	format := formatMessage
	if argv.outputFormat != "" {
		format = argv.outputFormat
	}
	return formatEvent(format, e)
}

func writeEvent(argv *argvT, keys []int64, event map[int64]eventT) error {
	format := formatStdout
	if argv.outputFormat != "" {
		format = argv.outputFormat
	}
	for _, k := range keys {
		err := formatEvent(format, event[k])
		if err != nil {
			return err
		}
	}
	return nil
}

func formatEvent(format string, e eventT) error {
	tmpl, err := template.New("format").Parse(format)
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, e)
}
