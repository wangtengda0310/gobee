# function-test E2E 测试

战斗测试模块的端到端测试，对应后端包 `pkg/function-test/`、前端页面 `/Test`。

## 文件索引

| 文件 | 覆盖范围 |
|------|----------|
| `function-test.spec.ts` | Header、左侧树面板、Tab 面板、Footer |

## 运行方式

```bash
npx playwright test function-test/
```
