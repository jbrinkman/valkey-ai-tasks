package mcp

// registerResources registers all resources with the MCP server
func (s *MCPGoServer) registerResources() {
	// Create and register the plan resource provider
	planResourceProvider := NewPlanResourceProvider(s.planRepo, s.taskRepo)
	planResourceProvider.RegisterResource(s)
}
