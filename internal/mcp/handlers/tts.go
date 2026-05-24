package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/dm"
	"github.com/pauvalls/grimorio/internal/services"
)

// TTSHandlers handles TTS-related MCP tools.
type TTSHandlers struct {
	ttsService *services.TTSService
}

// NewTTSHandlers creates new TTS handlers.
func NewTTSHandlers(service *services.TTSService) *TTSHandlers {
	return &TTSHandlers{ttsService: service}
}

// HandleSetDMMode handles the set_dm_mode tool.
func (h *TTSHandlers) HandleSetDMMode() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		modeStr := getStringArg(args, "mode")
		if modeStr == "" {
			return mcp.NewToolResultError("mode is required"), nil
		}

		mode := dm.DMMode(modeStr)
		if err := h.ttsService.SetMode(mode); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := map[string]any{
			"mode":        string(mode),
			"tts_enabled": h.ttsService.IsAvailable(),
		}
		jsonBytes, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleAssignNPCVoice handles the assign_npc_voice tool.
func (h *TTSHandlers) HandleAssignNPCVoice() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		npcName := getStringArg(args, "npc_name")
		voicePrompt := getStringArg(args, "voice_prompt")

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}
		if npcName == "" {
			return mcp.NewToolResultError("npc_name is required"), nil
		}
		if voicePrompt == "" {
			return mcp.NewToolResultError("voice_prompt is required"), nil
		}

		voiceID, created, err := h.ttsService.AssignNPCVoice(campaignID, npcName, voicePrompt)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := map[string]any{
			"voice_id":    voiceID,
			"created":     created,
			"tts_enabled": h.ttsService.IsAvailable(),
		}
		jsonBytes, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleTTSControl handles the tts_control tool.
func (h *TTSHandlers) HandleTTSControl() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		action := getStringArg(args, "action")
		if action == "" {
			return mcp.NewToolResultError("action is required"), nil
		}

		if err := h.ttsService.Control(action); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("TTS control action '%s' executed", action)), nil
	}
}

// HandleListTTSVoices handles the list_tts_voices tool.
func (h *TTSHandlers) HandleListTTSVoices() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		voices := h.ttsService.ListVoices(campaignID)
		jsonBytes, err := json.Marshal(voices)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal voices: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleGetTTSStatus handles the get_tts_status tool.
func (h *TTSHandlers) HandleGetTTSStatus() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status := h.ttsService.GetStatus()
		jsonBytes, err := json.Marshal(status)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal status: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
