package web_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/starclusterteam/go-starbox/log"
	"github.com/starclusterteam/go-starbox/web"
)

var _ = Describe("Context", func() {
	Describe("SetLogger and GetLogger", func() {
		var logger = log.Logger().With("key", "value")
		It("should set and get logger", func() {
			r := &http.Request{}
			newR := web.SetLogger(r, logger)

			l := web.GetLogger(newR)
			Expect(l).To(Equal(logger))

			l = web.GetLogger(r)
			Expect(l).To(Equal(logger))
		})
	})
})
