package apps

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func (a *App) GET(path string) string {
	return a.GETf("%s", path)
}

func (a *App) GETf(format string, s ...any) string {
	url := a.urlf(format, s...)
	var data []byte

	Eventually(func(g Gomega) *http.Response {
		GinkgoWriter.Printf("HTTP GET: %s\n", url)
		response, err := http.Get(url)
		g.Expect(err).NotTo(HaveOccurred())

		GinkgoWriter.Printf("HTTP Status: %s\n", response.Status)

		defer response.Body.Close()
		data, err = io.ReadAll(response.Body)
		g.Expect(err).NotTo(HaveOccurred())

		GinkgoWriter.Printf("Recieved: %s\n", string(data))

		return response
	}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(HaveHTTPStatus(http.StatusOK))

	return string(data)
}

func (a *App) PUT(data, path string) {
	a.PUTf(data, "%s", path)
}

func (a *App) PUTf(data, format string, s ...any) {
	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP PUT: %s\n", url)
	GinkgoWriter.Printf("Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Set("Content-Type", "text/html")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated, http.StatusOK))
}

func (a *App) DELETE(path string) {
	a.DELETEf("%s", path)
}

func (a *App) DELETEf(format string, s ...any) {
	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP DELETE: %s\n", url)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	Expect(err).NotTo(HaveOccurred())

	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusGone, http.StatusNoContent))
}

func (a *App) urlf(format string, s ...any) string {
	base := a.URL
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
