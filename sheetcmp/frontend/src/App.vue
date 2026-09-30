<script setup>
import { ref, computed } from "vue";
import { Dialogs } from "@wailsio/runtime";
import { Compare } from "../bindings/github.com/wangtengda0310/gobee/sheetcmp/internal/service/compareservice.js";

// ---------- 状态 ----------
const leftPath = ref("");
const rightPath = ref("");
const keyColsText = ref(""); // 关键列输入, 逗号分隔 (0-based)
const result = ref(null);
const error = ref("");
const busy = ref(false);

// 关键列文本 → 索引数组
const keyColumns = computed(() =>
  keyColsText.value
    .split(",")
    .map((s) => parseInt(s.trim(), 10))
    .filter((n) => !Number.isNaN(n))
);

// 表格模式: 每行的差异格索引 (col → kind), 供模板 O(1) 查询
function diffMap(row) {
  const m = {};
  for (const d of row.diffs || []) m[d.col] = d;
  return m;
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
  } catch (e) {
    error.value = String(e);
    result.value = null;
  } finally {
    busy.value = false;
  }
}

// 表格模式列数 (取左右最大)
function cellCount(row) {
  return Math.max(row.left?.length || 0, row.right?.length || 0);
}
function cellAt(arr, i) {
  return arr && i < arr.length ? arr[i] : "";
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
      placeholder="关键列(0-based, 逗号分隔; 留空按行号配对)"
      style="min-width: 220px"
    />
    <button @click="runCompare" :disabled="busy || !leftPath || !rightPath">
      {{ busy ? "比对中…" : "比对" }}
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
    </template>
    <template v-else>
      相同行 <b>{{ result.stats.matchedRows }}</b> ｜ 左独有(删)
      <b>{{ result.stats.leftOnlyRows }}</b> ｜ 右独有(增)
      <b>{{ result.stats.rightOnlyRows }}</b>
    </template>
  </div>

  <!-- 表格模式: 双栏单滚动容器 (天然同步滚动) -->
  <div v-if="result?.kind === 'sheet'" class="panes">
    <div class="pane">
      <h3>◀ {{ result.leftName }}</h3>
      <table>
        <thead>
          <tr>
            <th class="rownum">#</th>
            <th v-for="(h, i) in result.headers" :key="i">{{ h || "C" + (i + 1) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(row, ri) in result.rows"
            :key="ri"
            :class="{ 'row-left-only': row.leftRow && !row.rightRow, 'row-right-only': !row.leftRow && row.rightRow }"
          >
            <td class="rownum">{{ row.leftRow || "—" }}</td>
            <td
              v-for="c in cellCount(row)"
              :key="c"
              :class="{
                'cell-modified': diffMap(row)[c - 1]?.kind === 'modified',
                'cell-left': diffMap(row)[c - 1]?.kind === 'left',
                'cell-formula': diffMap(row)[c - 1]?.formulaDiffers,
              }"
            >
              {{ cellAt(row.left, c - 1) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="pane">
      <h3>{{ result.rightName }} ▶</h3>
      <table>
        <thead>
          <tr>
            <th class="rownum">#</th>
            <th v-for="(h, i) in result.headers" :key="i">{{ h || "C" + (i + 1) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(row, ri) in result.rows"
            :key="ri"
            :class="{ 'row-left-only': row.leftRow && !row.rightRow, 'row-right-only': !row.leftRow && row.rightRow }"
          >
            <td class="rownum">{{ row.rightRow || "—" }}</td>
            <td
              v-for="c in cellCount(row)"
              :key="c"
              :class="{
                'cell-modified': diffMap(row)[c - 1]?.kind === 'modified',
                'cell-right': diffMap(row)[c - 1]?.kind === 'right',
                'cell-formula': diffMap(row)[c - 1]?.formulaDiffers,
              }"
            >
              {{ cellAt(row.right, c - 1) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <!-- 文本模式 -->
  <div v-else-if="result?.kind === 'text'" class="panes">
    <div class="pane">
      <h3>◀ {{ result.leftName }}</h3>
      <pre class="linelist"><span
        v-for="(ln, i) in result.lines"
        :key="i"
        :class="{ 'line-left': ln.kind === 'left' }"
      ><span class="rownum">{{ ln.leftIdx || "" }}</span> {{ ln.text }}
</span></pre>
    </div>
    <div class="pane">
      <h3>{{ result.rightName }} ▶</h3>
      <pre class="linelist"><span
        v-for="(ln, i) in result.lines"
        :key="i"
        :class="{ 'line-right': ln.kind === 'right' }"
      ><span class="rownum">{{ ln.rightIdx || "" }}</span> {{ ln.text }}
</span></pre>
    </div>
  </div>
</template>
