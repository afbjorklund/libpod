package events

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// EventLogFile is the structure for event writing to a logfile. It contains the eventer
// options and the event itself.  Methods for reading and writing are also defined from it.
type EventLogFile struct {
	options EventerOptions
}

// Writes to the log file
func (e EventLogFile) Write(ee Event) error {
	f, err := os.OpenFile(e.options.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	defer f.Close()
	eventJSONString, err := ee.ToJSONString()
	if err != nil {
		return err
	}
	if _, err := f.WriteString(fmt.Sprintf("%s\n", eventJSONString)); err != nil {
		return err
	}
	return nil

}

// Reads from the log file
func (e EventLogFile) Read(options ReadOptions) error {
	eventOptions, err := generateEventOptions(options.Filters, options.Since, options.Until)
	if err != nil {
		return errors.Wrapf(err, "unable to generate event options")
	}
	t, err := e.getTail(options)
	if err != nil {
		return err
	}
	for line := range t.Lines {
		event, err := newEventFromJSONString(line.Text)
		if err != nil {
			return err
		}
		switch event.Type {
		case Image, Volume, Pod, Container:
		//	no-op
		default:
			return errors.Errorf("event type %s is not valid in %s", event.Type.String(), e.options.LogFilePath)
		}
		include := true
		for _, filter := range eventOptions {
			include = include && filter(event)
		}
		if include {
			options.EventChannel <- event
		}
	}
	close(options.EventChannel)
	return nil
}
