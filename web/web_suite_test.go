package web

import (
	"testing"

	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	mockCtrl *gomock.Controller
)

var _ = BeforeSuite(func() {
	mockCtrl = gomock.NewController(GinkgoT())
})

var _ = AfterSuite(func() {
	mockCtrl.Finish()
})

func TestWeb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "web")
}
