# Android 数据加载方案（设计 + 待决策）

> rain-qa-func Android 端"数据喂不进"瓶颈的解决方案设计。功能可用性的真瓶颈(非代码兼容性)。
> 相关:[Android-前端适配.md](Android-前端适配.md)(数据供给课题) | [Android-自动更新.md](Android-自动更新.md) | [Android-运行时调试.md](Android-运行时调试.md)

## 一、现状（调研结论 2026-07-10）

### Wails Android OpenFileDialog 实际工作（纠正旧误判）
MainActivity 已有完整 SAF 桥(`MainActivity.java:419`):
- `ACTION_OPEN_DOCUMENT` + `CATEGORY_OPENABLE` + `EXTRA_ALLOW_MULTIPLE` → 系统文件选择器
- `onActivityResult`(`:433`)→ `copyUriToCache`(`:475`)复制到 app cache → 返回文件路径
- **OpenFileDialog(选文件)✅ 工作**;之前 docs"未实现 onShowFileChooser"是误判(旧实测/版本)

### 关键限制
- `ACTION_OPEN_DOCUMENT` 只选**文件**,不选**目录**
- rain-qa-func 需选**目录**(Excel/资源/cases 目录,多文件)→ PathConfigInput selectFolder(CanChooseDirectories)在 Android 不触发或降级

### 数据依赖梳理
| 功能 | 数据 | 来源 |
|---|---|---|
| function-test | cases(用例 JSON) | `cases/fight_cases` **14M**(89 用例,仓库,可 embed);`cases/excel_cases` 150K;`cases/proto_cases` 48K |
| excel-test | Excel 配表(.xlsx) | `/d/work/config` 1022M(太大,需精简示例或 A 导入) |
| **game(基础)** | **resources .bytes(189 文件)** | **rain-robot 源码 `project/xcard/xcard_excel/resources` ⚠️仅 1.4M(非 go mod 包内,需拷仓库)** |
| hero-wiki/activity-wiki | Excel + 历史 JSON | config + 历史 JSON |
| proto-test | 协议文件 | proto |

**game resources 仅 1.4M**(纠正初估"几十 MB")。不在 rain-robot go mod 包内(mod 只打包 .go),在源码目录。**全量 embed 完全可行**(APK 增 1.4M)。版本匹配 go.mod `v0.0.0-20260702143628-7799a7e3edee`(2026-07-02)。

**game resources 是 function-test/excel-test/wiki 的共同基础**(用例执行/规则/Wiki 对比都依赖 game 数据加载)。无 game resources,功能页面"活但跑不动"。

### FuncCaseConfig 关键字段(`backend/pkg/types.go:4`)
- `JsonsDir`(jsons_dir)—— cases 用例目录(用例树加载路径)
- `ExcelResourcesDir`(excel_resources_dir)—— game resources .bytes 目录

## 二、方案对比（优先级 C→A→B→D,用户定）

### C. 内置示例数据(go:embed)—— MVP,让页面"活"
- 做法:打包示例数据进 APK,"加载示例"按钮释放到私有目录 + 配置指向
- 优:零配置,用户开箱即用(演示/基础验证)
- 劣:不解决用户**真实数据**;APK 增大(embed 体积);示例需维护
- **待决策点**(关键):
  1. **game resources 源**:rain-robot .bytes(外部 go mod,不能直接 go:embed go mod 文件)→ 需拷到仓库 `backend/pkg/exampledata/`。拷全量还是精简示例(几武将)?
  2. **版本匹配**:.bytes 版本须匹配 go.mod rain-robot proto(见 LRN-20260710-002)。embed 的 .bytes 锁哪个版本?
  3. **APK 体积接受度**:全量 resources(~几十 MB)+ cases(14M)+ 示例 Excel → APK 增 ~50M+(当前 60M → 110M+)。可接受?
  4. **示例范围**:仅 game resources(基础,让 game 加载)?还是 + function-test cases + excel-test 示例?

### A. zip 打包 + OpenFileDialog 选 zip —— 真实数据导入(用 Wails 现成桥)
- 做法:用户把数据目录→zip,应用 OpenFileDialog(SAF,已工作)选 zip,Go 解压到私有目录,配置指向
- 优:**用 Wails 现成 OpenFileDialog**(选文件),无需改 Java;真实数据;用户可控
- 劣:用户需先打包 zip(桌面 zip 工具);解压大 zip 耗时;zip 结构需约定(目录布局)
- 流程:用户桌面 `zip -r excel.zip config/` → Android 选 zip → Go 解压到 `files/imported/<hash>/` → 配置指向
- **系统干净**:`files/imported/` 每次导入先清旧(同功能单实例),卸载随 app 私有目录清

### B. 自加选目录(ACTION_OPEN_DOCUMENT_TREE)—— 最直观
- 做法:加 Java 桥 `pickDirectory()`(ACTION_OPEN_DOCUMENT_TREE) + content URI 树遍历复制 + Wails 暴露
- 优:用户直接选目录(最直观,免 zip)
- 劣:**需加 Java + Wails 桥 + content URI 树遍历**(DocumentFile 递归,慢);takePersistableUriPermission 权限;复杂度高
- 适用:A(zip)体验不够好时的升级

### D. 外部存储权限(MANAGE_EXTERNAL_STORAGE)—— 最简但权限敏感
- 做法:申请 MANAGE_EXTERNAL_STORAGE,读 /sdcard,用户放数据到 /sdcard/rain-qa-func/
- 优:用户放数据即用,免导入;实现最简(改路径 + 权限)
- 劣:**权限敏感**(Android 11+ Play 禁,自分发 OK 但用户体验差,系统警告"可访问所有文件");审核风险
- 适用:内部工具 + 用户接受权限

## 三、推荐路径（C → A,B/D 视反馈）

1. **C 先做**(MVP):让页面活,零配置演示。**范围待定**(见上待决策点,尤其 game resources)
2. **A 真实数据**:C 之后,zip 导入真实数据(用 Wails 现成 OpenFileDialog)
3. **B/D 视用户反馈**:若 zip 打包麻烦→B(选目录);若愿接受权限→D(/sdcard)

## 四、系统干净约束对数据加载的影响（CLAUDE.md）

- **示例数据可重置**(C):"加载示例"先清旧 `files/example-data/` 再释放(不累积)
- **导入数据单实例**(A/B):`files/imported/<功能>/` 每次导入先清旧,同功能不堆叠
- **下载 APK 安装后清理**:DownloadApk 下载的 APK,installApk 成功后删(自动更新不积旧 APK)—— **待补**(当前 DownloadApk 保留)
- **wails-picker cache 清理**:OpenFileDialog 复制的 cache 文件,用后清
- **配置切换清理**:路径变更时,旧数据(示例/导入)可选清理

## 五、C 实施计划（✅ 已完成 2026-07-13）

**独立包组织**(用户要求方便推翻删除):
- `backend/pkg/exampledata/`(embed.go + service.go + embed/resources 1.4M + embed/fight_cases 10 个)
- `frontend/src/pages/settings/components/example-data-card.vue`(独立组件)
- `frontend/bindings/.../exampledata/*.ts`(3 文件,手写 — alpha2.117 generate 只产 .js,项目用 .ts)
- `cmd/rain-qa-func/wails.go`(+1 import +1 RegisterService)

**推翻删除步骤**(整体清理,无残留):
1. 删 `backend/pkg/exampledata/` 包
2. 去 `cmd/rain-qa-func/wails.go` 的 exampledata import + RegisterService(2 行)
3. 去 `frontend/src/pages/settings/index.vue` 的 import + `<ExampleDataCard/>`(2 行)
4. 删 `frontend/src/pages/settings/components/example-data-card.vue`
5. 删 `frontend/bindings/.../exampledata/`(3 .ts)

**系统干净**:LoadExampleData 先 RemoveAll 旧 `example-data/` 再释放(多次加载不累积)。
**版本**:versionCode 5 已上传 itsnot.fun(2026-07-13),APK 60M→63M(增 ~3M)。

## 六、进度

- [x] 调研 Wails Android 文件能力(OpenFileDialog 选文件工作,选目录不)
- [x] 数据依赖梳理(game resources 是基础 + 外部)
- [x] C/A/B/D 方案对比 + 思考
- [x] **C 实施**(function-test 内置示例,独立包,versionCode 5 上传)
- [ ] 用户测试 C(真机 versionCode 5:设置→加载示例数据→战斗测试页)
- [ ] A 实施(zip 导入,C 后)
- [ ] B/D 视反馈
