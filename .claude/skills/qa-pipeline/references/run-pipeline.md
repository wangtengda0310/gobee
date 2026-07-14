# 触发流水线执行

使用 `web-access` skill 在蓝鲸流水线页面点击"启动新构建"。

## 操作步骤

1. 打开流水线主页：
   ```bash
   curl -s "http://localhost:3456/new?url=https://devops.devcloud.ztgame.com/projects/xcard/pipelines/<PIPELINE_ID>"
   ```

2. 点击"启动新构建"按钮：
   ```bash
   curl -s -X POST "http://localhost:3456/click?target=<TARGET_ID>" -d 'button.ant-btn-primary'
   ```

3. 修改变量值（如需）：
   - 变量输入框：`.ant-modal input.ant-input.w-full`
   - 需要触发 React input/change 事件才能生效

4. 点击对话框中的"执行"按钮：
   ```bash
   curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d "document.querySelector('.ant-modal button.ant-btn-primary').click()"
   ```

5. 确认 URL 中出现 `buildId=b-xxx` 表示启动成功。

## 参数

- `--pipeline=<PIPELINE_ID>`：流水线 ID
- `--branch=<BRANCH>`：可选，覆盖 GIT 插件分支
- `--skill=<SKILL>`：可选，设置启动变量 `CI_INVOKE_SKILL`
