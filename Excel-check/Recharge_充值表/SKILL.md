---
name: Recharge_充值表
description: |
  校验名将杀 Recharge_充值表.xlsx。
  当用户提到 Recharge_充值表.xlsx、充值表检查、充值配置校验时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# Recharge_充值表 — 充值配置表检查

校验 `Recharge_充值表.xlsx`。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**运行时输入**：用户提供 `Recharge_充值表.xlsx` 路径。本表会读取同目录（或 `--map`）下的 `Recharge_映射表.xlsx`，校验充值 `Id` 是否在映射中按端对应。

**业务标识列**：`Name`（商品）；`Platform`（端，同名可多行）。独有规则主要围绕 `Name`，并以 `BaseRelateId` 是否为空区分变体。

## 脚本

```bash
python "Recharge_充值表/scripts/check_Recharge_chongzhibiao.py" "<路径>"
python "Recharge_充值表/scripts/check_Recharge_chongzhibiao.py" "<路径>" --map "<Recharge_映射表路径>"
python "Recharge_充值表/scripts/check_Recharge_chongzhibiao.py" "<路径>" --json
```

Issue 展示列用 `Name`：`Id=<Id> | Name=<Name> | <字段> | <说明>`

Agent 向用户汇报时：原样列出脚本输出的每条 Issue 行；禁止用分类汇总表代替（细则见 [Excel-check/SKILL.md](../SKILL.md)「Agent 汇报硬性要求」）。

---

## 通用规则

### 结构化规则（脚本）

适用公共类型（落到本表）：

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Id` |
| S2 | 是 | `Name`, `RechargeType`, `RelateId`, `ProductId`, `Platform`, `Price` |
| S3 | 否 | — |
| S4 | 是 | `RechargeType`, `Platform`, `LimitType` |
| S5 | 否 | —（档位定值见独有规则） |
| S6 | 否 | —（跨端同名允许部分数值不同） |
| S7 | 是 | `LimitType`↔`LimitCount`；`OnShelfTime`↔`OffShelfTime` |
| S8 | 否 | — |
| S9 | 是 | `ProductId`；`Depend`/`Mutex` |
| S10 | 否 | — |
| S11 | 是 | `OnShelfTime`/`OffShelfTime` |

读表约定：空/`#` Id 跳过；连续空 3 行截断；第 2、3 行皆空列丢弃。

#### 字段细则

- **Id**：int，不重复；且必须在 `Recharge_映射表.xlsx` 中按 `Platform` 落入对应列（见独有规则）  
- **Name / RechargeType / RelateId / ProductId / Platform / Price**：非空  
- **RechargeType**：`MonthlyCard` \| `SeasonPass` \| `ShopGood` \| `LimitGift` \| `PickGoods` \| `Activity` \| `UnSale`  
- **Platform**：`andriod` \| `ios` \| `pc`  
- **ProductId**：无首尾空白/换行；`^qudao_\d+\.\d+$`  
- **Channel**：可空；有值 int  
- **Price**：数值 ≥0（独有锁价见下）  
- **OldPrice / Discount / RechargeMulti / RechargeGroup / GiftGoodID**：可空；有值按 float/int  
- **LimitType / LimitCount**：同空或同有；Type∈{1,2,3,4}  
- **OnShelfTime / OffShelfTime**：同空或同有；`YYYY-MM-DD HH:MM:SS`；上架 ≤ 下架  
- **Depend / Mutex**：可空；`^\d+(,\d+)*$`

### 语义规则

无（不适用 L1–L3）。

---

## 独有规则

### 结构化规则（脚本）

#### 1. 名称以金额开头时锁死 Price

若 `Name` 匹配 `^(\d+)元`，则 `Price` 必须等于该整数（如 `12元限时礼包` → `Price=12`）。

#### 2. 「N两黄金」Price = 描述数 / 10

若 `Name` 匹配 `^(\d+)两黄金$`，则 `Price` 必须等于该整数 ÷ 10（如 `60两黄金` → `Price=6`）。

**例外**：落在 `#支付中心黄金` 分区内的「N两黄金」**不做**本条价格校验。

#### 3. 赛季礼包成套

`Name` 匹配 `S{n}豪华版礼包` / `S{n}典藏版礼包`（`n` 为正整数）：

| 变体 | 判定 | Price | BaseRelateId | Depend |
|------|------|-------|--------------|--------|
| 豪华 | `…豪华版礼包` | 必须 `38` | 应为空 | 应为空 |
| 典藏·升级 | `…典藏版礼包` 且 Base 有值 | 必须 `90` | = 同季 `S{n}豪华版礼包` 的 `RelateId` | = 同季三端豪华充值 Id，顺序 `andriod,pc,ios`（三端升级行写同一串） |
| 典藏·直购 | `…典藏版礼包` 且 Base 为空 | 必须 `128` | 空 | 应为空 |

上架窗与期数：

- `OnShelfTime`：`S2` 及以后必须为 `YYYY-MM-15 00:00:00`；**`S1` 不校具体时刻**，但 `OnShelfTime` **不得为未来时间**（相对检查时刻）
- `OffShelfTime` 必须为 `YYYY-MM-14 23:59:59`（某月 14 日结束；豪华/典藏直购与升级档的截止月可不同，但时刻形态须为此）
- 名称中的 `n`（第几期）按表内各期 **上架年月** 升序推演：最早一期为 `S1`，次之为 `S2`，以此类推；行上的 `n` 必须等于推演结果（上架日仍须为每月 **15 日**，以便推演期数）
- 期数不可重复：同一 `n` 只能对应同一上架年月；同一上架年月只能对应一个 `n`

#### 4. 赛季皮肤礼盒

命中：`Name` 精确等于 `赛季皮肤礼盒`（其它命名变体暂不纳入）。

| 字段 | 规则 |
|------|------|
| `RechargeType` | 必须 `PickGoods` |
| `Price` | 必须 `128` |
| `BaseRelateId` | 必须非空 int（自选皮肤） |
| `LimitType` / `LimitCount` | 必须为 `1` / `1` |
| `Channel` / `Depend` / `Mutex` / `GiftGoodID` | 应为空 |
| `RechargeGroup` | 必须非空 int |

按期（同一 `RelateId`）成组：

- 组内所有行的 `OnShelfTime`、`OffShelfTime` 必须一致
- 组内所有行的 `RechargeGroup` 必须等于该组 **andriod** 行中最小的 `Id`
- 同一 `(RelateId, BaseRelateId)` 在 `andriod` / `pc` / `ios` 上各恰有一行

暂不校：外联 `ShopGoods`/`Item`、一期必配皮肤数量、`第六赛季皮肤礼盒` 等命名变体。

#### 5. 充值 Id ↔ 映射表

每个数据行的 `Id` 必须出现在 `Recharge_映射表.xlsx`（sheet 充值映射）中，且列与 `Platform` 对应：

| Platform | 映射表列（字段名） |
|----------|-------------------|
| `andriod` | `IdAndroid`（充值安卓id） |
| `pc` | `IdPc`（充值Pc-id） |
| `ios` | `IdIos`（充值ios-id） |

默认映射表路径：与充值表同目录的 `Recharge_映射表.xlsx`；可用 `--map` 覆盖。映射表缺失则报错并跳过本条逐行校验。

### 语义规则

经归纳无。

---

## 补充规则时（必须）

按用户要求为本表 **新增/修改规则**前：先对照本文件已有「通用规则」「独有规则」（及对应脚本实现）。

| 情况 | 处理 |
|------|------|
| 与现有规则实质重复 | 先反馈重复点，勿落盘；询问是否保留/合并/取消 |
| 与现有规则冲突 | 先列出冲突双方，停止实现；询问以哪方为准 |
| 无重复且无冲突 | 再写入本文件，并视需要改脚本 |

细则见 [Excel-check/SKILL.md](../SKILL.md)「使用者后续补充」。

---

## 工作流程

```
用户: 检查 Recharge_充值表.xlsx
→ python "Recharge_充值表/scripts/check_Recharge_chongzhibiao.py" "<路径>" [--map "<映射表>"]
→ 按本文件通用/独有结构化规则输出报告（含 Id↔映射表；无语义）
```
