package updater

import (
	"io"
	"net/http"
)

func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func closeFile(c io.Closer) {
	if c != nil {
		_ = c.Close()
	}
}
