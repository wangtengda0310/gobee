<!--
  TableRulePanel - 表级校验规则面板

  通用组件：展示当前 Sheet 的所有表级校验规则，支持启用/禁用和参数配置。
  通过 props 接收规则元数据和当前规则列表，通过 emits 通知父组件规则变更。

  支持同类型规则的多实例：每种规则类型可以添加多条实例，每条实例有独立的参数配置。
-->
<script setup lang="ts">
import {h, ref} from "vue";
import {ETableRule, TableRule, TableRuleMeta, TableRuleParamDef} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule";
import {NButton, NIcon, NSelect} from "naive-ui";
import {TrashOutline, ChevronDownOutline, ChevronUpOutline} from "@vicons/ionicons5";

// 折叠状态
const collapsed = ref(true)

const props = defineProps<{
  /** 所有可用的表级规则元数据（从后端获取） */
  ruleMetas: TableRuleMeta[]
  /** 当前已配置的表级规则列表 */
  rules: TableRule[]
  /** 是否已加载规则元数据 */
  loaded: boolean
}>()

const emit = defineEmits<{
  /** 规则列表变更时触发，由父组件负责持久化 */
  (e: 'update:rules', rules: TableRule[]): void
}>()

// 从 ParamDefs 构建默认参数 map
const buildDefaultParams = (paramDefs: TableRuleParamDef[]): { [key: string]: string } => {
  const params: { [key: string]: string } = {}
  for (const def of paramDefs) {
    if (def.default) {
      params[String(def.key)] = def.default
    }
  }
  return params
}

// 获取某类型的所有规则实例
const getRulesByType = (ruleType: ETableRule): TableRule[] => {
  return props.rules.filter(r => r.type === ruleType)
}

// 判断某类型是否有任何已启用的实例
const getTypeEnabled = (ruleType: ETableRule): boolean => {
  const rulesOfType = getRulesByType(ruleType)
  if (rulesOfType.length === 0) {
    return true // 默认启用，方便用户添加实例
  }
  return rulesOfType.some(r => r.enabled)
}

// 切换整个规则类型的启用状态（控制该类型下所有实例）
const toggleRuleEnabled = (ruleType: ETableRule) => {
  const rules = [...props.rules]
  const rulesOfType = rules.filter(r => r.type === ruleType)
  const currentlyEnabled = rulesOfType.some(r => r.enabled)

  if (rulesOfType.length === 0) {
    // 没有实例时，切换不改变任何状态，保持默认启用
    return
  }

  // 切换所有实例的启用状态
  for (const rule of rulesOfType) {
    rule.enabled = !currentlyEnabled
  }
  emit('update:rules', rules)
}

// 添加新实例
const addRuleInstance = (ruleType: ETableRule, meta: TableRuleMeta) => {
  const rules = [...props.rules]
  const newRule: TableRule = {
    type: ruleType,
    displayName: meta.displayName,
    uuid: crypto.randomUUID(),
    description: meta.description,
    params: buildDefaultParams(meta.paramDefs),
    enabled: true
  }
  rules.push(newRule)
  emit('update:rules', rules)
}

// 删除指定实例
const removeRuleInstance = (uuid: string) => {
  const rules = props.rules.filter(r => r.uuid !== uuid)
  emit('update:rules', rules)
}

// 更新指定实例的参数
const updateRuleInstanceParams = (uuid: string, paramName: string, value: string) => {
  const rules = [...props.rules]
  const rule = rules.find(r => r.uuid === uuid)
  if (!rule) return

  if (!rule.params) {
    rule.params = {}
  }
  rule.params[paramName] = value
  emit('update:rules', rules)
}

// 切换指定实例的启用状态
const toggleRuleInstanceEnabled = (uuid: string, enabled: boolean) => {
  const rules = [...props.rules]
  const rule = rules.find(r => r.uuid === uuid)
  if (!rule) return

  rule.enabled = enabled
  emit('update:rules', rules)
}
</script>

<template>
  <n-card
      :title="()=>{
        return h('div', {style: 'display: flex; gap: 10px; align-items: center;'}, [
          h('span', {style: 'color: #FF6B6B; min-width: 28px'}, '0.'),
          h('span', {style: 'min-width: 200px; color: #FF6B6B'}, '表级校验规则'),
          h('span', {style: 'color: #888; min-width: 200px'}, 'TABLE_RULES'),
          h('span', {style: 'color: #FF9F43; min-width: 200px'}, '表级'),
        ])
      }"
      id="tableRules"
      header-style="font-size: 18px; padding: 10px; background: linear-gradient(90deg, #2d1f1f 0%, #1a1a2e 100%)">
    <template #header-extra>
      <n-button text @click="collapsed = !collapsed" style="font-size: 20px; color: #aaa;">
        <n-icon size="22">
          <ChevronUpOutline v-if="!collapsed" />
          <ChevronDownOutline v-else />
        </n-icon>
      </n-button>
    </template>

    <template v-if="!collapsed">
    <!-- 加载中提示 -->
    <div v-if="!loaded" style="padding: 20px; text-align: center; color: #888">
      加载规则列表中...
    </div>

    <!-- 规则列表 -->
    <div v-else style="display: flex; flex-direction: column; gap: 10px">
      <n-card v-for="meta in ruleMetas" :key="meta.type"
              size="small"
              :style="{
                borderLeft: getTypeEnabled(meta.type) ? '3px solid #63e6be' : '3px solid #495057',
                opacity: getTypeEnabled(meta.type) ? 1 : 0.6
              }">

        <div style="display: flex; flex-direction: column; gap: 8px">
          <!-- 规则标题和启用开关 -->
          <div style="display: flex; justify-content: space-between; align-items: center">
            <div>
              <span style="font-weight: bold">{{ meta.displayName }}</span>
              <n-tag size="small" type="info" style="margin-left: 8px">{{ meta.type }}</n-tag>
            </div>
            <n-switch
                :value="getTypeEnabled(meta.type)"
                @update:value="toggleRuleEnabled(meta.type)">
              <template #checked>启用</template>
              <template #unchecked>禁用</template>
            </n-switch>
          </div>

          <!-- 规则描述 -->
          <div v-if="meta.description" style="color: #888; font-size: 12px">
            {{ meta.description }}
          </div>
          <!-- 规则实例列表（仅启用时显示） -->
          <div v-if="getTypeEnabled(meta.type) && meta.paramDefs && meta.paramDefs.length > 0"
               style="display: flex; flex-direction: column; gap: 12px; padding-top: 8px; border-top: 1px solid #333">

            <!-- 每个实例独立显示 -->
            <div v-for="(rule, index) in getRulesByType(meta.type)" :key="rule.uuid"
                 style="display: flex; flex-direction: column; gap: 8px; padding: 10px; background: rgba(0,0,0,0.2); border-radius: 6px;">

              <!-- 实例头部：序号、启用开关、删除按钮 -->
              <div style="display: flex; justify-content: space-between; align-items: center">
                <div style="display: flex; align-items: center; gap: 8px">
                  <span style="font-size: 12px; color: #aaa; font-weight: bold">实例 {{ index + 1 }}</span>
                  <n-switch
                      size="small"
                      :value="rule.enabled"
                      @update:value="(v: boolean) => toggleRuleInstanceEnabled(rule.uuid, v)">
                    <template #checked>启用</template>
                    <template #unchecked>禁用</template>
                  </n-switch>
                </div>
                <n-button
                    size="tiny"
                    type="error"
                    ghost
                    @click="removeRuleInstance(rule.uuid)">
                  <template #icon>
                    <n-icon><TrashOutline /></n-icon>
                  </template>
                  删除
                </n-button>
              </div>

              <!-- 实例参数（仅实例启用时显示） -->
              <div v-if="rule.enabled"
                   style="display: flex; flex-wrap: wrap; gap: 10px;">
                <div v-for="paramDef in meta.paramDefs" :key="paramDef.key"
                     style="display: flex; flex-direction: column; gap: 4px; min-width: 150px">
                  <label style="font-size: 12px; color: #aaa">{{ paramDef.label }}</label>
                  <n-input
                      v-if="paramDef.type === 'string'"
                      :value="rule.params?.[paramDef.key] || paramDef.default"
                      @update:value="(v: string) => updateRuleInstanceParams(rule.uuid, String(paramDef.key), v)"
                      :placeholder="paramDef.default"
                      size="small"
                  />
                  <n-input-number
                      v-else-if="paramDef.type === 'number'"
                      :value="Number(rule.params?.[paramDef.key] || paramDef.default)"
                      @update:value="(v: number | null) => updateRuleInstanceParams(rule.uuid, String(paramDef.key), String(v || paramDef.default))"
                      :placeholder="paramDef.default"
                      size="small"
                  />
                  <n-select
                      v-else-if="paramDef.type === 'select'"
                      :value="rule.params?.[paramDef.key] || paramDef.default"
                      @update:value="(v: string) => updateRuleInstanceParams(rule.uuid, String(paramDef.key), v)"
                      :options="(paramDef.options || []).map((o: any) => ({label: o.label, value: o.value}))"
                      size="small"
                  />
                  <n-input
                      v-else
                      :value="rule.params?.[paramDef.key] || paramDef.default"
                      @update:value="(v: string) => updateRuleInstanceParams(rule.uuid, String(paramDef.key), v)"
                      :placeholder="paramDef.default"
                      size="small"
                  />
                </div>
              </div>
            </div>

            <!--
              添加实例按钮：仅 COL_CONTINUOUS_CHECK 支持多实例
              原因：COL_CONTINUOUS_CHECK 需要针对不同列配置不同检查模式（如一个实例检查 SeasonStartTime 的日期间隔，另一个实例检查 ID 列的严格递增）。
              其他规则（包括有参数的通用通知规则如 NEW_ROW_NOTIFY）虽然也有 ParamDefs，但它们的参数是全局配置（如 Git 路径、ID 列名），添加多个实例不会产生不同的检查行为，反而造成用户困惑。
            -->
            <n-button
                v-if="meta.type === 'COL_CONTINUOUS_CHECK'"
                size="small"
                type="primary"
                ghost
                @click="addRuleInstance(meta.type, meta)">
              + 添加实例
            </n-button>
          </div>

          <!-- 无参数的规则类型：不显示添加实例按钮 -->
        </div>
      </n-card>
    </div>
    </template>
  </n-card>
</template>
