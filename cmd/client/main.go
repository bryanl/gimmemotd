package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gizak/termui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	buckets     map[string][]time.Time
	mu          sync.RWMutex
	expireAfter = 30 * time.Second
)

func init() {
	buckets = make(map[string][]time.Time)
}

type response struct {
	Hostname string `json:"hostname,omitempty"`
	Message  string `json:"message,omitempty"`
}

func main() {
	urlStr := flag.String("url", "http://localhost:8181", "server address")
	flag.Parse()

	logger := logrus.New()

	if err := termui.Init(); err != nil {
		logger.WithError(err).Fatal("initialize term ui")
	}
	defer termui.Close()

	quit := make(chan bool, 0)
	go func() {
		retrieveT := time.NewTicker(time.Millisecond * 100)

		var done bool

		for {
			select {
			case <-quit:
				done = true
			case <-retrieveT.C:
				retrieve(logger, *urlStr)
			}

			if done {
				logger.Println("quitting")
				break
			}
		}
	}()

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
		quit <- true
	})

	termui.Handle("/sys/wnd/resize", func(e termui.Event) {
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})

	termui.Handle("/timer/1s", func(e termui.Event) {
		update()
	})

	termui.Loop()
}

func retrieve(logger logrus.FieldLogger, urlStr string) {
	mu.Lock()
	defer mu.Unlock()

	s, err := fetch(urlStr)
	if err != nil {
		return
	}

	if _, ok := buckets[s]; !ok {
		buckets[s] = make([]time.Time, 0)
	}

	now := time.Now()
	buckets[s] = append(buckets[s], now)
}

func update() {
	mu.Lock()
	defer mu.Unlock()

	cutoff := time.Now().Add(-1 * expireAfter)

	var total int
	for name, bucket := range buckets {
		if len(bucket) < 1 {
			delete(buckets, name)
			continue
		}

		for i := len(bucket) - 1; i >= 0; i-- {
			if !bucket[i].After(cutoff) {
				buckets[name] = bucket[i+1:]
				break
			}
		}

		total += len(buckets[name])
	}

	bc := termui.NewBarChart()
	bc.Width = 80
	bc.BorderLabel = fmt.Sprintf("Versions (seen in last %s)", expireAfter.String())
	bc.Height = termui.TermHeight()
	bc.TextColor = termui.ColorYellow
	bc.BarColor = termui.ColorBlue
	bc.NumColor = termui.ColorWhite
	bc.BarWidth = bc.Width / 5

	var labels []string
	var data []int

	var names []string
	for name := range buckets {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		bucket := buckets[name]

		labels = append(labels, name)
		data = append(data, len(bucket))
	}

	bc.DataLabels = labels
	bc.Data = data

	termui.Clear()
	termui.Render(bc)
}

func fetch(urlStr string) (string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "request")
	}

	defer resp.Body.Close()

	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return "", errors.Wrap(err, "decode body")
	}

	return r.Message, nil
}

func buildGauge(name string, percentage int) *termui.Gauge {
	g := termui.NewGauge()

	g.Width = 50
	g.Height = 3
	g.BorderLabel = name
	g.BarColor = termui.ColorRed
	g.BorderFg = termui.ColorWhite
	g.BorderLabelFg = termui.ColorCyan
	g.Percent = percentage

	return g
}
