package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alereyleyva/agent-guard/internal/audit"
	"github.com/alereyleyva/agent-guard/internal/normalize"
	"github.com/alereyleyva/agent-guard/internal/policy"
	"github.com/alereyleyva/agent-guard/internal/provider"
)

type Flow struct {
	provider provider.Provider
	policy   *policy.Engine
	logger   audit.Logger
	client   *http.Client
}

type Result struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	StreamBody io.ReadCloser
}

func NewFlow(p provider.Provider, pol *policy.Engine, logger audit.Logger) *Flow {
	return &Flow{
		provider: p,
		policy:   pol,
		logger:   logger,
		client:   &http.Client{},
	}
}

func (f *Flow) Process(ctx context.Context, req normalize.NormalizedRequest) (*Result, error) {
	traceID := generateTraceID()
	reqHash := f.hashRequest(req)
	f.logger.Emit(
		audit.NewEvent(traceID, audit.EventTypeLLMRequest).
			WithProvider(f.provider.Name()).
			WithModel(req.Model).
			WithHash(reqHash).
			WithStream(req.Stream),
	)

	modelDecision := f.policy.EvaluateModel(req.Model)
	f.logger.Emit(
		audit.NewEvent(traceID, audit.EventTypePolicyDecision).
			WithProvider(f.provider.Name()).
			WithModel(req.Model).
			WithDecision(modelDecision.Action, modelDecision.RuleID, modelDecision.Reason),
	)

	if !modelDecision.IsAllowed() {
		return nil, NewPolicyDeniedError(modelDecision.Reason)
	}

	if req.Stream {
		return f.processStreaming(ctx, traceID, req)
	}

	return f.processNonStreaming(ctx, traceID, req)
}

func (f *Flow) processNonStreaming(ctx context.Context, traceID string, req normalize.NormalizedRequest) (*Result, error) {
	upstreamReq, err := f.provider.BuildUpstreamRequest(req)
	if err != nil {
		return nil, fmt.Errorf("building upstream request: %w", err)
	}
	upstreamReq = upstreamReq.WithContext(ctx)

	resp, err := f.client.Do(upstreamReq)
	if err != nil {
		return nil, fmt.Errorf("sending upstream request: %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	header := cloneHeader(resp.Header)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading upstream response: %w", err)
	}

	normalizedResp, err := f.parseResponse(resp, body)
	if err != nil {
		normalizedResp = normalize.NormalizedResponse{RawBody: body, Model: req.Model}
	}
	normalizedResp.RawBody = body

	respHash := audit.HashContent(body)
	modelName := normalizedResp.Model
	if modelName == "" {
		modelName = req.Model
	}
	f.logger.Emit(
		audit.NewEvent(traceID, audit.EventTypeLLMResponse).
			WithProvider(f.provider.Name()).
			WithModel(modelName).
			WithHash(respHash),
	)

	for _, toolCall := range normalizedResp.ToolCalls {
		toolName := toolCall.Function.Name
		f.logger.Emit(
			audit.NewEvent(traceID, audit.EventTypeToolProposal).
				WithProvider(f.provider.Name()).
				WithModel(modelName).
				WithToolName(toolName),
		)

		toolDecision := f.policy.EvaluateTool(toolName)
		f.logger.Emit(
			audit.NewEvent(traceID, audit.EventTypePolicyDecision).
				WithProvider(f.provider.Name()).
				WithModel(modelName).
				WithToolName(toolName).
				WithDecision(toolDecision.Action, toolDecision.RuleID, toolDecision.Reason),
		)
	}

	return &Result{
		StatusCode: statusCode,
		Header:     header,
		Body:       body,
	}, nil
}

func (f *Flow) processStreaming(ctx context.Context, traceID string, req normalize.NormalizedRequest) (*Result, error) {
	upstreamReq, err := f.provider.BuildUpstreamRequest(req)
	if err != nil {
		return nil, fmt.Errorf("building upstream request: %w", err)
	}
	upstreamReq = upstreamReq.WithContext(ctx)

	resp, err := f.client.Do(upstreamReq)
	if err != nil {
		return nil, fmt.Errorf("sending upstream request: %w", err)
	}

	f.logger.Emit(
		audit.NewEvent(traceID, audit.EventTypeLLMResponse).
			WithProvider(f.provider.Name()).
			WithModel(req.Model).
			WithStream(true),
	)

	return &Result{
		StatusCode: resp.StatusCode,
		Header:     cloneHeader(resp.Header),
		StreamBody: resp.Body,
	}, nil
}

func (f *Flow) hashRequest(req normalize.NormalizedRequest) string {
	data, _ := json.Marshal(req)
	return audit.HashContent(data)
}

func (f *Flow) parseResponse(resp *http.Response, body []byte) (normalize.NormalizedResponse, error) {
	cloned := *resp
	cloned.Body = io.NopCloser(bytes.NewReader(body))
	return f.provider.ParseUpstreamResponse(&cloned)
}

func cloneHeader(h http.Header) http.Header {
	cloned := make(http.Header, len(h))
	for key, values := range h {
		copied := make([]string, len(values))
		copy(copied, values)
		cloned[key] = copied
	}
	return cloned
}
