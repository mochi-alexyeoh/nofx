package provider

import (
	"net/http"

	"nofx/mcp"
)

const (
	DefaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"
	DefaultOpenRouterModel   = "openrouter/auto"
)

func init() {
	mcp.RegisterProvider(mcp.ProviderOpenRouter, func(opts ...mcp.ClientOption) mcp.AIClient {
		return NewOpenRouterClientWithOptions(opts...)
	})
}

type OpenRouterClient struct {
	*mcp.Client
}

func (c *OpenRouterClient) BaseClient() *mcp.Client { return c.Client }

func NewOpenRouterClient() mcp.AIClient {
	return NewOpenRouterClientWithOptions()
}

func NewOpenRouterClientWithOptions(opts ...mcp.ClientOption) mcp.AIClient {
	openrouterOpts := []mcp.ClientOption{
		mcp.WithProvider(mcp.ProviderOpenRouter),
		mcp.WithModel(DefaultOpenRouterModel),
		mcp.WithBaseURL(DefaultOpenRouterBaseURL),
	}

	allOpts := append(openrouterOpts, opts...)
	baseClient := mcp.NewClient(allOpts...).(*mcp.Client)

	client := &OpenRouterClient{Client: baseClient}
	baseClient.Hooks = client
	return client
}

func (c *OpenRouterClient) SetAPIKey(apiKey string, customURL string, customModel string) {
	c.APIKey = apiKey
	if customURL != "" {
		c.BaseURL = customURL
	}
	if customModel != "" {
		c.Model = customModel
	}
}

func (c *OpenRouterClient) SetAuthHeader(reqHeaders http.Header) {
	c.Client.SetAuthHeader(reqHeaders)
}
