<script setup lang="ts">
// example-data-card - 内置示例数据加载卡片(Android 数据供给 C 方案)
//
// 独立组件,推翻/删除时:
//   1. 删本文件 + composables 无(逻辑内联)
//   2. 去 settings/index.vue 的 import ExampleDataCard 和 <ExampleDataCard/> 标签
//   3. 后端删 backend/pkg/exampledata/ 包
//
// 释放内置示例(go:embed resources 1.4M + 精简 fight_cases 10 个)到私有目录,
// 配置 function-test 指向。可重复加载(先清旧,系统干净)。
import { ref } from 'vue'
import { LoadExampleData } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/exampledata/exampledataservice'
import type { LoadResult } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/exampledata/models'
import { loadCases } from '@/pages/function-test/composables/Func'

const loading = ref(false)
const result = ref<LoadResult | null>(null)
const errorMsg = ref('')

async function loadExample() {
  errorMsg.value = ''
  loading.value = true
  try {
    result.value = await LoadExampleData()
    // 配置已写盘(function_test section),直接触发战斗测试页用例树重载:
    // loadCases 刷新配置(await GetConfig → JsonsDir/ExcelResourcesDir 新值)+ InitExcel +
    // GetCategories → Object.assign 更新全局 dataRef(用例树 :data="dataRef" 绑定)。
    // 无需 reload 应用——loadCases 开头 await GetConfig 修了 Option.ts 模块级竞态,
    // dataRef 是 reactive,战斗测试页用例树自动刷新(进页面即见 10 分类)。
    try {
      await loadCases()
    } catch (lcErr: any) {
      errorMsg.value = '示例已加载,用例树刷新失败: ' + (lcErr instanceof Error ? lcErr.message : String(lcErr))
    }
  } catch (e: any) {
    errorMsg.value = e instanceof Error ? e.message
      : (typeof e === 'string' ? e : (e?.message || JSON.stringify(e)))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <!-- 加载示例数据卡片(C 方案:内置示例,各端可用,用户主动触发) -->
  <n-card title="加载示例数据" class="setting-card">
    <div class="setting-row">
      <n-button type="primary" :loading="loading" :disabled="loading" @click="loadExample">
        {{ loading ? '加载中...' : '加载示例数据' }}
      </n-button>
    </div>
    <div v-if="result" style="margin-top: 10px;">
      <div style="color: #fff;">
        已加载: {{ result.fightCaseCount }} 个战斗用例 + game resources({{ result.fileCount }} 文件)
      </div>
      <div class="setting-hint" style="color: #67c23a; margin-top: 4px;">
        ✓ 已加载用例树,前往"战斗测试"页查看
      </div>
    </div>
    <div v-if="errorMsg" class="setting-hint" style="color: #ff6b6b; margin-top: 8px;">{{ errorMsg }}</div>
    <div class="setting-hint" style="margin-top: 8px;">
      释放内置示例(resources 1.4M + 精简用例)到私有目录,配置战斗测试页指向。可重复加载(先清旧再放新)
    </div>
  </n-card>
</template>
