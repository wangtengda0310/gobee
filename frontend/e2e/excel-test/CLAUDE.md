# excel-test E2E 测试

配表测试模块的端到端测试，对应后端包 `pkg/excel-test/`、前端页面 `/Excel`。

## 文件索引

| 文件 | 覆盖范围 |
|------|----------|
| `excel-test.spec.ts` | 主测试套件：加载配置、树形导航、负责人管理、执行日志 |
| `regex-cross-col-rule.spec.ts` | REGEX_CROSS_COL / REGEX_EXTRACT_RANGE 规则配置界面 |
| `filter-condition-row.spec.ts` | FilterConditionRow 公共过滤条件组件（关系链/跨表检查） |
| `chain-reference-warn-before.spec.ts` | 关系链检查预警配置 |

## 运行方式

```bash
npx playwright test excel-test/
```
