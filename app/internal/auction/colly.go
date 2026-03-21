package auction

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/gocolly/colly/v2"

	"github.com/gocolly/colly/v2/debug"
)

type TibiaCharCollyLogDebugger struct {
	// Output is the log destination, anything can be used which implements them
	// io.Writer interface. Leave it blank to use STDERR
	Output io.Writer

	// Prefix appears at the beginning of each generated log line
	Prefix string

	// Flag defines the logging properties.
	Flag    int
	logger  *log.Logger
	counter int32
	start   time.Time
}

func (l *TibiaCharCollyLogDebugger) Init() error {
	l.counter = 0

	l.start = time.Now()

	if l.Output == nil {
		l.Output = os.Stderr
	}

	l.logger = log.New(l.Output, l.Prefix, l.Flag|log.LstdFlags|log.LUTC)

	return nil
}

func (l *TibiaCharCollyLogDebugger) Event(e *debug.Event) {
	counter := atomic.AddInt32(&l.counter, 1)
	l.logger.Printf("[%06d] Id: %d [TraceId: %06d %s] %q (%s)\n", counter, e.CollectorID, e.RequestID, e.Type, e.Values, time.Since(l.start))
}

type CollyFactory struct{}

func (cf *CollyFactory) CreateScrapper(prefix string) *CollyScrapper {
	//TODO: this should work with a flag
	//return NewScrapper(WithDebugger(prefix))
	return NewScrapper()
}

type CollyScrapper struct {
	Collector *colly.Collector
	debugger  debug.Debugger
}

type Option func(*CollyScrapper)

func NewScrapper(opts ...Option) *CollyScrapper {
	cs := &CollyScrapper{}

	for _, opt := range opts {
		opt(cs)
	}

	headers := map[string]string{
		"accept":                      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language":             "en,es-ES;q=0.9,es;q=0.8",
		"cache-control":               "max-age=0",
		"priority":                    "u=0, i",
		"sec-ch-ua":                   `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`,
		"sec-ch-ua-arch":              `"x86"`,
		"sec-ch-ua-bitness":           `"64"`,
		"sec-ch-ua-full-version":      `"143.0.7499.192"`,
		"sec-ch-ua-full-version-list": `"Google Chrome";v="143.0.7499.192", "Chromium";v="143.0.7499.192", "Not A(Brand";v="24.0.0.0"`,
		"sec-ch-ua-mobile":            "?0",
		"sec-ch-ua-model":             `""`,
		"sec-ch-ua-platform":          `"Linux"`,
		"sec-ch-ua-platform-version":  `""`,
		"sec-fetch-dest":              "document",
		"sec-fetch-mode":              "navigate",
		"sec-fetch-site":              "cross-site",
		"sec-fetch-user":              "?1",
		"upgrade-insecure-requests":   "1",
	}

	c := colly.NewCollector(
		colly.AllowedDomains(constants.TibiaOfficialWebsite),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"),
		colly.Headers(headers),
	)

	if cs.debugger != nil {
		c.SetDebugger(cs.debugger)
	}

	cs.Collector = c

	return cs
}

func WithDebugger(prefix string) Option {
	return func(c *CollyScrapper) {
		c.debugger = &TibiaCharCollyLogDebugger{Prefix: fmt.Sprintf("[%s] ", prefix)}
	}
}
