import {ref} from 'vue'
import {ExtraCaseTreeOption} from "./use-case-data";
import {TreeOption} from "naive-ui";

// 添加类别
export const showAddCateModal = ref(false)
export const addCateModalData = ref({
    text: "",
    index: -1
})

// 添加Case
export const showAddCaseModal = ref(false)
export const addCaseModalData = ref<{
    cate: string,
    index: number,
    case?: string
}>({
    cate: "",
    index: 0,
})

// 添加Case
export const showFuncCaseOptionModal = ref(false)

// 修改Case名
export const showModifyCaseNameModal = ref(false)
export const modifyCaseNameModalData = ref<{
    caseRef?: TreeOption & ExtraCaseTreeOption
    newName: string
}>({
    caseRef: undefined,
    newName: ""
})

// 修改Cate名
export const showModifyCateNameModal = ref(false)
export const modifyCateNameModalData = ref<{
    cateRef?: TreeOption & ExtraCaseTreeOption
    newName: string
}>({
    cateRef: undefined,
    newName: ""
})