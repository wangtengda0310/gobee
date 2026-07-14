/**
 * Excel 树形数据和历史记录管理
 *
 * 管理响应式的 Excel 文件树数据和统计信息
 */
import {computed, reactive} from "vue";
import {TreeOption} from "naive-ui";
import {ExtraExcelTreeOption} from "./use-tree";
import {createData} from "./use-excel-check-data";

/** Excel 文件树数据 */
export const dataRef = reactive<(TreeOption & ExtraExcelTreeOption)[]>(createData() as (TreeOption & ExtraExcelTreeOption)[])

/** 成功和失败的 Sheet 数量统计 */
export const succAndFailSheetNum = computed(()=>{
    return dataRef.reduce((acc, cur)=>{
        const statistic = cur.children ? cur.children.reduce((sAcc,sCur)=>{
            if ((sCur as ExtraExcelTreeOption).sheetError == '') {
                sAcc[0]+=1
            } else {
                sAcc[1]+=1
            }
            return sAcc
        }, [0,0]) : [0,0]
        acc[0]+=statistic[0]
        acc[1]+=statistic[1]
        return acc
    },[0,0])
})
