package cli_test

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"github.com/cf-platform-eng/kibosh/pkg/bazaar"
	"github.com/cf-platform-eng/kibosh/pkg/bazaar/cli"
	"github.com/cf-platform-eng/kibosh/pkg/httphelpers"
	"io/ioutil"
	"os"
)

var _ = Describe("Save charts", func() {
	var b bytes.Buffer
	var out *bufio.Writer

	var bazaarAPIRequest *http.Request
	var bazaarAPITestServer *httptest.Server
	var file *os.File

	BeforeEach(func() {
		b = bytes.Buffer{}
		out = bufio.NewWriter(&b)

		var err error
		file, err = ioutil.TempFile("", "foo")
		Expect(err).To(BeNil())
		_, err = file.Write([]byte("I am really a tgz of a chart"))
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		bazaarAPITestServer.Close()
		os.Remove(file.Name())
	})

	It("calls save charts", func() {
		responseBody, _ := json.Marshal(bazaar.DisplayResponse{
			Message: "Yay",
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(responseBody)
			r.ParseMultipartForm(4096)
			bazaarAPIRequest = r
		})
		bazaarAPITestServer = httptest.NewServer(handler)

		c := cli.NewChartsSaveCmd(out)
		c.Flags().Set("target", bazaarAPITestServer.URL)
		c.Flags().Set("user", "bob")
		c.Flags().Set("password", "monkey123")

		err := c.RunE(c, []string{
			file.Name(),
		})

		out.Flush()

		Expect(err).To(BeNil())
		Expect(string(b.Bytes())).To(ContainSubstring("Yay"))

		Expect(bazaarAPIRequest.Method).To(Equal("POST"))

		formFile, _, err := bazaarAPIRequest.FormFile("chart")
		Expect(err).To(BeNil())

		fileContents, err := ioutil.ReadAll(formFile)
		Expect(err).To(BeNil())

		Expect(fileContents).To(Equal([]byte("I am really a tgz of a chart")))
	})

	It("correctly auths request", func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("{}"))
			bazaarAPIRequest = r
		})
		bazaarAPITestServer = httptest.NewServer(handler)

		c := cli.NewChartsSaveCmd(out)
		c.Flags().Set("target", bazaarAPITestServer.URL)
		c.Flags().Set("user", "bob")
		c.Flags().Set("password", "monkey123")

		err := c.RunE(c, []string{
			file.Name(),
		})
		out.Flush()

		Expect(err).To(BeNil())
		Expect(bazaarAPIRequest.Header.Get("Authorization")).To(
			Equal(httphelpers.BasicAuthHeaderVal("bob", "monkey123")),
		)
	})

	It("auth failure  save charts", func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		})
		bazaarAPITestServer = httptest.NewServer(handler)

		c := cli.NewChartsSaveCmd(out)
		c.Flags().Set("target", bazaarAPITestServer.URL)
		c.Flags().Set("user", "bob")
		c.Flags().Set("password", "monkey123")

		err := c.RunE(c, []string{
			file.Name(),
		})

		out.Flush()

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("401"))
	})
})
