package web_test

import (
	"fmt"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/starclusterteam/go-starbox/web"
)

var _ = Describe("URL Signature", func() {
	var signTestKey = "sign key"
	var createSignTests = []struct {
		u        *url.URL // input
		expected string   // expected result
	}{
		{mustParseURL("http://proxy/?height=100&width=100"), "f526780761875227ec4d576ad03054bbe7d24666"},
		{mustParseURL("http://proxy/?height=100&width=100&s=fake+signature"), "f526780761875227ec4d576ad03054bbe7d24666"},
		{mustParseURL("http://proxy/?height=100&width=100&url=http%3A%2F%2Fgoogle.com%2F"), "13567c785c6e846df12c72835406b10ce80f3c5e"},
		{mustParseURL("http://proxy/?height=100&url=http%3A%2F%2Fgoogle.com%2F&width=100"), "13567c785c6e846df12c72835406b10ce80f3c5e"},
	}

	var validSignTests = []struct {
		u        *url.URL // input
		expected bool     // expected result
	}{
		{mustParseURL("http://proxy/?height=100&width=100&s=f526780761875227ec4d576ad03054bbe7d24666"), true},
		{mustParseURL("http://proxy/?height=100&width=100&s=fake+signature"), false},
		{mustParseURL("http://proxy/?height=100&width=100&url=http%3A%2F%2Fgoogle.com%2F&s=13567c785c6e846df12c72835406b10ce80f3c5e"), true},
		{mustParseURL("http://proxy/?height=374&url=https://google.com/*/img&width=640&s=7a3bc1512eda5a7d9e23df0470cb1dd798851bb9"), true},
		{mustParseURL("http://proxy/?height=374&url=https%3A%2F%2Fgoogle.com%2F*%2Fimg&width=640&s=30b21c7278b18db65e7409c9020f1eb2f769b43c"), true},
		{mustParseURL("http://proxy/?url=https%3A%2F%2Fgoogle.com%2F*%2Fimg&width=640&s=30b21c7278b18db65e7409c9020f1eb2f769b43c&height=374&"), true},
	}

	Describe("CreateSign", func() {
		It("should create sign", func() {
			for _, tt := range createSignTests {
				Expect(web.CreateSign(signTestKey, tt.u)).To(Equal(tt.expected))
			}
		})
	})

	Describe("SignURL", func() {
		It("should sign url", func() {
			for _, tt := range createSignTests {
				// shallow copy
				var u = &url.URL{}
				*u = *tt.u

				web.SignURL(signTestKey, u)

				Expect(u.Query().Get("s")).To(Equal(tt.expected))
			}
		})
	})

	Describe("ValidSign", func() {
		It("should validate signature", func() {
			for _, tt := range validSignTests {
				Expect(web.ValidSign(signTestKey, tt.u)).To(Equal(tt.expected))
			}
		})
	})

})

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(fmt.Sprintf("failed to parse url(%s): %v", rawurl, err))
	}

	return u
}
