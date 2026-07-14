/**
 * 设置页面状态管理模块
 *
 * 管理飞书通知和 MCP 服务配置，提供全局设置的加载、保存和状态切换功能。
 * 飞书通知和 MCP 配置是应用级别的配置，不与战斗测试特定配置耦合。
 */
import { ref } from "vue";
import { MCPConfigService } from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings";
import { ExcelConfigService } from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings";
import { FuncCaseConfigService } from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";

// ==================== 飞书配置 ====================
export const FeiShuNtf = ref(false);
export const FeiShuGuid = ref("36732a0b-9b65-4456-8294-17044223114f");

// ==================== MCP 配置 ====================
export const MCPEnabled = ref(true);
export const MCPPort = ref(8765);
export const MCPHost = ref("127.0.0.1");
export const MCPRunning = ref(false);
export const MCPLoading = ref(false);

// ==================== 统一策划配表目录配置 ====================
export const ExcelDir = ref("../../config");

// ==================== 初始化加载 ====================

/**
 * 加载所有全局设置（页面挂载时调用）
 */
export const loadAllSettings = async () => {
    await Promise.all([
        loadFeishuConfig(),
        loadMCPConfig(),
        loadExcelConfig()
    ]);
};

/**
 * 加载飞书配置
 */
const loadFeishuConfig = async () => {
    try {
        const config = await FuncCaseConfigService.GetConfig();
        if (config) {
            FeiShuNtf.value = config.fei_shu_ntf;
            FeiShuGuid.value = config.fei_shu_guid;
        }
    } catch (err) {
        console.error("加载飞书配置失败:", err);
    }
};

/**
 * 保存飞书配置
 * 使用 UpdateConfig 进行部分更新，避免覆盖其他配置项
 */
export const saveFeishuConfig = async () => {
    try {
        await FuncCaseConfigService.UpdateConfig({
            fei_shu_ntf: FeiShuNtf.value,
            fei_shu_guid: FeiShuGuid.value
        });
    } catch (err) {
        console.error("保存飞书配置失败:", err);
    }
};

// ==================== MCP 配置相关 ====================

/**
 * 加载 MCP 配置
 */
export const loadMCPConfig = async () => {
    MCPLoading.value = true;
    try {
        const config = await MCPConfigService.GetConfig();
        if (config) {
            MCPEnabled.value = config.enabled;
            MCPPort.value = config.port;
            MCPHost.value = config.host;
        }
        MCPRunning.value = await MCPConfigService.IsRunning();
    } catch (err) {
        console.error("加载 MCP 配置失败:", err);
    } finally {
        MCPLoading.value = false;
    }
};

/**
 * 保存 MCP 配置并重启服务
 */
export const saveMCPConfig = async () => {
    MCPLoading.value = true;
    try {
        await MCPConfigService.SaveConfig({
            enabled: MCPEnabled.value,
            port: MCPPort.value,
            host: MCPHost.value
        });
        // 保存后更新运行状态
        MCPRunning.value = await MCPConfigService.IsRunning();
    } catch (err) {
        console.error("保存 MCP 配置失败:", err);
    } finally {
        MCPLoading.value = false;
    }
};

/**
 * 切换 MCP 启用状态（会自动重启服务）
 */
export const toggleMCPEnabled = async (enabled: boolean) => {
    MCPLoading.value = true;
    try {
        await MCPConfigService.SaveConfig({
            enabled: enabled,
            port: MCPPort.value,
            host: MCPHost.value
        });
        // 更新运行状态
        MCPRunning.value = await MCPConfigService.IsRunning();
    } catch (err) {
        console.error("切换 MCP 服务失败:", err);
        // 恢复原状态
        MCPEnabled.value = !enabled;
    } finally {
        MCPLoading.value = false;
    }
};

// ==================== 统一策划配表目录配置相关 ====================

/**
 * 加载统一策划配表目录配置
 */
const loadExcelConfig = async () => {
    try {
        const config = await ExcelConfigService.GetConfig();
        if (config) {
            ExcelDir.value = config.excel_dir || "../../config";
        }
    } catch (err) {
        console.error("加载策划配表目录配置失败:", err);
    }
};

/**
 * 保存统一策划配表目录配置
 */
export const saveExcelConfig = async () => {
    try {
        await ExcelConfigService.SaveConfig({
            excel_dir: ExcelDir.value
        });
    } catch (err) {
        console.error("保存策划配表目录配置失败:", err);
    }
};
