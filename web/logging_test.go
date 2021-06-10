package web

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	"github.com/starclusterteam/go-starbox/mock"
)

var _ = Describe("Logging", func() {
	Describe("logger", func() {
		var mockLogger *mock.MockInterface

		BeforeEach(func() {
			mockLogger = mock.NewMockInterface(mockCtrl)
		})

		It("behaves correctly", func() {
			fakeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("really"))
			})

			mockLoggerInjector := func(h http.Handler, logger *mock.MockInterface) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					SetLogger(r, logger)
					h.ServeHTTP(w, r)
				})
			}

			responseRecorder := httptest.NewRecorder()
			loggingHandler := mockLoggerInjector(logger(fakeHandler), mockLogger)

			mockLogger.EXPECT().With("method", "GET").Return(mockLogger)
			mockLogger.EXPECT().With("url", "http://example.com/localhost/").Return(mockLogger)
			mockLogger.EXPECT().With("latency", gomock.Any()).Return(mockLogger)
			mockLogger.EXPECT().With("status", http.StatusTeapot).Return(mockLogger)
			mockLogger.EXPECT().Info("request")

			loggingHandler.ServeHTTP(responseRecorder, httptest.NewRequest("GET", "/localhost/", nil))

			By("passing everything to true writer")
			Expect(responseRecorder.Code).To(Equal(http.StatusTeapot))
			Expect(ioutil.ReadAll(responseRecorder.Result().Body)).To(Equal([]byte("really")))

			By("logging context data")
		})
	})
})
