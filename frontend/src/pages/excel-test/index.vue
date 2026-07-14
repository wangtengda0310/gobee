<!--
  ExcelTestEditor - Excel 配表测试页面

  提供 Excel 配表的加载、配置、检查和日志查看功能
-->
<script setup lang="ts">
import {h, ref} from "vue";
import {activeKey, menuOptions} from "./composables/menu";
import {activeTab as excelActiveTab} from "./composables/func";
import {NIcon, useDialog, useMessage} from "naive-ui";
import {pattern, showIrrelevantNodes} from "@shared/composables/use-tree-search";
import {dataRef, succAndFailSheetNum} from "./composables/use-tree-and-history";
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
} from "./composables/use-tree";
import {
  handleClickOutside,
  handleSelect,
  optionsRef,
  showDropdownRef,
  xRef,
  yRef
} from "@shared/composables/use-tree-dropdown";
import ExcelCheckPanel from "./components/excel-check-panel.vue";
import {MdOptions} from "@vicons/ionicons4";
import {MessageCircle2} from "@vicons/tabler";
import {Catalog} from "@vicons/carbon";
import ExcelCheckLog from "./components/excel-check-log.vue";
import ExcelCheckManager from "./components/excel-check-manager.vue";
import {checkLog} from "./composables/use-excel-check-log";
import {ExcelResourceDir, ExcelCaseDir, saveExcelConfig} from "./composables/option";
import PathConfigInput from "@shared/components/path-config-input/index.vue";

const inverted = ref(false)

// 触控设备(移动端,pointer:coarse)默认折叠左侧 sider 并显示折叠触发器,给右侧让出空间。
// 桌面(pointer:fine)isTouchDevice=false:无 trigger、默认展开,行为完全不变。
const isTouchDevice = typeof window !== 'undefined' && window.matchMedia('(pointer: coarse)').matches

const dialog = useDialog()
const message = useMessage()

// 保存配置
const saveConfig = () => {
  saveExcelConfig().then(() => {
    message.success("配置已保存")
  }).catch(err => {
    message.error("保存配置失败: " + err)
  })
}

</script>

<template>
  <div id="Excel">
    <n-layout position="absolute">
      <n-layout-header style="height: 34px; display: flex; align-items: center; justify-content: space-between" :inverted="inverted" bordered>
        <n-menu mode="horizontal" :inverted="inverted" :options="menuOptions(dialog, message)"
                v-model:value="activeKey"/>
        <n-space align="center" :size="10" :wrap="false" style="padding-right: 12px">
          <PathConfigInput
            v-model:excel-dir="ExcelResourceDir"
            v-model:second-value="ExcelCaseDir"
            excel-label="配表"
            second-label="用例"
            :on-save="saveConfig"
          />
        </n-space>
      </n-layout-header>
      <n-layout position="absolute" style="top: 34px; bottom: 64px" has-sider>
        <n-layout-sider
            bordered
            :show-trigger="isTouchDevice"
            collapse-mode="width"
            :collapsed-width="50"
            :width="240"
            :default-collapsed="isTouchDevice"
            :native-scrollbar="false"
            :inverted="inverted"
            :show-collapsed-content="false"
        >
          <n-flex vertical>
            <div style="position: sticky; top: 0; z-index: 1; background-color: #2b2b2b">
              <n-input v-model:value="pattern" placeholder="搜索"/>
              <div style="display: flex; justify-content: space-around;">
<!--                <n-switch v-model:value="showIrrelevantNodes" :round="false">-->
<!--                  <template #checked>-->
<!--                    仅展示过滤-->
<!--                  </template>-->
<!--                  <template #unchecked>-->
<!--                    展示全部-->
<!--                  </template>-->
<!--                </n-switch>-->
<!--                <n-switch v-model:value="showExcelCheckDesc" :round="false">-->
<!--                  <template #checked>-->
<!--                    显示描述-->
<!--                  </template>-->
<!--                  <template #unchecked>-->
<!--                    隐藏描述-->
<!--                  </template>-->
<!--                </n-switch>-->
              </div>
            </div>
            <div>
              <n-tree
                  style="overflow-x: hidden"
                  block-line
                  draggable
                  :data="dataRef"
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
                  expand-on-click
              />
              <!--右键菜单-->
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
            </div>
          </n-flex>
        </n-layout-sider>
        <n-flex id="Content" style="width: 100%">
          <n-tabs size="small" :bar-width="100" type="line" animated
                  justify-content="space-evenly" style="height: 100%"
                  v-model:value="excelActiveTab"
          >
            <n-tab-pane style="padding: 1px; height: 100%" name="manager" display-directive="show:lazy"
                        :tab="()=> h('div', [h(NIcon,  {style: 'vertical-align: middle;margin-bottom:3px', component:MessageCircle2, size:16}), '负责人'])">
              <ExcelCheckManager/>
            </n-tab-pane>
            <n-tab-pane style="padding: 1px; height: 100%" name="option" display-directive="show:lazy"
                        :tab="()=> h('div', [h(NIcon,  {style: 'vertical-align: middle;margin-bottom:3px', component:MdOptions, size:16}), '用例配置'])">
              <ExcelCheckPanel/>
            </n-tab-pane>
            <n-tab-pane style="box-sizing: border-box; padding: 5px; height: 100%" name="report" display-directive="show:lazy"
                        :tab="()=> h('div', [h(NIcon,  {style: 'vertical-align: middle;margin-bottom:3px', component:Catalog, size:16}), '执行日志'])">
              <ExcelCheckLog/>
            </n-tab-pane>
          </n-tabs>
        </n-flex>
        <n-layout/>
      </n-layout>
      <n-layout-footer class="footer" position="absolute" :inverted="inverted" bordered
                       style="height: 64px; padding: 0 20px; display: flex; align-items: center; gap: 20px">
        <n-statistic tabular-nums style="--n-value-font-size: 16px">
          <n-number-animation :from="0" :to="dataRef.length"/>
          <template #suffix>
            张配表xlsx文件
          </template>
        </n-statistic>
        <n-statistic tabular-nums style="--n-value-font-size: 16px">
          <n-number-animation :from="0" :to="succAndFailSheetNum[0]+succAndFailSheetNum[1]"/>
          <template #suffix>
            张Sheet
          </template>
        </n-statistic>
        <n-statistic tabular-nums style="--n-value-font-size: 16px; --n-value-text-color: #1bf41b">
          <n-number-animation :from="0" :to="succAndFailSheetNum[0]"/>
          <template #suffix>
            张成功Sheet
          </template>
        </n-statistic>
        <n-statistic tabular-nums style="--n-value-font-size: 16px; --n-value-text-color: #ffe95c">
          <n-number-animation :from="0" :to="succAndFailSheetNum[1]"/>
          <template #suffix>
            张错误Sheet
          </template>
        </n-statistic>
        <n-statistic tabular-nums style="--n-value-font-size: 16px; --n-value-text-color: red">
          <n-number-animation :from="0" :to="checkLog?.reduce((acc, cur)=>{
            if (cur?.Ok) return acc
            acc += cur?.ErrCells.length || 0
            return acc
          },0) || -1"/>
          <template #prefix>
            本次检查共有
          </template>
          <template #suffix>
            个错误单元格
          </template>
        </n-statistic>
      </n-layout-footer>
    </n-layout>
  </div>
</template>

<style scoped>
#Excel {
  position: relative;
  width: 100%;
  height: 100%;

  box-sizing: border-box;

  color: white;
}
</style>
