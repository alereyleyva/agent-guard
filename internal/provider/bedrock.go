package provider

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alereyleyva/agent-guard/internal/config"
	"github.com/alereyleyva/agent-guard/internal/normalize"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const bedrockServiceName = "bedrock"

type BedrockProvider struct {
	region   string
	endpoint string
	signer   *v4.Signer
	creds    aws.CredentialsProvider
}

func NewBedrock(region, endpoint, accessKeyID, secretAccessKey, sessionToken string) (*BedrockProvider, error) {
	if region == "" {
		return nil, fmt.Errorf("bedrock region is required")
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}

	if accessKeyID != "" || secretAccessKey != "" || sessionToken != "" {
		if accessKeyID == "" || secretAccessKey == "" {
			return nil, fmt.Errorf("bedrock access_key_id and secret_access_key must both be set")
		}
		creds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, sessionToken)
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(creds))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	if endpoint == "" {
		endpoint = fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
	}

	return &BedrockProvider{
		region:   region,
		endpoint: strings.TrimSuffix(endpoint, "/"),
		signer:   v4.NewSigner(),
		creds:    awsCfg.Credentials,
	}, nil
}

func init() {
	RegisterFactory("bedrock", bedrockFactory)
}

func bedrockFactory(cfg config.ProviderConfig) (Provider, error) {
	bedrockCfg := cfg.Bedrock
	return NewBedrock(bedrockCfg.Region, bedrockCfg.Endpoint, bedrockCfg.AccessKeyID, bedrockCfg.SecretAccessKey, bedrockCfg.SessionToken)
}

func (p *BedrockProvider) Name() string {
	return "bedrock"
}

func (p *BedrockProvider) BuildUpstreamRequest(req normalize.NormalizedRequest) (*http.Request, error) {
	bedrockReq := buildBedrockConverseRequest(req)
	body, err := json.Marshal(bedrockReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	path := "converse"
	if req.Stream {
		path = "converse-stream"
	}
	modelID := url.PathEscape(req.Model)
	url := fmt.Sprintf("%s/model/%s/%s", p.endpoint, modelID, path)

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if req.Stream {
		httpReq.Header.Set("Accept", "application/vnd.amazon.eventstream")
	} else {
		httpReq.Header.Set("Accept", "application/json")
	}

	payloadHash := hashSHA256(body)
	httpReq.Header.Set("X-Amz-Content-Sha256", payloadHash)

	creds, err := p.creds.Retrieve(context.Background())
	if err != nil {
		return nil, fmt.Errorf("retrieving aws credentials: %w", err)
	}
	if err := p.signer.SignHTTP(context.Background(), creds, httpReq, payloadHash, bedrockServiceName, p.region, time.Now()); err != nil {
		return nil, fmt.Errorf("signing request: %w", err)
	}

	return httpReq, nil
}

func (p *BedrockProvider) ParseUpstreamResponse(resp *http.Response) (normalize.NormalizedResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return normalize.NormalizedResponse{}, fmt.Errorf("reading response body: %w", err)
	}

	normalized, err := parseBedrockConverseResponse(body)
	if err != nil {
		return normalize.NormalizedResponse{RawBody: body}, err
	}

	return normalized, nil
}

func buildBedrockConverseRequest(req normalize.NormalizedRequest) bedrockConverseRequest {
	messages := make([]bedrockMessage, 0, len(req.Messages))
	system := make([]bedrockContentBlock, 0)

	for i, msg := range req.Messages {
		role := strings.ToLower(msg.Role)
		if role == "system" {
			if msg.Content != "" {
				content := msg.Content
				system = append(system, bedrockContentBlock{Text: &content})
			}
			continue
		}

		contentBlocks := make([]bedrockContentBlock, 0)
		if msg.Content != "" {
			content := msg.Content
			contentBlocks = append(contentBlocks, bedrockContentBlock{Text: &content})
		}

		if role == "assistant" {
			for j, toolCall := range msg.ToolCalls {
				toolID := toolCall.ID
				if toolID == "" {
					toolID = fmt.Sprintf("toolcall-%d-%d", i, j)
				}

				input := parseToolArguments(toolCall.Function.Arguments)
				contentBlocks = append(contentBlocks, bedrockContentBlock{
					ToolUse: &bedrockToolUse{
						ToolUseID: toolID,
						Name:      toolCall.Function.Name,
						Input:     input,
					},
				})
			}
		}

		if role == "tool" {
			toolID := msg.ToolCallID
			if toolID == "" {
				toolID = fmt.Sprintf("toolcall-%d", i)
			}
			toolContent := make([]bedrockContentBlock, 0)
			if msg.Content != "" {
				content := msg.Content
				toolContent = append(toolContent, bedrockContentBlock{Text: &content})
			}
			contentBlocks = []bedrockContentBlock{
				{
					ToolResult: &bedrockToolResult{
						ToolUseID: toolID,
						Content:   toolContent,
					},
				},
			}
			role = "user"
		}

		if len(contentBlocks) > 0 {
			messages = append(messages, bedrockMessage{Role: role, Content: contentBlocks})
		}
	}

	var toolConfig *bedrockToolConfig
	if len(req.Tools) > 0 {
		tools := make([]bedrockTool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			if tool.Type != "function" {
				continue
			}
			schema := tool.Function.Parameters
			if schema == nil {
				schema = map[string]interface{}{"type": "object"}
			}
			tools = append(tools, bedrockTool{
				ToolSpec: bedrockToolSpec{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					InputSchema: bedrockToolInputSchema{JSON: schema},
				},
			})
		}
		if len(tools) > 0 {
			toolConfig = &bedrockToolConfig{Tools: tools}
		}
	}

	return bedrockConverseRequest{
		Messages:   messages,
		System:     system,
		ToolConfig: toolConfig,
	}
}

func parseBedrockConverseResponse(body []byte) (normalize.NormalizedResponse, error) {
	var resp bedrockConverseResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return normalize.NormalizedResponse{RawBody: body}, fmt.Errorf("parsing response: %w", err)
	}

	message := resp.Output.Message
	if message.Role == "" && len(message.Content) == 0 {
		return normalize.NormalizedResponse{RawBody: body}, fmt.Errorf("missing output message")
	}

	normalized := normalize.NormalizedResponse{RawBody: body}
	var contentBuilder strings.Builder

	for _, block := range message.Content {
		if block.Text != nil {
			contentBuilder.WriteString(*block.Text)
		}
		if block.ToolUse != nil {
			args := "{}"
			if block.ToolUse.Input != nil {
				if data, err := json.Marshal(block.ToolUse.Input); err == nil {
					args = string(data)
				}
			}
			normalized.ToolCalls = append(normalized.ToolCalls, normalize.ToolCall{
				ID:   block.ToolUse.ToolUseID,
				Type: "function",
				Function: normalize.FunctionCall{
					Name:      block.ToolUse.Name,
					Arguments: args,
				},
			})
		}
	}

	normalized.Content = contentBuilder.String()
	return normalized, nil
}

func parseToolArguments(args string) interface{} {
	if args == "" {
		return map[string]interface{}{}
	}

	var decoded interface{}
	if err := json.Unmarshal([]byte(args), &decoded); err != nil {
		return map[string]interface{}{"raw": args}
	}
	return decoded
}

func hashSHA256(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
