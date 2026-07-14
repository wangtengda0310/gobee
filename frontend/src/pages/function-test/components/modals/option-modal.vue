<script setup lang="ts">
import {showFuncCaseOptionModal} from "../../composables/Modals";
import {
  AutoSave,
  autoSaveInterval,
  Concurrency,
  DebugLevel,
  DebugLog,
  Desc,
  ExcelResourcesDir,
  InterceptEnabled,
  JsonsDir,
  LoginTime,
  openAutoSave,
  RobotPrefix,
  RoomOpTime,
  ServerAddr,
  ServerPort,
  SingleCaseRunCount
} from "../../composables/Option";
// 飞书配置从全局设置导入（用于保存时同步）
import { FeiShuGuid, FeiShuNtf } from "../../../settings/composables/use-settings";
import {computed} from "vue";
import {useMessage} from "naive-ui";
import {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";
import {FuncCaseConfigService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import PathConfigInput from "@shared/components/path-config-input/index.vue";

const bodyStyle = {
  width: '600px'
}
const segmented = {
  content: 'soft',
  footer: 'soft'
} as const

const options = computed(() => {
  const prefix = RobotPrefix.value.endsWith('_qa') ? RobotPrefix.value : RobotPrefix.value + '_qa'
  return [{
    label: prefix,
    value: prefix
  }]
})

const saveConfig = (msg: MessageApiInjection) => {
  if (AutoSave.value && autoSaveInterval.value == null) {
    autoSaveInterval.value = openAutoSave()
  }

  FuncCaseConfigService.SaveConfig({
    jsons_dir: JsonsDir.value,
    excel_resources_dir: ExcelResourcesDir.value,
    server_addr: ServerAddr.value,
    server_port: ServerPort.value,
    desc: Desc.value,
    robot_prefix: RobotPrefix.value,
    single_case_run_count: SingleCaseRunCount.value,
    login_time: LoginTime.value,
    room_op_time: RoomOpTime.value,
    fei_shu_ntf: FeiShuNtf.value,
    fei_shu_guid: FeiShuGuid.value,
    debug_level: DebugLevel.value,
    debug_log: DebugLog.value,
    concurrency: Concurrency.value,
    auto_save: AutoSave.value,
    intercept_enabled: InterceptEnabled.value
  }).then(res => {
    msg.success("已保存配置")
  }).catch(err => {

  })
}

const message = useMessage()
</script>

<template>
  <n-modal
      v-model:show="showFuncCaseOptionModal"
      transform-origin="center"
      class="custom-card"
      preset="card"
      :style="bodyStyle"
      title="设置"
      size="small"
      :bordered="false"
      :segmented="segmented"
      :show-mask="false"
      @after-leave="()=>{saveConfig(message)}"
  >
    <template #header-extra>
    </template>
    <div style="display: flex; flex-direction: column; gap: 10px">
      <PathConfigInput
        v-model:excel-dir="ExcelResourcesDir"
        v-model:second-value="JsonsDir"
        excel-label="Excel资源"
        second-label="用例目录"
        excel-placeholder="Excel 资源目录路径"
        second-placeholder="用例文件夹位置"
        layout="flex"
        :on-save="() => saveConfig(message)"
      />
      <div style="display: flex; gap: 10px; align-items: center; justify-content: start">
        <div style="flex: 1 1 0">
          <div>服务器地址</div>
          <n-input v-model:value="ServerAddr" placeholder="服务器地址"/>
        </div>
        <div style="flex: 1 1 0">
          <div>服务器端口</div>
          <n-input-number v-model:value="ServerPort" placeholder="服务器地址"/>
        </div>
      </div>
      <div>
        <div>用例描述</div>
        <n-input v-model:value="Desc" placeholder="用例描述"/>
      </div>
      <div style="display: flex; gap: 10px; align-items: center; justify-content: start">
        <div style="flex: 1 1 0">
          <div>机器人前缀</div>
          <n-auto-complete
              v-model:value="RobotPrefix"
              :input-props="{
            autocomplete: 'disabled',
          }"
              :options="options"
              placeholder="机器人前缀"
              clearable
          />
        </div>
        <div style="flex: 1 1 0">
          <div>单个用例执行次数</div>
          <n-input-number v-model:value="SingleCaseRunCount" placeholder="单个用例执行次数" disabled/>
        </div>
        <div style="flex: 1 1 0">
          <div>并发量</div>
          <n-input-number v-model:value="Concurrency" placeholder="对局操作时间(单位s)" min="1" max="300"/>
        </div>
        <div style="flex: 1 1 0">
          <div>登录打散时间</div>
          <n-input-number v-model:value="LoginTime" placeholder="登录打散时间" min="0" max="999999"/>
        </div>
      </div>
      <div style="display: flex; gap: 10px; align-items: center; justify-content: start">
        <div style="flex: 1 1 0">
          <div>对局操作时间</div>
          <n-input-number v-model:value="RoomOpTime" placeholder="对局操作时间(单位s)" min="0" max="999999"/>
        </div>
        <div style="flex: 1 1 0">
          <div>开启DEBUG等级日志</div>
          <n-switch v-model:value="DebugLevel" :round="false">
            <template #checked>
              开启
            </template>
            <template #unchecked>
              关闭
            </template>
          </n-switch>
        </div>
        <div style="flex: 1 1 0">
          <div>打印压测机器人日志</div>
          <n-switch v-model:value="DebugLog" :round="false">
            <template #checked>
              开启
            </template>
            <template #unchecked>
              关闭
            </template>
          </n-switch>
        </div>
      </div>
      <div style="display: flex; justify-content: space-between">
        <div>自动保存:</div>
        <n-switch v-model:value="AutoSave" placeholder="自动保存" :round="false">
          <template #checked>
            开启
          </template>
          <template #unchecked>
            关闭
          </template>
        </n-switch>
      </div>
    </div>
    <template #footer>
    </template>
  </n-modal>
</template>

<style scoped>

</style>