package id

import (
	"fmt"

	"github.com/segmentio/ksuid"
)

func Generate(prefix string) string {
	uid := ksuid.New()
	return fmt.Sprintf("%s_%s", prefix, uid.String())
}

