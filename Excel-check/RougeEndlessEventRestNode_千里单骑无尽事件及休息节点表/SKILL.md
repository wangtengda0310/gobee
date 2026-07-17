---
name: RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表
description: |
  校验名将杀 RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表.xlsx。
  当用户提到 RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表.xlsx、RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表配置检查时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表 — 配置表检查

校验 `RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表.xlsx`（sheet：`无尽事件节点表|RougeEndlessEventNode`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`EventNodeId`
**运行时输入**：用户提供 `RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表.xlsx` 路径。默认不读外联表。

**业务标识列**：`EventNodeId`。

**读表补充**：类型行=1，字段名行=2，数据起始行（0-based）=`4`；连续空 3 行按主键 `EventNodeId` 截断；第 2、3 行皆空列丢弃。

## 脚本

```bash
python "RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表/scripts/check_RougeEndlessEventRestNode_qianlidanqiwujinshijianjixiuxijiedianbiao.py" "<路径>"
python "RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表/scripts/check_RougeEndlessEventRestNode_qianlidanqiwujinshijianjixiuxijiedianbiao.py" "<路径>" --json
```

Issue：`EventNodeId=<EventNodeId> | EventNodeId=<EventNodeId> | <字段> | <说明>`

Agent 向用户汇报时：原样列出脚本输出的每条 Issue 行；禁止用分类汇总表代替（细则见 [Excel-check/SKILL.md](../SKILL.md)「Agent 汇报硬性要求」）。

---

## 通用规则

### 结构化规则（脚本）

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `EventNodeId` |
| S2 | 部分 | 见字段细则 |
| S3 | 否 | — |
| S4 | 部分 | bool / 枚举有值时 |
| S5–S8 | 否 | —（首轮未归纳） |
| S9 | 部分 | 类型行为数组等时 |
| S10 | 否 | — |
| S11 | 部分 | 成对时间字段（若存在） |

主要字段：`EventNodeId`, `EventNodeType`, `Enemys`, `EnemyFloor`, `BattleEndNodes`, `EventSection1`, `EventSectionTitle1`, `EventSectionDesc1`, `EventSection2`, `EventSectionTitle2`, `EventSectionDesc2`, `EventSection3`, `EventSectionTitle3`, `EventSectionDesc3`, `IsStartNode`, `EventNodeHeroList`, `EventNodeName`, `EventNodeDes`

#### 字段细则

- **EventNodeId**：int，不重复（S1）
- 除主键外其余字段默认**可空**；有值时按类型行做格式校验（S9 等，见脚本）
- 若存在 `Name` / `Title` / `SkillName`：非空（S2）

### 语义规则

无（首轮未归纳出需 LLM 的语义规则；后续可按样例增补）。

---

## 独有规则

### 结构化规则（脚本）

经归纳无（首轮仅落地通用结构校验；后续可按样例/业务增补）。

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
用户: 检查 RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表.xlsx
→ python "RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表/scripts/check_RougeEndlessEventRestNode_qianlidanqiwujinshijianjixiuxijiedianbiao.py" "<路径>"
→ 按本文件通用/独有结构化规则输出报告（首轮无语义）
```
