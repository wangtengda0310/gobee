# gobee — 个人项目

本仓库为**个人工具积累**，与任何公司/雇主项目无关。

## 合规红线（必须遵守）

为避免合规审计风险，本仓库**严禁**出现公司相关信息，包括但不限于：

- 公司/雇主名称、域名、邮箱（如 git author 身份）
- 内网 IP 段、内部服务地址（镜像仓/Harbor/内部 API 等）
- 内部仓库名、工单/需求单号、内部系统名称
- 任何能关联到特定公司项目或雇主的措辞

涉及的位置：代码注释、文档/README、commit message、分支名/tag、配置文件、
前端资源（文案/图标/链接）。

## 具体规则

1. **环境相关配置一律参数化**（命令行参数/环境变量/本地不入库的配置文件），
   不写死任何特定环境的地址。
2. **git 提交身份用个人邮箱**（仓库级已配置 `wangtengda0310@users.noreply.github.com`；
   勿让公司 gitconfig 的身份提交本仓库）。
3. **从公司项目搬运代码前先清洗**：域名、内网 IP、单号、内部仓库名、公司专属
   配置全部移除或参数化；已推送的敏感内容需要重写历史 + force-push。
4. 提交前快速自查：

   ```bash
   # 按实际情况替换模式（公司域名 / 内网 IP 段 / 单号 / 内部仓库名）
   git grep -iE "<company-domain>|<internal-ip-prefix>\.|<ticket-pattern>"
   git log --format=%B | grep -iE "<company-domain>|<ticket-pattern>"
   ```
