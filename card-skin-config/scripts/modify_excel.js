#!/usr/bin/env node
/**
 * 手牌皮肤配置 — Excel 自动修改（ExcelJS，保留格式）
 * 
 * 用法:
 *   node modify_excel.js --skin-name "纶衍太极" --pinyin "guanyantaiji" --level 至臻
 * 
 * 前置: npm install exceljs
 * Excel 文件需放在 workspace/名将杀配置/ 下
 */

const ExcelJS = require('exceljs');
const path = require('path');

// 文件路径：workspace/名将杀配置/（软链接 → Samba）
const WORKSPACE = '/Users/zt-3803045/.openclaw/workspace/名将杀配置';

const LEVEL_MAP = { '精良': 2, '卓越': 3, '至臻': 4 };

// ─── 工具函数 ───

/** 安全获取单元格文本（处理 richText、number、string 等类型） */
function cellText(cell) {
  const v = cell.value;
  if (!v && v !== 0) return '';
  if (typeof v === 'string') return v;
  if (typeof v === 'number') return String(v);
  if (v.richText) return v.richText.map(t => t.text || '').join('');
  return cell.text || '';
}

/** 替换字符串中的拼音部分 */
function replacePinyin(str, oldP, newP) {
  if (!str || oldP === newP) return str;
  if (oldP && str.includes(oldP)) return str.replaceAll(oldP, newP);
  return str;
}

// ─── 主流程 ───

async function main() {
  const args = Object.fromEntries(
    process.argv.slice(2).flatMap((_, i, a) => i % 2 ? [] : [[a[i].replace('--', ''), a[i + 1]]])
  );
  const skinName = args['skin-name'];
  const pinyin = args['pinyin'];
  const level = args['level'];
  const debug = 'debug' in args;

  if (!skinName || !pinyin || !level) {
    console.error('用法: node modify_excel.js --skin-name "名称" --pinyin "pinyin" --level 精良|卓越|至臻 [--debug]');
    process.exit(1);
  }

  const quality = LEVEL_MAP[level];
  if (!quality) { console.error('等级必须是 精良、卓越 或 至臻'); process.exit(1); }

  console.log(`⚙️ 配置: ${skinName} | ${pinyin} | ${level}(品质${quality})\n`);

  // ─── Item.xlsx ───
  const wb1 = new ExcelJS.Workbook();
  await wb1.xlsx.readFile(path.join(WORKSPACE, 'Item.xlsx'));
  const ws1 = wb1.getWorksheet(1);

  let lastRow, lastId = -1;
  ws1.eachRow((row, rn) => {
    if (rn === 1) return;
    if (row.getCell(3).value === 'CardSkin') {
      const id = parseInt(row.getCell(1).value);
      if (!isNaN(id) && id > lastId) { lastId = id; lastRow = rn; }
    }
  });

  const refRow = ws1.getRow(lastRow);
  const refIcon = cellText(refRow.getCell(22));
  const oldP = refIcon.replace(/\.\w+$/, '').match(/([a-z]{4,})$/)?.[1] || '';
  console.log(`📋 Item.xlsx: 最后 CardSkin Row ${lastRow}, ID=${lastId}, 参考拼音="${oldP}"`);

  // 插入行 + 复制值
  ws1.spliceRows(lastRow + 1, 0, []);
  const newRow = ws1.getRow(lastRow + 1);
  for (let c = 1; c <= 40; c++) {
    newRow.getCell(c).value = cellText(refRow.getCell(c));
  }

  const newId = lastId + 1;
  newRow.getCell(1).value = newId;
  newRow.getCell(2).value = skinName;
  newRow.getCell(4).value = quality;
  newRow.getCell(22).value = replacePinyin(refIcon, oldP, pinyin);

  if (!debug) {
    await wb1.xlsx.writeFile(path.join(WORKSPACE, 'Item.xlsx'));
    console.log(`✅ Item: Row ${lastRow + 1}, ID=${newId}, Icon=${newRow.getCell(22).value}`);
  } else {
    console.log(`[DEBUG] Item: Row ${lastRow + 1}, ID=${newId}`);
  }

  // ─── CardSkin_手牌皮肤表.xlsx ───
  const wb2 = new ExcelJS.Workbook();
  await wb2.xlsx.readFile(path.join(WORKSPACE, 'CardSkin_手牌皮肤表.xlsx'));
  const ws2 = wb2.getWorksheet(1);

  let lastSR, lastSId = -1;
  ws2.eachRow((row, rn) => {
    if (rn <= 2) return;
    const id = parseInt(row.getCell(1).value);
    if (!isNaN(id) && id > lastSId) { lastSId = id; lastSR = rn; }
  });

  const refSRow = ws2.getRow(lastSR);
  console.log(`📋 CardSkin: 最后 Row ${lastSR}, ID=${lastSId}`);

  ws2.spliceRows(lastSR + 1, 0, []);
  const newSRow = ws2.getRow(lastSR + 1);
  for (let c = 1; c <= ws2.columnCount; c++) {
    newSRow.getCell(c).value = cellText(refSRow.getCell(c));
  }

  newSRow.getCell(1).value = newId;
  newSRow.getCell(2).value = skinName;
  newSRow.getCell(3).value = null; // 不配卡面图

  for (let c = 4; c <= 11; c++) {
    const ref = cellText(refSRow.getCell(c));
    if (ref) newSRow.getCell(c).value = replacePinyin(ref, oldP, pinyin);
  }

  if (!debug) {
    await wb2.xlsx.writeFile(path.join(WORKSPACE, 'CardSkin_手牌皮肤表.xlsx'));
    console.log(`✅ CardSkin: Row ${lastSR + 1}, ID=${newId}`);
  } else {
    console.log(`[DEBUG] CardSkin: Row ${lastSR + 1}, ID=${newId}`);
  }

  console.log(`\n══════════════════════`);
  console.log(`  ✅ 配置完成`);
  console.log(`  道具id: ${newId} | 名称: ${skinName} | 品质: ${quality}(${level})`);
  console.log(`  ⚠️ 卡面图路径未配置，请手动补充`);
  console.log(`══════════════════════`);
}

main().catch(e => { console.error('❌', e.message); process.exit(1); });
