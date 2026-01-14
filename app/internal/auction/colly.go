package auction

import (
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

	/*&TibiaCharCollyLogDebugger{Prefix: fmt.Sprintf("[%s] ", prefix)}*/
	c := colly.NewCollector(
		colly.AllowedDomains(constants.TibiaOfficialWebsite),
	)

	if cs.debugger != nil {
		c.SetDebugger(cs.debugger)
	}

	return cs
}

func WithDebugger(d debug.Debugger) Option {
	return func(c *CollyScrapper) {
		c.debugger = d
	}
}
