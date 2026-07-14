# function-test E2E 测试

战斗测试模块的端到端测试，对应后端包 `pkg/function-test/`、前端页面 `/Test`。

## 文件索引

| 文件 | 覆盖范围 |
|------|----------|
| `function-test.spec.ts` | Header、左侧树面板、Tab 面板、Footer、主 Tab 导航稳定性 |
| `function-test-steps-layout.spec.ts` | 用例步骤 Tab：动作/断言卡片标题行布局、序号连续性 |
| `function-test-tooltip-and-divider.spec.ts` | 下拉分组分割线+染色（卡牌 牌堆灰/座位身份色）、应用智能描述按钮 tooltip、技能下拉选项 tooltip、资产断言「卡」下拉分组（hover 文字区触发；map 未加载时 skip 兜底） |

## 运行方式

```bash
npx playwright test function-test/
```
