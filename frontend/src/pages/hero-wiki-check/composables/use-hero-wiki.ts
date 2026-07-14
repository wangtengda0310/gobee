/**
 * 武将 Wiki 检查页面 composable
 *
 * 从 index.vue 提取的所有状态管理和业务逻辑，
 * 包括配置管理、检查执行、diff 数据转换、筛选逻辑等。
 */
import {computed, ref, onMounted} from "vue"
import {useMessage} from "naive-ui"
import {HeroWikiResCheckService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/hero-wiki-check"
import * as HeroWikiConfigService from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/hero-wiki-check/herowikiconfigservice.js"
import {SkillDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill"
import {SkillLinesDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_lines"
import {HeroLinesDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_lines"
import {CountryDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/country"
import {HeroSkinItemDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_item"
import {HeroDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero"
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models"
import {DataContainer} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
import {DiffIndexMap} from "./hero-wiki.types"

/** diff 类型筛选 */
type DiffTypeFilter = 'added' | 'modified' | 'removed' | null

/** 删除武将的详情项（包含完整数据用于显示） */
interface RemovedHeroItem {
    eHeroId: string
    name: string
    heroInfo: HeroDiff       // 虚拟的 HeroDiff 对象,用于传递给 HeroPanel
    heroWikiData: HeroCompleteData | null // 完整的 Wiki 数据
}

// ========== 模块级全局状态 ==========
const excelDir = ref<string>('../../config/excel')
const oldJsonPath = ref<string>('tmp/test.json')

// 初始化配置（使用 Wails 自动生成的绑定）
HeroWikiConfigService.GetConfig().then(config => {
  excelDir.value = config?.excel_dir || '../../config/excel'
  oldJsonPath.value = config?.old_json_dir ? `${config.old_json_dir}/test.json` : 'tmp/test.json'
}).catch(err => {
  console.error('加载配置失败:', err)
  excelDir.value = '../../config/excel'
  oldJsonPath.value = 'tmp/test.json'
})

// 保存配置
const saveConfig = async () => {
  try {
    const oldJsonDir = oldJsonPath.value.replace(/\/[^/]+$/, '')
    await HeroWikiConfigService.UpdateConfig({
      excel_dir: excelDir.value,
      old_json_dir: oldJsonDir
    })
  } catch (error) {
    console.error('保存配置失败:', error)
    throw error
  }
}
const isLoading = ref(false)
const isSaving = ref(false)
const errorMsg = ref('')
const diffExcels = ref<DataContainer | null>(null)

// ========== 筛选状态 ==========
const searchName = ref('')
const filterCountry = ref<string[]>([])
const filterIsNewHero = ref<boolean | null>(null)
const filterIsGacha = ref<boolean | null>(null)
const filterIsOpen = ref<boolean | null>(null)
const filterDiffType = ref<DiffTypeFilter>(null)

/**
 * 武将 Wiki 检查页面 composable
 *
 * 封装配置状态、执行操作、diff 数据转换、筛选逻辑等全部业务逻辑
 */
export function useHeroWikiCheck() {
    const message = useMessage()

    // ========== 执行操作 ==========

    /** 执行检查，对比 Excel 配置中的武将数据变化 */
    const runCheck = async () => {
        isLoading.value = true
        errorMsg.value = ''
        try {
            const res = await HeroWikiResCheckService.Check(excelDir.value, oldJsonPath.value)
            console.log(res)
            diffExcels.value = res
        } catch (err) {
            console.log(err)
            errorMsg.value = String(err)
        } finally {
            isLoading.value = false
        }
    }

    /** 保存当前检查结果到 oldJsonPath 指定的路径 */
    const saveResult = async () => {
        if (!diffExcels.value) return
        isSaving.value = true
        try {
            await HeroWikiResCheckService.Save(oldJsonPath.value, diffExcels.value)
            message.success('保存成功')
        } catch (err) {
            message.error('保存失败: ' + String(err))
        } finally {
            isSaving.value = false
        }
    }

    // ========== Diff 数据转换 ==========

    /**
     * 将 diffExcels 中的各类 Diff 数组转为 Map 索引
     *
     * 将按 ID/枚举值可快速查找的结构返回,供 HeroPanel 使用
     */
    const transMap = (): DiffIndexMap => {
        const skillsDiffMap = diffExcels.value?.SkillDiff?.reduce((acc, cur) => {
            acc.set(Number(cur.Id), cur)
            return acc
        }, new Map<number, SkillDiff>()) ?? new Map<number, SkillDiff>()
        const skillLinesDiffMap = diffExcels.value?.SkillLineDiff?.reduce((acc, cur) => {
            acc.set(cur.Id, cur)
            return acc
        }, new Map<number, SkillLinesDiff>()) ?? new Map<number, SkillLinesDiff>()
        const skillLinesDiffMap_Enum = diffExcels.value?.SkillLineDiff?.reduce((acc, cur) => {
            acc.set(cur.SkillId, cur)
            return acc
        }, new Map<string, SkillLinesDiff>()) ?? new Map<string, SkillLinesDiff>()
        const heroLinesDiffMap = diffExcels.value?.HeroLinesDiff?.reduce((acc, cur) => {
            acc.set(cur.Id, cur)
            return acc
        }, new Map<number, HeroLinesDiff>()) ?? new Map<number, HeroLinesDiff>()
        const countryDiffMap = diffExcels.value?.CountryDiff?.reduce((acc, cur) => {
            acc.set(cur.ECountry, cur)
            return acc
        }, new Map<string, CountryDiff>()) ?? new Map<string, CountryDiff>()
        const heroSkinItemDiffMap = diffExcels.value?.HeroSkinItemDiff?.reduce((acc, cur) => {
            if (acc.has(Number(cur.HeroId))) {
                acc.get(Number(cur.HeroId))!.push(cur)
            } else {
                acc.set(Number(cur.HeroId), [cur])
            }
            return acc
        }, new Map<number, HeroSkinItemDiff[]>()) ?? new Map<number, HeroSkinItemDiff[]>()
        return {
            skillsDiffMap,
            skillLinesDiffMap,
            skillLinesDiffMap_Enum,
            heroLinesDiffMap,
            countryDiffMap,
            heroSkinItemDiffMap
        }
    }

    // ========== Diff 统计 ==========

    /** 是否存在 diff 结果 */
    const hasDiffResult = computed(() => {
        return diffExcels.value?.HeroWikiDiffResult?.HeroesDiff &&
            Object.keys(diffExcels.value?.HeroWikiDiffResult.HeroesDiff).length > 0
    })

    /** 全局 diff 统计信息 */
    const diffSummary = computed(() => {
        return diffExcels.value?.HeroWikiDiffResult?.Summary
    })

    // 筛选状态已提升到模块级

    /** 国家选项列表（动态从 CountryDiff 获取） */
    const countryOptions = computed(() => {
        if (!diffExcels.value?.CountryDiff) return []
        return diffExcels.value.CountryDiff.map(country => ({
            label: country.Name,      // 显示名称
            value: country.ECountry   // 枚举值(与 hero.Country 匹配)
        }))
    })

    // ========== 删除武将 ==========

    /** 删除武将的详情列表（包含完整数据用于显示） */
    const removedHeroesDetail = computed((): RemovedHeroItem[] => {
        if (!diffExcels.value?.HeroWikiDiffResult || !diffSummary.value?.RemovedHeroes) return []
        return diffSummary.value.RemovedHeroes.map(eHeroId => {
            const detail = diffExcels.value?.HeroWikiDiffResult?.HeroesDiff[eHeroId]
            const wikiData = diffExcels.value?.HeroWikiDiffResult?.RemovedHeroesData?.[eHeroId]
            // 从完整的 Wiki 数据中构建虚拟的 HeroDiff 对象
            const heroInfo: HeroDiff = new HeroDiff()
            // EHeroId 必须使用外层的 eHeroId,确保唯一性
            heroInfo.EHeroId = eHeroId
            if (wikiData?.Basic) {
                heroInfo.Id = wikiData.Basic.Id
                heroInfo.Name = wikiData.Basic.Name
                heroInfo.IsOpen = wikiData.Basic.IsOpen
                heroInfo.OpenDate = wikiData.Basic.OpenDate
                heroInfo.Gender = wikiData.Basic.Gender
                heroInfo.Point = wikiData.Basic.Point
                heroInfo.HpLimit = wikiData.Basic.HpLimit
                heroInfo.HandLimit = wikiData.Basic.HandLimit
                heroInfo.EquipLimit = wikiData.Basic.EquipLimit
                heroInfo.Country = wikiData.Basic.Country
                heroInfo.IsAlwaysZhuGong = wikiData.Basic.IsAlwaysZhuGong
                heroInfo.Skill = wikiData.Basic.ExcludeIdentity || []
                heroInfo.ExcludeIdentity = wikiData.Basic.ExcludeIdentity || []
                heroInfo.NotUseModeType = wikiData.Basic.NotUseModeType || []
                heroInfo.HeroType = wikiData.Basic.HeroType
                heroInfo.EHeroType = wikiData.Basic.EHeroType
                heroInfo.CanMelt = wikiData.Basic.CanMelt
                heroInfo.MeltName = wikiData.Basic.MeltName || []
                heroInfo.IsNewHero = wikiData.Basic.IsNewHero
                heroInfo.IsGacha = wikiData.Basic.IsGacha
                heroInfo.BelongExpansionPack = wikiData.Basic.BelongExpansionPack
            }
            return {
                eHeroId,
                name: detail?.Name || wikiData?.Basic?.Name || eHeroId,
                heroInfo,
                heroWikiData: wikiData || null
            }
        })
    })

    // ========== 筛选逻辑 ==========

    /** 设置 diff 类型筛选（再次点击已激活的筛选则取消） */
    const setDiffTypeFilter = (type: DiffTypeFilter) => {
        if (filterDiffType.value === type) {
            filterDiffType.value = null
        } else {
            filterDiffType.value = type
        }
    }

    /** 过滤后的武将列表 */
    const filteredHeroes = computed(() => {
        if (!diffExcels.value?.HeroDiff) return []
        return diffExcels.value.HeroDiff.filter((hero) => {
            // 名称搜索
            if (searchName.value && !hero.Name.includes(searchName.value)) {
                return false
            }
            // 势力筛选
            if (filterCountry.value.length > 0 && !filterCountry.value.includes(hero.Country)) {
                return false
            }
            // 新武将筛选
            if (filterIsNewHero.value !== null && hero.IsNewHero !== filterIsNewHero.value) {
                return false
            }
            // 抽卡武将筛选
            if (filterIsGacha.value !== null && hero.IsGacha !== filterIsGacha.value) {
                return false
            }
            // 是否开放筛选
            if (filterIsOpen.value !== null && hero.IsOpen !== filterIsOpen.value) {
                return false
            }
            // diff类型筛选
            if (filterDiffType.value === 'added') {
                const isAdded = diffSummary.value?.AddedHeroes?.includes(hero.EHeroId)
                if (!isAdded) return false
            } else if (filterDiffType.value === 'modified') {
                const isModified = diffSummary.value?.ModifiedHeroes?.includes(hero.EHeroId)
                if (!isModified) return false
            } else if (filterDiffType.value === 'removed') {
                // 当筛选删除时,正常武将列表不显示任何内容
                return false
            }
            return true
        })
    })

    /** 过滤后的锚点列表 */
    const filteredAnchors = computed(() => {
        if (!filteredHeroes.value) return []
        return filteredHeroes.value.map((hero, k) => ({
            seq: k,
            hero: hero
        }))
    })

    return {
        // 配置状态
        excelDir,
        oldJsonPath,
        isLoading,
        isSaving,
        errorMsg,
        diffExcels,
        // 执行操作
        runCheck,
        saveResult,
        saveConfig, // 新增保存配置方法
        // Diff 数据转换
        transMap,
        // Diff 统计
        hasDiffResult,
        diffSummary,
        // 筛选状态
        searchName,
        filterCountry,
        filterIsNewHero,
        filterIsGacha,
        filterIsOpen,
        filterDiffType,
        // 筛选逻辑
        countryOptions,
        setDiffTypeFilter,
        filteredHeroes,
        filteredAnchors,
        // 删除武将
        removedHeroesDetail,
    }
}
