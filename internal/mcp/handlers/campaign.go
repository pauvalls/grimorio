package handlers

import (
	"context"
	"fmt"

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

		pdfPath, err := h.service.CompilePDF(campaign, title)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("PDF compiled: %s", pdfPath)), nil
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
