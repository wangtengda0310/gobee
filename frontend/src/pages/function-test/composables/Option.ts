import {ref} from "vue";
import {FuncCaseConfigService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {saveModifiedCases} from "./Func";

export const JsonsDir = ref("json")
export const ExcelResourcesDir = ref("../rain-robot/project/xcard/xcard_excel/resources")
export const ServerAddr = ref("10.254.114.241")
export const Desc = ref("本地测试")
export const ServerPort = ref(20144)
export const RobotPrefix = ref("pf_qa")
export const SingleCaseRunCount = ref(1)
export const LoginTime = ref(1)
export const RoomOpTime = ref(300_000)
export const DebugLevel = ref(false)
export const DebugLog = ref(false)
export const Concurrency = ref(1)
export const AutoSave = ref(true)
export const InterceptEnabled = ref(false)

FuncCaseConfigService.GetConfig().then(res => {
    JsonsDir.value = res?.jsons_dir ?? "cases/fight_cases"
    ExcelResourcesDir.value = res?.excel_resources_dir ?? "../rain-robot/project/xcard/xcard_excel/resources"
    ServerAddr.value = res?.server_addr ?? "10.254.114.241"
    ServerPort.value = res?.server_port ?? 20144
    Desc.value = res?.desc ?? "本地测试"
    RobotPrefix.value = res?.robot_prefix ?? "pf_qa"
    SingleCaseRunCount.value = res?.single_case_run_count ?? 1
    LoginTime.value = res?.login_time ?? 1
    RoomOpTime.value = res?.room_op_time ?? 300_000
    DebugLevel.value = res?.debug_level ?? false
    DebugLog.value = res?.debug_log ?? false
    Concurrency.value = res?.concurrency ?? 1
    AutoSave.value = res?.auto_save ?? true
    InterceptEnabled.value = res?.intercept_enabled ?? false
}).catch(err => {
    console.log(err)
})

export let autoSaveInterval = ref<ReturnType<typeof setInterval>>()
export const openAutoSave = () => {
    return setInterval(() => {
        if (AutoSave.value == false) {
            clearInterval(autoSaveInterval.value)
            autoSaveInterval.value = undefined
            return
        }
        saveModifiedCases()
    }, 10_000)
}

// 自动执行一次
if (AutoSave.value && autoSaveInterval.value == null) {
    autoSaveInterval.value = openAutoSave()
}