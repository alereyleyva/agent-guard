package provider

import (
	"net/http"

	"github.com/alereyleyva/agent-guard/internal/normalize"
)

type Provider interface {
	Name() string
	BuildUpstreamRequest(req normalize.NormalizedRequest) (*http.Request, error)
	ParseUpstreamResponse(resp *http.Response) (normalize.NormalizedResponse, error)
}
