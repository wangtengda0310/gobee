<script setup lang="ts">
import {computed, onUnmounted, ref} from "vue";
import {activeKey, menuOptions} from "./composables/Menu";
import {activeTab as fightActiveTab} from "./composables/Func";
import {
  cascade,
  checkedKeysRef,
  checkStrategy,
  expandedKeysRef,
  handleCheckedKeysChange,
  handleDrop,
  handleExpandedKeysChange,
  handleLoad,
  nodeProps,
  renderLabel
} from "./composables/Tree";
import {handleTreeSelectedKeysUpdate, selectedCaseKeysRef} from "./composables/use-case-selection";
import {pattern, showIrrelevantNodes} from "../../shared/composables/use-tree-search";
import {
  handleClickOutside,
  handleSelect,
  optionsRef,
  showDropdownRef,
  xRef,
  yRef
} from "../../shared/composables/use-tree-dropdown";
import AddCateModal from "./components/modals/add-cate-modal.vue";
import AddCaseModal from "./components/modals/add-case-modal.vue";
import {Events} from "@wailsio/runtime";
import {insertLogCache} from "./composables/RobotTestLog";
import InitYanWuPanel from "./components/init-yanwu-panel.vue";
import StepsPanel from "./components/steps-panel.vue";
import RobotTestLog from "./components/robot-test-log.vue";
import {dataRef, nowCaseData, showCasesDesc, ExtraCaseTreeOption} from "./composables/use-case-data";
import OptionModal from "./components/modals/option-modal.vue";
import {useDialog, useMessage} from "naive-ui";
import RenameCaseModal from "./components/modals/rename-case-modal.vue";
import RenameCateModal from "./components/modals/rename-cate-modal.vue";
import FooterCaseLogStatistic from "./components/footer-case-log-statistic.vue";
import {actionsSelectOption} from "./composables/StepActionsAndAssetsSelect";
import {TreeOption} from "naive-ui";

/** 底栏统计（从 FooterStatistic.ts 内联） */
const footerStatisticCaseNum = computed(() =>
    dataRef.reduce((previousValue, currentValue) => {
      previousValue += currentValue.children ? currentValue.children.length : 0
      return previousValue
    }, 0)
)
const footerStatisticStepNum = computed(() =>
    dataRef.reduce((previousValue, currentValue) => {
      previousValue += currentValue.children ? currentValue.children.reduce((acc: number, cur: (TreeOption & ExtraCaseTreeOption)) => {
        acc += cur.caseSteps ? cur.caseSteps.length : 0
        return acc
      }, 0) : 0
      return previousValue
    }, 0)
)

const inverted = ref(false)

// 订阅机器人执行日志事件
const unsubscribe = Events.On('robotLog', (msg) => {
  insertLogCache(msg.data[1], msg.data[0])
})

const dialog = useDialog()
const message = useMessage()

onUnmounted(() => {
  unsubscribe()
})
</script>

<template>
  <div id="Test">
    <n-layout position="absolute">
      <n-layout-header style="height: 34px; display: flex; align-items: center" :inverted="inverted" bordered>
        <n-menu mode="horizontal" :inverted="inverted" :options="menuOptions(dialog, message)"
                v-model:value="activeKey"/>
      </n-layout-header>
      <n-layout position="absolute" style="top: 34px; bottom: 64px" has-sider>
        <n-layout-sider
            bordered
            show-trigger
            collapse-mode="width"
            :collapsed-width="50"
            :width="240"
            :native-scrollbar="false"
            :inverted="inverted"
            :show-collapsed-content="false"
        >
          <n-flex vertical>
            <n-input v-model:value="pattern" placeholder="搜索"/>
            <div style="display: flex; justify-content: space-around">
              <n-switch v-model:value="showIrrelevantNodes" :round="false">
                <template #checked>
                  仅展示过滤
                </template>
                <template #unchecked>
                  展示全部
                </template>
              </n-switch>
              <n-switch v-model:value="showCasesDesc" :round="false">
                <template #checked>
                  显示描述
                </template>
                <template #unchecked>
                  隐藏描述
                </template>
              </n-switch>
            </div>
            <!--            checkable-->
            <n-tree
                block-line
                draggable
                selectable
                :data="dataRef"
                :selected-keys="selectedCaseKeysRef"
                @update:selected-keys="handleTreeSelectedKeysUpdate"
                :checked-keys="checkedKeysRef"
                :on-load="handleLoad"
                :expanded-keys="expandedKeysRef"
                :check-strategy="checkStrategy"
                :allow-checking-not-loaded="cascade"
                :cascade="cascade"
                @drop="handleDrop"
                @update:checked-keys="handleCheckedKeysChange"
                @update:expanded-keys="handleExpandedKeysChange"
                :show-irrelevant-nodes="showIrrelevantNodes"
                :pattern="pattern"
                :node-props="(e)=>nodeProps(e, dialog, message)"
                checkbox-placement="left"
                :render-label="renderLabel"
            />
            <n-dropdown
                trigger="manual"
                placement="bottom-start"
                :show="showDropdownRef"
                :options="optionsRef as any"
                :x="xRef"
                :y="yRef"
                @select="handleSelect"
                @clickoutside="handleClickOutside"
            />
            <AddCateModal/>
            <AddCaseModal/>
            <RenameCaseModal/>
            <RenameCateModal/>
            <OptionModal/>
          </n-flex>
        </n-layout-sider>
        <n-flex id="Content" vertical style="width: 100%; flex: 1; min-width: 0; min-height: 0; overflow: hidden">
          <n-tabs
              class="fight-main-tabs"
              data-testid="fight-main-tabs"
              size="small"
              :bar-width="100"
              type="line"
              justify-content="space-evenly"
              style="height: 100%; flex: 1; min-height: 0"
              v-model:value="fightActiveTab"
          >
            <n-tab-pane style="padding: 1px; height: 100%; overflow: hidden" name="option" tab="用例配置" display-directive="show:lazy">
              <n-scrollbar style="max-height: 100%">
                <InitYanWuPanel/>
              </n-scrollbar>
            </n-tab-pane>
            <n-tab-pane style="box-sizing: border-box; padding: 5px; height: 100%; overflow: hidden; display: flex" name="step"
                        tab="用例步骤" display-directive="show:lazy">
              <n-scrollbar style="flex: 1 1 0;max-height: 100%">
                <StepsPanel/>
              </n-scrollbar>
              <n-scrollbar style="flex: 0 0 120px;max-height: 100%">
                <n-anchor :show-rail="false" :show-background="true" :bound="114" style="margin: 10px 0">
                  <n-anchor-link v-for="(col, index) in nowCaseData?.caseSteps"
                                 :href="'#'+('动作 ' + (index + 1))">
                    <template #title>
                      <div style="display: flex; align-items: center">
                        <div style="white-space: pre-line; line-height: 1.4; display: flex">
                          <div style="flex: 0 0 25px">
                            {{ (index + 1) + '. ' }}
                          </div>
                          <div style="flex: 1 1 0">
                            {{ (actionsSelectOption.find(o => o.value == col?.action)?.chs) }}
                          </div>
                        </div>
                      </div>
                    </template>
                  </n-anchor-link>
                </n-anchor>
              </n-scrollbar>
            </n-tab-pane>
            <n-tab-pane style="box-sizing: border-box; padding: 5px; height: 100%; overflow: hidden" name="report" tab="执行日志" display-directive="show:lazy">
              <RobotTestLog/>
            </n-tab-pane>
          </n-tabs>
        </n-flex>
      </n-layout>
      <n-layout-footer class="footer" position="absolute" :inverted="inverted" bordered
                       style="height: 64px; padding: 0 20px; display: flex; align-items: center; gap: 20px">
        <n-statistic tabular-nums style="--n-value-font-size: 16px">
          <n-number-animation :from="0" :to="footerStatisticCaseNum"/>
          <template #suffix>
            条用例
          </template>
        </n-statistic>
        <n-statistic tabular-nums style="--n-value-font-size: 16px">
          <n-number-animation :from="0" :to="footerStatisticStepNum"/>
          <template #suffix>
            个动作
          </template>
        </n-statistic>
        <FooterCaseLogStatistic/>
      </n-layout-footer>
    </n-layout>
  </div>
</template>

<style scoped>
#Test {
  position: absolute;
  width: 100%;
  height: 100%;

  box-sizing: border-box;

  color: white;
}

#Content {
  width: 100%;
  height: 100%;
}

.fight-main-tabs {
  width: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* 固定主 Tab 导航宽度，避免执行日志 pane 懒加载后整行标题偏移；不用 animated 避免切换时导航错位 */
.fight-main-tabs :deep(> .n-tabs-nav) {
  flex-shrink: 0;
}

.fight-main-tabs :deep(> .n-tabs-nav .n-tabs-wrapper) {
  width: 100%;
}

/* 仅约束主 Tab 直接子级内容区，避免影响执行日志内嵌 Tab 的切换 */
.fight-main-tabs :deep(> .n-tabs-pane-wrapper) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.footer :deep(.n-statistic .n-statistic-value) {
  margin-top: 0;
}
</style>