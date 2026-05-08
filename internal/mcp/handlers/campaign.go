package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// CampaignHandlers handles campaign-related MCP tools
type CampaignHandlers struct {
	service *services.CampaignService
}

// NewCampaignHandlers creates new campaign handlers
func NewCampaignHandlers(service *services.CampaignService) *CampaignHandlers {
	return &CampaignHandlers{service: service}
}

// HandleCreateCampaign handles the create_campaign tool
func (h *CampaignHandlers) HandleCreateCampaign() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		name := getStringArg(args, "name")
		title := getStringArg(args, "title")
		setting := getStringArg(args, "setting")

		if name == "" {
			return mcp.NewToolResultError("campaign name is required"), nil
		}

		campaign, err := h.service.CreateCampaign(name, title, setting)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Campaign '%s' created successfully", campaign.Name)), nil
	}
}

// HandleSaveAct handles the save_act tool
func (h *CampaignHandlers) HandleSaveAct() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		actNum := getIntArg(args, "act_number")
		title := getStringArg(args, "title")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}
		if actNum <= 0 {
			return mcp.NewToolResultError("act_number must be a positive integer"), nil
		}
		if title == "" {
			return mcp.NewToolResultError("title is required"), nil
		}

		if err := h.service.SaveAct(campaign, actNum, title, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Act %d '%s' saved to campaign '%s'", actNum, title, campaign)), nil
	}
}

// HandleCompilePDF handles the compile_pdf tool
func (h *CampaignHandlers) HandleCompilePDF() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		title := getStringArg(args, "title")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
		defer cancel()

		pdfPath, err := h.service.CompilePDF(ctx, campaign, title)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("PDF compiled: %s", pdfPath)), nil
	}
}

// HandleSaveLore handles the save_lore tool
func (h *CampaignHandlers) HandleSaveLore() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		if err := h.service.SaveLore(campaign, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Lore saved to campaign '%s'", campaign)), nil
	}
}

// HandleSaveNPCs handles the save_npcs tool
func (h *CampaignHandlers) HandleSaveNPCs() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		if err := h.service.SaveNPCs(campaign, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("NPCs saved to campaign '%s'", campaign)), nil
	}
}

// HandleSaveEncounters handles the save_encounters tool
func (h *CampaignHandlers) HandleSaveEncounters() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		if err := h.service.SaveEncounters(campaign, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Encounters saved to campaign '%s'", campaign)), nil
	}
}

// HandleSaveBestiary handles the save_bestiary tool
func (h *CampaignHandlers) HandleSaveBestiary() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		if err := h.service.SaveBestiary(campaign, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Bestiary saved to campaign '%s'", campaign)), nil
	}
}

// HandleSaveMaps handles the save_maps tool
func (h *CampaignHandlers) HandleSaveMaps() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		if err := h.service.SaveMaps(campaign, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Maps saved to campaign '%s'", campaign)), nil
	}
}

// HandleSaveIntroduction handles the save_introduction tool
func (h *CampaignHandlers) HandleSaveIntroduction() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		if err := h.service.SaveIntroduction(campaign, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Introduction saved to campaign '%s'", campaign)), nil
	}
}

// HandleSaveSettingGuide handles the save_setting_guide tool
func (h *CampaignHandlers) HandleSaveSettingGuide() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		if err := h.service.SaveSettingGuide(campaign, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Setting guide saved to campaign '%s'", campaign)), nil
	}
}

// HandleSaveAppendices handles the save_appendices tool
func (h *CampaignHandlers) HandleSaveAppendices() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		if err := h.service.SaveAppendices(campaign, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Appendices saved to campaign '%s'", campaign)), nil
	}
}

// HandleGetTemplate handles the get_template tool
func (h *CampaignHandlers) HandleGetTemplate() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		tmplType := getStringArg(args, "type")
		if tmplType == "" {
			return mcp.NewToolResultError("type is required"), nil
		}

		template, err := h.service.GetTemplate(tmplType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(template), nil
	}
}
