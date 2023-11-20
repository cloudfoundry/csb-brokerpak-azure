package brokerpaks

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func newClient() *client {
	return &client{
		token: os.Getenv("GITHUB_TOKEN"),
	}
}

// client is a microscopic GitHub client allowing HTTP GET
type client struct {
	token string
}

// get will do a HTTP GET to a body
func (c client) get(path, mimeType string) io.ReadCloser {
	req := must(http.NewRequest(http.MethodGet, path, nil))

	req.Header.Add("Accept", mimeType)
	if c.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	}

	res := must(http.DefaultClient.Do(req))
	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("expected HTTP 200 but got %d: %s", res.StatusCode, res.Status))
	}
	return res.Body
}

// download will do an HTTP GET to a file
func (c client) download(target, uri string) {
	fh := must(os.Create(target))
	defer fh.Close()

	body := c.get(uri, "application/octet-stream")
	defer body.Close()

	_, err := io.Copy(fh, body)
	if err != nil {
		panic(err)
	}
}

func must[A any](input A, err error) A {
	if err != nil {
		panic(err)
	}

	return input
}
