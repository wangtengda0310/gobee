<!--
  ExcelCheckLog - Excel 检查日志组件

  显示检查错误区域（Ok=false 的表级错误 + 列级错误），按 Sheet 分组展示。
-->
<script setup lang="ts">
import {checkLog, tableCheckResults, checking} from "../composables/use-excel-check-log";
import {computed, h} from "vue";
import {ColCheckResult, TableCheckResult} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule";
import {DataTableColumns, NCollapse, NCollapseItem, NTag, NScrollbar, NCard, NDataTable} from "naive-ui";

// ========== 数据分组 ==========

// 合并列级和表级检查结果，按 SheetName 分组
const logFormatted = computed(() => {
  const sheetMap = new Map<string, RowData>()

  // 处理列级检查结果
  checkLog.value?.forEach(colResult => {
    const sheetName = colResult?.SheetName ?? '空'
    let row = sheetMap.get(sheetName)
    if (!row && colResult) {
      row = {
        key: -1,
        sheetName: sheetName,
        colResult: [colResult],
        colResultLen: 0,
        tableResults: []
      }
      sheetMap.set(sheetName, row)
    } else if (row && colResult) {
      row.colResult.push(colResult)
    }
  })

  // 处理表级检查结果，按 sheetName 关联到对应的表
  tableCheckResults.value.forEach(tableResult => {
    const sheetName = tableResult.sheetName ?? tableResult.tableName ?? '空'
    let row = sheetMap.get(sheetName)
    if (!row) {
      row = {
        key: -1,
        sheetName: sheetName,
        colResult: [],
        colResultLen: 0,
        tableResults: [tableResult]
      }
      sheetMap.set(sheetName, row)
    } else {
      row.tableResults.push(tableResult)
    }
  })

  // 转为数组并排序
  const sheetResList = Array.from(sheetMap.values())
  sheetResList.sort((a, b) => a.sheetName.localeCompare(b.sheetName))

  // 计算错误数和 key
  sheetResList.forEach((s, i) => {
    s.key = i
    s.colResultLen = s.colResult.reduce((acc, res) => {
      if (!res.Ok) {
        acc += 1
      }
      return acc
    }, 0)
  })
  return sheetResList
})

// 检查错误（包含列级错误或表级错误的表）
const errorResults = computed(() => {
  return logFormatted.value.filter(row =>
    row.colResultLen > 0 || row.tableResults.some(t => t?.ok === false)
  )
})

// ========== 类型定义 ==========

interface RowData {
  key: number
  sheetName: string
  colResult: ColCheckResult[]
  colResultLen: number
  tableResults: TableCheckResult[]
}

// ========== 表格列定义 ==========

function createColumns(): DataTableColumns<RowData> {
  return [
    {
      type: 'selection'
    },
    {
      type: 'expand',
      expandable: rowData => true,
      renderExpand: (rowData) => {
        const children: any[] = []

        // 表级检查错误（如果有）
        const tableErrors = rowData.tableResults.filter(t => !t.ok)
        if (tableErrors.length > 0) {
          const tableResultItems = tableErrors.map(tr => {
            const items = [
              h(NTag, {
                type: 'error',
                size: 'small'
              }, () => '✗'),
              h('span', {style: 'font-weight: bold; min-width: 150px;'}, tr.displayName),
              h('span', {style: 'color: #FF6B6B;'}, tr.reason)
            ]

            // 如果有错误单元格，显示详细错误列表
            if (tr.errCells && tr.errCells.length > 0) {
              const errCellsItems = tr.errCells.map(ec => {
                if (!ec) return null
                const excelRow = ec.ExcelRow > 0 ? ec.ExcelRow : ec.Index + 1
                return h('div', {
                  style: 'padding-left: 20px; color: #FF6B6B; font-size: 12px;'
                }, `行 ${excelRow}: ${ec.Reason}`)
              })
              items.push(h('div', {style: 'width: 100%;'}, [...errCellsItems]))
            }

            return h('div', {
              style: 'display: flex; flex-direction: column; gap: 5px; padding: 8px 0; border-bottom: 1px solid #333;'
            }, items)
          })
          children.push(h('div', {style: 'margin-bottom: 10px;'}, [
            h('div', {style: 'color: #FF6B6B; font-weight: bold; margin-bottom: 8px;'}, '表级检查错误:'),
            ...tableResultItems
          ]))
        }

        // 列级检查结果（如果有错误）
        const errorCols = rowData.colResult.sort((a, b) => {
          if (a.ColIndex != null && b.ColIndex != null) {
            return a.ColIndex - b.ColIndex
          } else {
            return 1
          }
        }).filter(r => !r.Ok).map(r => {

          const errorCells = r.ErrCells.filter(c => c).map((c: any) => {
            const excelRow = (c.ExcelRow > 0) ? c.ExcelRow : ((c.Index != null) ? c.Index + 1 : -1)
            return h('div', {}, [
              r.TableName?.endsWith("_enum.xlsx") ? h('span', {}, `第${excelRow}行,错误原因:${c?.Reason}`) :
                  h('span', {}, `第${excelRow}行,错误原因:${c?.Reason}`),
            ])
          })

          return h(NCollapse, () => [
            h(NCollapseItem, {title: r.ColName || '空'}, () => errorCells),
          ])
        })

        if (errorCols.length > 0) {
          children.push(h('div', {}, [
            h('div', {style: 'color: #FF9F43; font-weight: bold; margin-bottom: 8px;'}, '列级检查错误:'),
            ...errorCols
          ]))
        }

        return h('div', {style: 'padding: 10px;'}, children)
      }
    },
    {
      title: '#',
      key: 'key',
      render: (row, index) => {
        return `${index + 1}`
      }
    },
    {
      title: '表名',
      key: 'sheetName'
    },
    {
      title: '错误列数',
      key: 'colResultLen'
    },
    {
      title: '表级错误',
      key: 'tableErrors',
      render: (row) => {
        const failed = row.tableResults.filter(t => !t.ok).length
        if (failed === 0) return '-'
        return h(NTag, {
          type: 'error',
          size: 'small'
        }, () => `${failed}`)
      }
    },
  ]
}

const columns = createColumns()
const pagination = {
  pageSize: 10
}
</script>

<template>
  <n-scrollbar style="max-height: 100%; flex: 1 1 0;" id="CardsTop">
    <!-- 执行中 - 覆盖所有结果面板 -->
    <n-card v-if="checking" style="margin: 10px;">
      <div style="text-align: center; padding: 40px; color: #666;">
        <span style="font-size: 48px;">⏳</span>
        <p style="margin-top: 16px;">正在执行检查...</p>
      </div>
    </n-card>

    <!-- 检查完成后显示结果 -->
    <template v-else>
      <!-- 检查错误区域（Ok=false 的结果 + 列级错误） -->
      <n-card v-if="errorResults.length > 0" style="margin: 10px; border-left: 4px solid #FF6B6B;">
        <template #header>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-size: 18px;">❌</span>
            <span>检查错误</span>
            <n-tag type="error" size="small">{{ errorResults.length }}</n-tag>
          </div>
        </template>
        <n-data-table
            :columns="columns"
            :data="errorResults"
            :pagination="pagination"
        />
      </n-card>
    </template>
  </n-scrollbar>
</template>

<style scoped>
:deep(.n-card) {
  margin-bottom: 10px;
}

:deep(.n-collapse-item__header-main) {
  width: 100%;
}
</style>
