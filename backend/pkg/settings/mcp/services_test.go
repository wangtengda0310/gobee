package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDefaultMcpServices(t *testing.T) {
	services, err := BuildDefaultMcpServices()
	require.NoError(t, err)
	require.NotNil(t, services)

	assert.NotNil(t, services.McpJsonCaseService)
	assert.NotNil(t, services.McpFuncCaseConfigService)
	assert.NotNil(t, services.McpExcelCheckService)
	assert.NotNil(t, services.McpExcelConfigService)
	assert.NotNil(t, services.McpGameExcelService)
	assert.NotNil(t, services.McpExcelTestGameService)
	assert.NotNil(t, services.McpHeroResCheckService)
	assert.NotNil(t, services.McpHeroWikiResCheckService)
	assert.NotNil(t, services.McpActivityWikiCheckService)
	assert.NotNil(t, services.McpConfigService)
	assert.NotNil(t, services.McpFeishuNotifyConfigService)
	assert.NotNil(t, services.McpRobotExtService)
}
