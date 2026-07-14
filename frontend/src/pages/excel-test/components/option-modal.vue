<!--
  OptionModal - Excel 测试配置模态框

  用于配置 Excel 资源目录和用例目录
-->
<script setup lang="ts">
import {useMessage} from "naive-ui";
import {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";
import {ExcelConfigService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test";
import {ExcelCaseDir, ExcelResourceDir, ClientPath} from "../composables/option";
import {showExcelCaseOptionModal} from "../composables/use-excel-check-data";

const bodyStyle = {
  width: '600px'
}
const segmented = {
  content: 'soft',
  footer: 'soft'
} as const

const saveConfig = (msg: MessageApiInjection) => {

  ExcelConfigService.SaveConfig({
    excel_resources_dir: ExcelResourceDir.value,
    excel_case_dir: ExcelCaseDir.value,
    client_path: ClientPath.value,
  }).then(res => {
    msg.success("已保存配置")
  }).catch(err => {

  })
}

const message = useMessage()
</script>

<template>
  <n-modal
      v-model:show="showExcelCaseOptionModal"
      transform-origin="center"
      class="custom-card"
      preset="card"
      :style="bodyStyle"
      title="设置"
      size="small"
      :bordered="false"
      :segmented="segmented"
      :show-mask="true"
      @after-leave="()=>{saveConfig(message)}"
  >
    <template #header-extra>
    </template>
    <div style="display: flex; flex-direction: column; gap: 10px">
      <div>
        <div>配表目录</div>
        <n-input v-model:value="ExcelResourceDir" placeholder="配表目录"/>
      </div>
      <div>
        <div>用例目录</div>
        <n-input v-model:value="ExcelCaseDir" placeholder="用例目录"/>
      </div>
    </div>
    <template #footer>
    </template>
  </n-modal>
</template>

<style scoped>

</style>
