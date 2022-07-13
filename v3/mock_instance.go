package testerator

import (
	"io"
	"net/http"

	"google.golang.org/appengine/v2/aetest"
)

var _ aetest.Instance = (*mockAEInstance)(nil)

type mockAEInstance struct {
}

func (m mockAEInstance) Close() error {
	return nil
}

func (m mockAEInstance) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, urlStr, body)
}
