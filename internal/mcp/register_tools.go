package mcp

// registerTools registers all the task management tools with the MCP server
func (s *MCPGoServer) registerTools() {
	// Plan tools
	s.registerPlanTools()

	// Task tools
	s.registerTaskTools()

	// Notes tools
	s.registerNotesTools()
}
