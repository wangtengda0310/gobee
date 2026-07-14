# settings E2E 测试

设置页及相关功能，对应后端包 `pkg/settings/`、前端页面 `/Settings`。

## 文件索引

| 文件 | 覆盖范围 |
|------|----------|
| `settings.spec.ts` | 设置页主功能 |
| `roadmap.spec.ts` | 路线图抽屉面板（已迁移到设置页内） |
| `p2p-lazy-loading.spec.ts` | P2P/IPFS/WebTorrent 面板动态加载 |
| `debug-layout.spec.ts` | 布局调试 |
| `debug-layout-simple.spec.ts` | 简化布局调试 |

## 运行方式

```bash
npx playwright test settings/
```
