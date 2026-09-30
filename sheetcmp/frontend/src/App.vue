<script setup>
import { ref, computed } from "vue";
import { Dialogs } from "@wailsio/runtime";
import { Compare, Apply } from "../bindings/github.com/wangtengda0310/gobee/sheetcmp/internal/service/compareservice.js";
import DiffTable from "./components/DiffTable.vue";

// ---------- 比对状态 ----------
const leftPath = ref("");
const rightPath = ref("");
const keyColsText = ref("");
const result = ref(null);
const error = ref("");
const busy = ref(false);

// ---------- 同步编辑状态机 (攒批 + 撤销/重做, 保存时一次性写回) ----------
const pendingEdits = ref([]); // 待写回: {target,row,col,value,formula,clearFormula}
const undone = ref([]); // redo 栈
const onlyDiff = ref(false);

const keyColumns = computed(() =>
  keyColsText.value
    .split(",")
    .map((s) => parseInt(s.trim(), 10))
    .filter((n) => !Number.isNaN(n))
);

// 只看差异: 差异格行 + 单侧独有行
const filteredRows = computed(() => {
  if (!result.value || !onlyDiff.value) return result.value?.rows || [];
  return result.value.rows.filter((r) => (r.diffs && r.diffs.length) || !r.leftRow || !r.rightRow);
});

// 各侧待写回格键集 (Set<"rowIdx:colIdx">)
const pendingLeft = computed(() => pendingKeys("left"));
const pendingRight = computed(() => pendingKeys("right"));
function pendingKeys(side) {
  const s = new Set();
  for (const e of pendingEdits.value) {
    if (e.target === side) s.add(e.row + ":" + e.col);
  }
  return s;
}

// ---------- 动作 ----------
async function pick(side) {
  const picked = await Dialogs.OpenFile({
    title: side === "left" ? "选择左文件" : "选择右文件",
    filters: [
      { name: "表格/文本", pattern: "*.xlsx;*.xlsm;*.csv;*.txt;*.json;*.md;*.py;*.go;*.js;*.ts;*.yaml;*.xml" },
      { name: "所有文件", pattern: "*.*" },
    ],
  });
  const path = Array.isArray(picked) ? picked[0] : picked;
  if (!path) return;
  if (side === "left") leftPath.value = path;
  else rightPath.value = path;
}

async function runCompare() {
  if (!leftPath.value || !rightPath.value) return;
  busy.value = true;
  error.value = "";
  try {
    result.value = await Compare({
      leftPath: leftPath.value,
      rightPath: rightPath.value,
      keyColumns: keyColumns.value,
      sheetName: "",
    });
    pendingEdits.value = []; // 新比对丢弃未保存编辑 (写回后才会走到这里)
    undone.value = [];
  } catch (e) {
    error.value = String(e);
    result.value = null;
  } finally {
    busy.value = false;
  }
}

// 点击差异格: 目标侧此格取对侧值 (只写目标侧, 原版语义)
function onSyncCell(side, row, col) {
  const other = side === "left" ? "right" : "left";
  const otherVals = row[other] || [];
  const otherFormulas = row[other === "left" ? "leftFormulas" : "rightFormulas"] || [];
  const edit = {
    target: side,
    row: side === "left" ? row.leftRow : row.rightRow,
    col: col + 1, // DTO 用 1-based 列号
    value: otherVals[col] ?? "",
    formula: otherFormulas[col] || "",
    clearFormula: false,
  };
  // 对侧无公式而目标侧有 → 清除目标公式 (对齐 xlsx 测试固化的语义)
  const myFormulas = side === "left" ? row.leftFormulas : row.rightFormulas;
  if (!edit.formula && myFormulas && myFormulas[col]) edit.clearFormula = true;

  pendingEdits.value.push(edit);
  undone.value = []; // 新操作清空 redo 栈
}

function undo() {
  const e = pendingEdits.value.pop();
  if (e) undone.value.push(e);
}
function redo() {
  const e = undone.value.pop();
  if (e) pendingEdits.value.push(e);
}

// 保存: 按目标文件分组一次性写回, 成功后重新比对刷新视图
async function saveAll() {
  if (!pendingEdits.value.length) return;
  busy.value = true;
  error.value = "";
  try {
    const groups = { left: [], right: [] };
    for (const e of pendingEdits.value) groups[e.target].push(e);
    for (const [target, edits] of Object.entries(groups)) {
      if (!edits.length) continue;
      await Apply({
        targetPath: target === "left" ? leftPath.value : rightPath.value,
        sheetName: "",
        edits: edits.map((e) => ({
          row: e.row, col: e.col, value: e.value, formula: e.formula, clearFormula: e.clearFormula,
        })),
      });
    }
    await runCompare();
  } catch (e) {
    error.value = "写回失败: " + String(e);
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="toolbar">
    <button @click="pick('left')">打开左…</button>
    <input type="text" v-model="leftPath" placeholder="左文件路径" />
    <button @click="pick('right')">打开右…</button>
    <input type="text" v-model="rightPath" placeholder="右文件路径" />
    <input
      type="text"
      v-model="keyColsText"
      placeholder="关键列(0-based, 逗号分隔)"
      style="min-width: 180px"
    />
    <button @click="runCompare" :disabled="busy || !leftPath || !rightPath">
      {{ busy ? "处理中…" : "比对" }}
    </button>
    <label class="hint"><input type="checkbox" v-model="onlyDiff" /> 只看差异</label>
    <button @click="undo" :disabled="!pendingEdits.length">撤销</button>
    <button @click="redo" :disabled="!undone.length">重做</button>
    <button @click="saveAll" :disabled="busy || !pendingEdits.length" class="primary">
      保存写回 ({{ pendingEdits.length }})
    </button>
  </div>

  <div v-if="error" class="error">{{ error }}</div>

  <div v-if="result" class="stats">
    <template v-if="result.kind === 'sheet'">
      配对行 <b>{{ result.stats.matchedRows }}</b> ｜ 左独有行
      <b>{{ result.stats.leftOnlyRows }}</b> ｜ 右独有行
      <b>{{ result.stats.rightOnlyRows }}</b> ｜ 修改格
      <b>{{ result.stats.modifiedCells }}</b> ｜ 左独格
      <b>{{ result.stats.leftOnlyCells }}</b> ｜ 右独格
      <b>{{ result.stats.rightOnlyCells }}</b>
      <span class="hint">　点击差异格 = 此格取对侧值; 保存后写回原文件</span>
    </template>
    <template v-else>
      相同行 <b>{{ result.stats.matchedRows }}</b> ｜ 左独有(删)
      <b>{{ result.stats.leftOnlyRows }}</b> ｜ 右独有(增)
      <b>{{ result.stats.rightOnlyRows }}</b>
      <span class="hint">　文本模式仅比对, 不支持写回</span>
    </template>
  </div>

  <!-- 表格模式: 双栏单滚动容器 (天然同步滚动) -->
  <div v-if="result?.kind === 'sheet'" class="panes">
    <div class="pane">
      <h3>◀ {{ result.leftName }}</h3>
      <DiffTable :result="result" side="left" :rows="filteredRows" :pending="pendingLeft" @sync-cell="(row, col) => onSyncCell('left', row, col)" />
    </div>
    <div class="pane">
      <h3>{{ result.rightName }} ▶</h3>
      <DiffTable :result="result" side="right" :rows="filteredRows" :pending="pendingRight" @sync-cell="(row, col) => onSyncCell('right', row, col)" />
    </div>
  </div>

  <!-- 文本模式 -->
  <div v-else-if="result?.kind === 'text'" class="panes">
    <div class="pane">
      <h3>◀ {{ result.leftName }}</h3>
      <pre class="linelist"><span
        v-for="(ln, i) in result.lines" :key="i"
        :class="{ 'line-left': ln.kind === 'left' }"
      ><span class="rownum">{{ ln.leftIdx || "" }}</span> {{ ln.text }}
</span></pre>
    </div>
    <div class="pane">
      <h3>{{ result.rightName }} ▶</h3>
      <pre class="linelist"><span
        v-for="(ln, i) in result.lines" :key="i"
        :class="{ 'line-right': ln.kind === 'right' }"
      ><span class="rownum">{{ ln.rightIdx || "" }}</span> {{ ln.text }}
</span></pre>
    </div>
  </div>
</template>
