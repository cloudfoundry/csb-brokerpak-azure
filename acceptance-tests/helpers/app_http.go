package helpers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func (a AppInstance) GET(format string, s ...interface{}) string {
	url := a.url(format, s...)
	fmt.Fprintf(GinkgoWriter, "HTTP GET: %s\n", url)
	response, err := http.Get(url)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusOK))

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	Expect(err).NotTo(HaveOccurred())

	fmt.Fprintf(GinkgoWriter, "Recieved: %s\n", string(data))
	return string(data)
}

func (a AppInstance) PUT(data, format string, s ...interface{}) {
	url := a.url(format, s...)
	fmt.Fprintf(GinkgoWriter, "HTTP PUT: %s\n", url)
	fmt.Fprintf(GinkgoWriter, "Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	request.Header.Set("Content-Type", "text/html")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated, http.StatusOK))
}

func (a AppInstance) DELETE(format string, s ...interface{}) {
	url := a.url(format, s...)
	fmt.Fprintf(GinkgoWriter, "HTTP DELETE: %s\n", url)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	Expect(err).NotTo(HaveOccurred())

	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusGone, http.StatusNoContent))
}

func (a AppInstance) url(format string, s ...interface{}) string {
	base := fmt.Sprintf("http://%s.%s", a.name, DefaultSharedDomain())
	path := fmt.Sprintf(format, s...)
	switch {
	case len(path) == 0:
		return base
	case path[0] != '/':
		return fmt.Sprintf("%s/%s", base, path)
	default:
		return base + path
	}
}
