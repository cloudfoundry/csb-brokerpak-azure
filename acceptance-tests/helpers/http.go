package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func HTTPGet(url string) string {
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

func HTTPPost(url, data string) {
	fmt.Fprintf(GinkgoWriter, "HTTP POST: %s\n", url)
	fmt.Fprintf(GinkgoWriter, "Sending data: %s\n", data)
	response, err := http.Post(url, "text/html", strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(SatisfyAny(HaveHTTPStatus(http.StatusCreated), HaveHTTPStatus(http.StatusOK)))
}

func HTTPPostJSON(url string, data interface{}) {
	payload, err := json.Marshal(data)
	Expect(err).NotTo(HaveOccurred())
	fmt.Fprintf(GinkgoWriter, "HTTP POST: %s\n", url)
	fmt.Fprintf(GinkgoWriter, "Sending JSON data: %s\n", string(payload))
	response, err := http.Post(url, "application/json", bytes.NewReader(payload))
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(SatisfyAny(HaveHTTPStatus(http.StatusCreated), HaveHTTPStatus(http.StatusOK)))
}

func HTTPPostFile(url string, fileContent []byte) {
	fmt.Fprintf(GinkgoWriter, "HTTP POST: %s\n", url)
	fmt.Fprintf(GinkgoWriter, "Sending data: %s\n", string(fileContent))

	response, err := http.Post(url, "multipart/form-data", bytes.NewReader(fileContent))
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated))
}

func HTTPPut(url, data string) {
	fmt.Fprintf(GinkgoWriter, "HTTP PUT: %s\n", url)
	fmt.Fprintf(GinkgoWriter, "Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	request.Header.Set("Content-Type", "text/html")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(SatisfyAny(HaveHTTPStatus(http.StatusCreated), HaveHTTPStatus(http.StatusOK)))
}

func HTTPDelete(url string) {
	fmt.Fprintf(GinkgoWriter, "HTTP DELETE: %s\n", url)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	Expect(err).NotTo(HaveOccurred())

	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusGone))
}
