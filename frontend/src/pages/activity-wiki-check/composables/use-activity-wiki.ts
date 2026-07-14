/**
 * 活动 Wiki 检查页面 composable
 *
 * 从 index.vue 提取的所有状态管理和业务逻辑，
 * 包括配置管理、检查执行、筛选逻辑等。
 * 参考 hero-wiki-check/composables/use-hero-wiki.ts 的 composable 模式。
 */
import {computed, ref, watch} from "vue"
import {useMessage} from "naive-ui"
import {Events} from "@wailsio/runtime"
import {Check, GetRuleCoverageWithErrors} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/activitywikicheckservice.js"
import * as ActivityWikiConfigService from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/activitywikiconfigservice.js"
import {ActivityWikiDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/activitywiki_def/models.js"
import type {ActivityCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/activitywiki_def/models.js"
import type {RuleCoverageData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/models.js"
import {useResourceCheck} from "./use-resource-check"

// ========== 模块级全局状态 ==========
const excelDir = ref<string>('../../config/excel')
const secondPath = ref('')
const clientPath = ref<string>('')

// 初始化配置（使用 Wails 自动生成的绑定）
ActivityWikiConfigService.GetConfig().then(config => {
  excelDir.value = config?.excel_dir || '../../config/excel'
  clientPath.value = config?.client_path || ''
}).catch(err => {
  console.error('加载配置失败:', err)
  excelDir.value = '../../config/excel'
})

// 保存配置
const saveConfig = async () => {
  try {
    await ActivityWikiConfigService.UpdateConfig({
      excel_dir: excelDir.value,
      client_path: clientPath.value,
    })
  } catch (error) {
    console.error('保存配置失败:', error)
    throw error
  }
}
const isLoading = ref(false)
const errorMsg = ref('')

// ========== 活动 diff 数据 ==========
const activityWikiDiff = ref<ActivityWikiDiff | null>(null)

// ========== 规则覆盖数据 ==========
const ruleCoverage = ref<RuleCoverageData | null>(null)

// ========== 筛选状态 ==========
const searchName = ref('')
const filterActivityType = ref<string | null>(null)
const filterShowTab = ref<boolean | null>(null)

// ========== 规则覆盖数据加载 ==========

/** 加载规则覆盖数据（从后端缓存读取配表检查结果，合并错误计数） */
const loadRuleCoverage = async () => {
  try {
    const data = await GetRuleCoverageWithErrors("cases/excel_cases", excelDir.value)
    ruleCoverage.value = data
  } catch (err) {
    console.error('加载规则覆盖数据失败:', err)
  }
}

// ========== 监听配表检查完成事件（模块级，只注册一次） ==========
// 配表测试页面执行检查后发送此事件，Wiki 页面收到后刷新角标数据
Events.On('excelCheckCompleted', () => {
  if (activityWikiDiff.value) {
    loadRuleCoverage()
  }
})

/**
 * 收集所有活动数据中的资源路径字段
 *
 * 遍历活动列表，提取需要检查的资源路径（图片、音频、预制体等）
 */
function collectResourcePaths(activities: Record<string, any>): string[] {
  const paths: string[] = []
  const add = (v: string | undefined | null) => {
    if (v && typeof v === 'string') paths.push(v)
  }

  for (const act of Object.values(activities)) {
    if (!act) continue
    // activity-panel.vue 中的资源字段
    // ShopGood.Icon
    if (act.ShopGoods && Array.isArray(act.ShopGoods)) {
      for (const goods of act.ShopGoods) {
        if (goods) add(goods.Icon)
      }
    }
    // HeroSkinCollition
    add(act.HeroSkinCollition?.NameImg)
    add(act.HeroSkinCollition?.NameBg)
    // ItemHeroSkin
    add(act.ItemHeroSkin?.Path)
    // HeroSkinItem
    add(act.HeroSkinItem?.SeatSpecialImg)
    add(act.HeroSkinItem?.CollitionTagImg)
    // HeroSkinSpine
    add(act.HeroSkinSpine?.MainBgFx)
    add(act.HeroSkinSpine?.SpineAnimAudio)
    add(act.HeroSkinSpine?.KillAudio)
    // Pet（灵宠，数组）
    if (act.Pets && Array.isArray(act.Pets)) {
      for (const pet of act.Pets) {
        if (!pet) continue
        add(pet.PrefabPath)
        add(pet.SquareHeadIcon)
        add(pet.HeadIcon)
        add(pet.Silhouette)
        add(pet.PopBg)
        add(pet.PopIcon)
        add(pet.PopTitle)
        add(pet.PetWeekTaskBg)
      }
    }
    // DrawPet（结缘亭，三期数据）
    for (const key of ['PrevDrawPet', 'DrawPet', 'NextDrawPet']) {
      const dp = act[key]
      if (dp) {
        add(dp.DrawPetTitleIcon)
        add(dp.DrawPetTitleDescBg)
        add(dp.DrawRuleContent)
      }
    }
  }
  return [...new Set(paths)]
}

/**
 * 活动 Wiki 检查页面 composable
 *
 * 封装配置状态、执行操作、diff 数据、筛选逻辑等全部业务逻辑
 */
export function useActivityWikiCheck() {
  const message = useMessage()

  // 资源检查 composable（依赖 clientPath）
  const resourceCheck = useResourceCheck(clientPath)

  // ========== 执行操作 ==========

  /** 执行检查，对比 Excel 配置中的活动数据变化 */
  const runCheck = async () => {
    isLoading.value = true
    errorMsg.value = ''
    try {
      const res = await Check(excelDir.value)
      if (res?.ActivityWikiDiff) {
        activityWikiDiff.value = res.ActivityWikiDiff
      } else {
        message.warning('未找到活动数据')
      }
      // 检查完成后加载规则覆盖数据（从缓存读取配表检查结果）
      await loadRuleCoverage()
      // 检查完成后收集所有资源路径并触发资源检查
      if (clientPath.value && activityWikiDiff.value?.Activities) {
        const paths = collectResourcePaths(activityWikiDiff.value.Activities)
        if (paths.length > 0) {
          resourceCheck.checkPaths(paths)
        }
      }
    } catch (err) {
      console.error(err)
      errorMsg.value = String(err)
      message.error('检查失败: ' + String(err))
    } finally {
      isLoading.value = false
    }
  }

  // 筛选状态已提升到模块级

  /** 计算属性：活动类型选项列表（动态获取） */
  const activityTypeOptions = computed(() => {
    if (!activityWikiDiff.value?.Activities) return []
    const typeSet = new Set<string>()
    Object.values(activityWikiDiff.value.Activities).forEach((activity: ActivityCompleteData | null | undefined) => {
      if (activity?.Basic?.ActivityType) {
        typeSet.add(activity.Basic.ActivityType)
      }
    })
    return Array.from(typeSet).map(type => ({
      label: type,
      value: type
    }))
  })

  /** 计算属性：过滤后的活动列表 */
  const filteredActivities = computed(() => {
    if (!activityWikiDiff.value?.Activities) return []
    const activities = Object.values(activityWikiDiff.value.Activities).filter((a): a is ActivityCompleteData => a !== null)
    return activities.filter((activity: ActivityCompleteData) => {
      if (!activity) return false
      // 名称搜索
      if (searchName.value && !activity.Basic?.Name?.includes(searchName.value)) {
        return false
      }
      // 活动类型筛选
      if (filterActivityType.value && activity.Basic?.ActivityType !== filterActivityType.value) {
        return false
      }
      // 是否显示页签筛选
      if (filterShowTab.value !== null && activity.Basic?.ShowTab !== filterShowTab.value) {
        return false
      }
      return true
    })
  })

  /** 计算属性：过滤后的锚点列表 */
  const filteredAnchors = computed(() => {
    const anchors = filteredActivities.value.map((activity, index) => ({
      seq: index,
      activity: activity,
      isSeasonPass: false as boolean
    }))
    // 如果有战令数据，在锚点列表末尾添加战令节点
    if (filteredSeasonPasses.value.length > 0) {
      anchors.push({
        seq: anchors.length,
        activity: null as any,
        isSeasonPass: true as boolean
      })
    }
    return anchors
  })

  /** 计算属性：战令列表（从 activityWikiDiff.SeasonPasses 获取） */
  const filteredSeasonPasses = computed(() => {
    if (!activityWikiDiff.value?.SeasonPasses) return []
    return Object.values(activityWikiDiff.value.SeasonPasses).filter((sp): sp is NonNullable<typeof sp> => sp !== null)
  })

  return {
    // 配置状态
    excelDir,
    secondPath,
    clientPath,
    isLoading,
    errorMsg,
    // 数据
    activityWikiDiff,
    ruleCoverage,
    // 执行操作
    runCheck,
    saveConfig,
    // 筛选状态
    searchName,
    filterActivityType,
    filterShowTab,
    // 筛选逻辑
    activityTypeOptions,
    filteredActivities,
    filteredAnchors,
    // 战令数据
    filteredSeasonPasses,
    // 资源检查
    resourceCheck,
  }
}
