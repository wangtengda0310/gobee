# 快照校验流程（revision 门控）

每轮生成开始时（阶段一 1.0，定位技能之前的前置门控）校验本地快照是否仍与源表一致；不一致则刷新。校验信号是**文档级 `revision`**。

## 关键技术事实（实测）

- `+info` **不返回 revision**，只返回结构（sheets / 行列数 / 合并范围）。
- `+read` 返回**文档级 revision**：整个文档一个版本号，任何 sheet 的任何编辑都会跳变（实测不同 sheet 返回同值）。
- 因此单靠 `+info` **抓不到「改名」**（行数/合并不变、仅文字变）。校验信号必须取 `+read` 的 revision，不能用 `+info`。

## 启动校验流程

```
① +info                          → 拿结构 + 合并范围（检索全程都要用，每轮必拉）
② 一次小 +read（任一 sheet A1:A1）→ 拿文档级 revision
③ 比对 revision 与快照头部记录
     一致   → 所有快照有效，直接用，不刷新
     不一致 → 重读相关 sheet，刷新快照，更新记录的 revision
     取不到 → 重试一次；仍失败则不静默沿用旧快照，如实告知用户「无法校验快照新鲜度」，由用户决定是否在可能过期的快照上继续
```

命令实例：

```bash
lark-cli sheets +info --url "https://ztgame.feishu.cn/sheets/Z9kFs9JWdhqxQ5tt0I9csmytnVg"
lark-cli sheets +read --url "https://ztgame.feishu.cn/sheets/Z9kFs9JWdhqxQ5tt0I9csmytnVg" --range "iwM7X5!A1:A1"
```

> 比对的是**单一文档 revision**（总校验和），一致即全部快照有效，不逐字段比。

## revision 变化时 — 全量内容刷新（硬规则）

> ⚠️ **禁止仅改 revision 号**。revision 变化意味着源数据已改动，每一处写死的数据（行号、模块数量、合并范围、术语坐标）都必须从 live 数据重新派生，不得假设「模块名没变就没事」或凭上次的偏移量推断新位置。
> 流程：重读全部源数据 → 逐项比对 → 重写快照 → 改 revision 号。**只改号不核内容是错误操作**，会导致后续生成使用过期行号、漏模块、坐标错位。

| 快照 | 刷新方式 |
|------|---------|
| [sheet-structure.md](sheet-structure.md) | 重读各 sheet 结构/合并范围，重写；重读牌 sheet B/C 列去重，重建牌类型层级表（笔误归并、不含牌名）；基本术语大类合并范围若有移位也一并更新 |
| [skill-test-point-modules.md](skill-test-point-modules.md) | 重读「**技能的**测试用例生成点」sheet（`BVTn1E`），重建 11 个一级模块清单 |
| [card-test-point-modules.md](card-test-point-modules.md) | 重读「**牌的**测试用例生成点」sheet（`hYAw0h`），重建一级模块清单。PM 增量中，模块数可能频繁变化 |
| [test-point-knowledge.md](test-point-knowledge.md) | 从**两份模块快照**（技能侧 + 牌侧）的模块名重新抽机制关键词 → 去「基本术语」查定义 → 重建。注意基本术语大类范围可能漂移，关键词的源坐标 `cFEl74 C<行号>` 要随之更新 |

刷新后将各快照头部的 `快照 revision` 与 `快照时间` 更新为当前值。**全部四个快照都必须在内容刷新完成后才改 revision，顺序不能反。**

- **不回滚**：校验在阶段一，早于阶段三，刷新即用最新，不需回滚已生成内容。
- **模块清单永不硬编码**：阶段三的一级模块清单/数量永远从对应路径的快照派生（技能路径 → skill-test-point-modules.md；牌路径 → card-test-point-modules.md）。
- **陌生新模块**：刷新后若出现快照规则未覆盖的新模块，生成已知部分用例，并提示用户人工确认新模块规则。

## 现读不入快照的数据

动作词汇表（基本术语三组）、整牌「效果文案」按需查询是**现读数据**——每轮读进内存、不落盘、不走 revision 门控，不在本文件的快照校验范围内。其结构与用法见各自所属文档。

> 本质：快照是加速器，revision 是哨兵。每轮只拉一次 `+info` + 一次小 `+read`，校验通过后全程用快照定位。
