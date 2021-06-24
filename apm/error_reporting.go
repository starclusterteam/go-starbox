package apm

import (
	"context"

	"github.com/starclusterteam/go-starbox/log"
)

var GlobalReporter ErrorReporting

func init() {
	GlobalReporter = NewNullReporter()
}

type ErrorReporting interface {
	Report(ctx context.Context, err error) error
	ReportAsync(ctx context.Context, err error)
}

type NullReporter struct{}

func NewNullReporter() ErrorReporting {
	return &NullReporter{}
}

func (c *NullReporter) Report(ctx context.Context, err error) error {
	return nil
}

func (c *NullReporter) ReportAsync(ctx context.Context, err error) {}

type ConsoleReporter struct{}

func NewConsoleReporter() ErrorReporting {
	return &ConsoleReporter{}
}

func (c *ConsoleReporter) Report(ctx context.Context, err error) error {
	log.Errorf("received error: %v", err)
	return nil
}

func (c *ConsoleReporter) ReportAsync(ctx context.Context, err error) {
	log.Errorf("received error: %v", err)
}

func SetNullReporter() {
	GlobalReporter = NewNullReporter()
}

func SetConsoleReporter() {
	GlobalReporter = NewConsoleReporter()
}
