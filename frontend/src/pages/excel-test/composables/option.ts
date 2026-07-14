/**
 * Excel 配置选项管理
 *
 * 管理 Excel 资源目录、用例目录和客户端项目路径的配置
 */
import {ref} from "vue";
import {ExcelConfigService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test";

/** Excel 资源目录路径 */
export const ExcelResourceDir = ref("C:/Users/v-tangfangda/GolandProjects/rain-excel-checker/xlsx/resources")

/** Excel 用例目录路径 */
export const ExcelCaseDir = ref("C:/Users/v-tangfangda/GolandProjects/rain-excel-checker/xlsx/test")

/** 客户端项目路径，用于资源校验 */
export const ClientPath = ref("")

// 初始化时从后端加载配置
ExcelConfigService.GetConfig().then(res => {
    ExcelResourceDir.value = res?.excel_resources_dir ?? "C:/Users/v-tangfangda/GolandProjects/rain-excel-checker/xlsx/resources"
    ExcelCaseDir.value = res?.excel_case_dir ?? "C:/Users/v-tangfangda/GolandProjects/rain-excel-checker/xlsx/test"
    ClientPath.value = res?.client_path ?? ""
}).catch(err => {
    console.log(err)
})

/**
 * 保存配置到后端
 */
export const saveExcelConfig = () => {
    return ExcelConfigService.SaveConfig({
        excel_resources_dir: ExcelResourceDir.value,
        excel_case_dir: ExcelCaseDir.value,
        client_path: ClientPath.value,
    })
}
