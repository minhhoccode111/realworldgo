package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"auth/internal/server"
)

var _ = Describe("Routes", func() {
	var ( 
		s *server.Server
		testServer *httptest.Server
	)

	BeforeEach(func() {
		s = &server.Server{}
		testServer = httptest.NewServer(http.HandlerFunc(s.HelloWorldHandler))
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("HelloWorldHandler", func() {
		Context("when a request is made to the hello world endpoint", func() {
			It("returns a 200 OK status and the expected message", func() {
				resp, err := http.Get(testServer.URL)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK), "Expected status OK")

				expected := `{"message":"Hello, World!"}`
				body, err := io.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				bodyString := strings.TrimSpace(string(body))
				Expect(bodyString).To(Equal(expected), "Expected response body to be %q", expected)
			})
		})
	})
})