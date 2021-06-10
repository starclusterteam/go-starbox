package metrics

import "context"

type Server interface {
	Run() error
	Stop(context.Context) error
}
