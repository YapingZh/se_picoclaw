package agent

import (
    "context"
    "errors"
    "os/exec"

    "github.com/sipeed/picoclaw/pkg/logger"
)

type Text2audioHook struct{}

func (h *Text2audioHook) Name() string { return "text2audio" }

func (h *Text2audioHook) AfterLLM(ctx context.Context, resp *LLMHookResponse) (*LLMHookResponse, HookDecision, error) {
	if h == nil || resp == nil {
		return resp, HookDecision{Action: HookActionContinue}, errors.New("null object")
	}

        logger.WarnCF("text2audioHook", "hook after LLM", map[string]any{
			"name":       "text2audioHook",
			"text":      resp.Response.Content,
	})


	cmd := exec.Command("python3", "/usr/local/text2auto.py", resp.Response.Content)
    	_, err := cmd.CombinedOutput()
    	if err != nil {
		logger.WarnCF("text2audioHook", "hook after LLM", map[string]any{
                        "name":      "text2audioHook",
                        "text":      resp.Response.Content,
			"error":     err,
        	})
    	}

        return resp, HookDecision{Action: HookActionContinue}, errors.New("success")
}

func (h *Text2audioHook) BeforeLLM(ctx context.Context, req *LLMHookRequest) (*LLMHookRequest, HookDecision, error) {
	if h == nil || req == nil {
		return req, HookDecision{Action: HookActionContinue}, errors.New("null object")
	}


        logger.WarnCF("text2audioHook", "hook before LLM", map[string]any{
                        "name":       "text2audioHook",
        })

        return req, HookDecision{Action: HookActionContinue}, errors.New("success")
}

