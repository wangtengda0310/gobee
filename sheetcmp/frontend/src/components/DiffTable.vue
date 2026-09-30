<script setup>
// DiffTable: 单侧表格渲染 (差异高亮 + 待写回标记 + 点击差异格发起同步)。
// props.result: CompareResult; props.side: "left"|"right"; props.rows: 过滤后的行
// props.pending: Set<"row:col"> 该侧待写回格; emit sync-cell(row, col) = 此格取对侧值
const props = defineProps({
  result: { type: Object, required: true },
  side: { type: String, required: true },
  rows: { type: Array, required: true },
  pending: { type: Set, default: () => new Set() },
});
const emit = defineEmits(["sync-cell"]);

const other = props.side === "left" ? "right" : "left";
const rowNumKey = props.side === "left" ? "leftRow" : "rightRow";

// 差异格索引: col → CellView
function diffAt(row, col) {
  for (const d of row.diffs || []) if (d.col === col) return d;
  return null;
}
function cellCount(row) {
  return Math.max(row.left?.length || 0, row.right?.length || 0);
}
function cellAt(row, i) {
  const arr = row[props.side];
  return arr && i < arr.length ? arr[i] : "";
}
function pendingKey(row, col) {
  return row[rowNumKey] + ":" + col;
}
</script>

<template>
  <table>
    <thead>
      <tr>
        <th class="rownum">#</th>
        <th v-for="(h, i) in result.headers" :key="i">{{ h || "C" + (i + 1) }}</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="(row, ri) in rows"
        :key="ri"
        :class="{
          'row-left-only': row.leftRow && !row.rightRow,
          'row-right-only': !row.leftRow && row.rightRow,
        }"
      >
        <td class="rownum">{{ row[rowNumKey] || "—" }}</td>
        <td
          v-for="c in cellCount(row)"
          :key="c"
          :class="{
            'cell-modified': diffAt(row, c - 1)?.kind === 'modified',
            'cell-left': diffAt(row, c - 1)?.kind === 'left',
            'cell-right': diffAt(row, c - 1)?.kind === 'right',
            'cell-formula': diffAt(row, c - 1)?.formulaDiffers,
            'cell-pending': pending.has(pendingKey(row, c)),
            'cell-syncable': !!diffAt(row, c - 1) && !!row[rowNumKey],
          }"
          :title="diffAt(row, c - 1) && row[rowNumKey] ? '点击: 此格取对侧值' : ''"
          @click="diffAt(row, c - 1) && row[rowNumKey] && emit('sync-cell', row, c - 1)"
        >
          {{ cellAt(row, c - 1) }}<span v-if="pending.has(pendingKey(row, c))" class="pend-mark">✓</span>
        </td>
      </tr>
    </tbody>
  </table>
</template>
