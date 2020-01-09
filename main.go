package main

import (
	"bufio"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
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
	State string
}

type reT struct {
	match   string
	replace string
}

const (
	version      = "0.6.0"
	formatStdout = `{{.Epoch}} {{.Diff}} {{.State}} {{ .Summary | urlquery -}}
{{- if .Description }} {{ .Description | urlquery }}
{{- else }} -
{{- end -}}
{{- if .Location }} {{ .Location | urlquery }}
{{- else }} -
{{- end }}
`
	formatMessage = `{{.Date}}: {{.State}}: {{.Summary}}
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
	duration := flag.Duration("duration", 24*time.Hour,
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
	help := flag.Bool("help", false,
		"Usage")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	var url string
	if flag.NArg() > 0 {
		url = flag.Arg(0)
	}

	startTime := time.Now()
	if *start != 0 {
		startTime = time.Unix(*start, 0)
	}

	return &argvT{
		url:          url,
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

	var r io.Reader
	if argv.url != "" {
		resp, err := http.Get(argv.url)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()
		r = resp.Body
	} else {
		r = os.Stdin
	}

	c := gocal.NewParser(r)
	c.Start = &argv.start
	c.End = &argv.end

	if err := c.Parse(); err != nil {
		log.Fatalln(err)
	}

	m := make(map[int64]bool)
	event := make(map[int64][]eventT)

	for _, e := range c.Events {
		start := e.Start.Unix()
		end := e.End.Unix()

		if e.Description != "" {
			// https://github.com/apognu/gocal/pull/6
			e.Description = strings.Replace(e.Description, `\n`, "\n", -1)
			e.Description = strings.Replace(e.Description, `\N`, "\n", -1)
		}

		if argv.verbose > 1 {
			fmt.Printf("%+v\n", e)
		}

		if e.Start.UnixNano() >= argv.start.UnixNano() {
			event[start] = append(event[start], eventT{
				Event:  e,
				Epoch:  start,
				Diff:   start - argv.start.Unix(),
				Date:   e.Start.Local().Format(argv.dateFormat),
				State: "start",
			})
			m[start] = true
		}

		event[end] = append(event[end], eventT{
			Event:  e,
			Epoch:  e.End.Unix(),
			Diff:   e.End.Unix() - argv.start.Unix(),
			Date:   e.End.Local().Format(argv.dateFormat),
			State: "end",
		})
		m[end] = true
	}

	keys := toSortedArray(m)

	ev := waitEvent
	if !argv.wait {
		ev = writeEvent
	}

	if err := ev(argv, keys, event); err != nil {
		log.Fatalln(err)
	}
}

func toSortedArray(m map[int64]bool) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func waitfor(argv *argvT, seconds int64) {
	if argv.dryrun {
		fmt.Printf("sleep: %d\n", seconds)
		return
	}
	time.Sleep(time.Duration(seconds) * time.Second)
}

func waitEvent(argv *argvT, keys []int64, event map[int64][]eventT) error {
	if len(keys) == 0 {
		waitfor(argv, argv.waitMin)
		return nil
	}
	e := event[keys[0]][0]
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
	return formatEvent(format, event[keys[0]])
}

func writeEvent(argv *argvT, keys []int64, event map[int64][]eventT) error {
	format := formatStdout
	if argv.outputFormat != "" {
		format = argv.outputFormat
	}
	for _, k := range keys {
		if err := formatEvent(format, event[k]); err != nil {
			return err
		}
	}
	return nil
}

func formatEvent(format string, event []eventT) error {
	funcMap := template.FuncMap{
		"text": text,
	}
	tmpl, err := template.New("format").Funcs(funcMap).Parse(format)
	if err != nil {
		return err
	}

	stdout := bufio.NewWriter(os.Stdout)
	for _, e := range event {
		if err := tmpl.Execute(stdout, e); err != nil {
			return err
		}
		if err := stdout.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func text(s string) string {
	m := []reT{
		{
			match:   `(?i)(<b>|</b>)`,
			replace: "*",
		},
		{
			match:   `(?i)<br>`,
			replace: "\n",
		},
		{
			match:   `(?i)(<i>|</i>)`,
			replace: "_",
		},
		{
			match:   `(?i)(<pre>|</pre>)`,
			replace: "\n```\n",
		},
		{
			match:   `(?i)(<dl>|<ol>|<ul>|</dl>|</ol></ul>)`,
			replace: "\n\n",
		},
		{
			match:   `(?i)<li>`,
			replace: "* ",
		},
		{
			match:   `(?i)</li>`,
			replace: "\n",
		},
	}

	for _, r := range m {
		re := regexp.MustCompile(r.match)
		s = re.ReplaceAllLiteralString(s, r.replace)
	}

	return html.UnescapeString(s)
}
