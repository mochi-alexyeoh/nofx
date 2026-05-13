package backtest

import (
	"fmt"
	"strings"

	"nofx/mcp"
	mcpprovider "nofx/mcp/provider"
	mcppayment "nofx/mcp/payment"
)

// configureMCPClient creates/clones an MCP client based on configuration (returns mcp.AIClient interface).
// Note: mcp.New() returns an interface type; here we convert to concrete implementation before copying to avoid concurrent shared state.
func configureMCPClient(cfg BacktestConfig, base mcp.AIClient) (mcp.AIClient, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.AICfg.Provider))

	// DeepSeek
	if provider == "" || provider == "inherit" || provider == "default" {
		client := cloneBaseClient(base)
		if cfg.AICfg.APIKey != "" || cfg.AICfg.BaseURL != "" || cfg.AICfg.Model != "" {
			client.SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		}
		return client, nil
	}

	switch provider {
	case "deepseek":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("deepseek provider requires api key")
		}
		ds := mcpprovider.NewDeepSeekClientWithOptions()
		ds.(*mcpprovider.DeepSeekClient).SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return ds, nil
	case "qwen":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("qwen provider requires api key")
		}
		qc := mcpprovider.NewQwenClientWithOptions()
		qc.(*mcpprovider.QwenClient).SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return qc, nil
	case "claude":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("claude provider requires api key")
		}
		cc := mcpprovider.NewClaudeClientWithOptions()
		cc.(*mcpprovider.ClaudeClient).SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return cc, nil
	case "kimi":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("kimi provider requires api key")
		}
		kc := mcpprovider.NewKimiClientWithOptions()
		kc.(*mcpprovider.KimiClient).SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return kc, nil
	case "gemini":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("gemini provider requires api key")
		}
		gc := mcpprovider.NewGeminiClientWithOptions()
		gc.(*mcpprovider.GeminiClient).SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return gc, nil
	case "grok":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("grok provider requires api key")
		}
		grokC := mcpprovider.NewGrokClientWithOptions()
		grokC.(*mcpprovider.GrokClient).SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return grokC, nil
	case "openai":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("openai provider requires api key")
		}
		oaiC := mcpprovider.NewOpenAIClientWithOptions()
		oaiC.(*mcpprovider.OpenAIClient).SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return oaiC, nil
	case "minimax":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("minimax provider requires api key")
		}
		mmC := mcpprovider.NewMiniMaxClientWithOptions()
		mmC.(*mcpprovider.MiniMaxClient).SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return mmC, nil
	case "blockrun-base", "blockrun-sol":
		return nil, fmt.Errorf("%s provider is not available in this NOFX build", provider)
	case "claw402":
		if cfg.AICfg.APIKey == "" {
			return nil, fmt.Errorf("claw402 provider requires wallet private key")
		}
		claw := mcppayment.NewClaw402ClientWithOptions()
		claw.(*mcppayment.Claw402Client).SetAPIKey(cfg.AICfg.APIKey, "", cfg.AICfg.Model)
		return claw, nil
	case "custom":
		if cfg.AICfg.BaseURL == "" || cfg.AICfg.APIKey == "" || cfg.AICfg.Model == "" {
			return nil, fmt.Errorf("custom provider requires base_url, api key and model")
		}
		client := cloneBaseClient(base)
		client.SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported ai provider %s", cfg.AICfg.Provider)
	}
}

// cloneBaseClient copies the base client to avoid shared mutable state.
func cloneBaseClient(base mcp.AIClient) *mcp.Client {
	if c, ok := base.(*mcp.Client); ok && c != nil {
		cp := *c
		return &cp
	}
	return mcp.NewClient().(*mcp.Client)
}
