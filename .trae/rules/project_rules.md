1. 代码需要通过编译
1. 编写完代码需要补充对应的单元测试，达到分支覆盖率100%
   代码需要通过单元测试
2. 编写代码需要分析并根据使用的框架决定是否编写集成测试
   代码需要通过集成测试
3. 代码需要通过golangci-lint的检查
4. 分析哪些方法适合使用模糊测试，并编写模糊测试
   代码需要通过模糊测试
5. 代码需要有整体的使用手册
6. 开始新项目时先通过对话确认需求在编写设计文档，根据设计文档编写代码
7. 当前目录中存在多个golang项目，每个项目各自维护go.mod，执行相关命令时注意工作目录以及golang工程结构
8. 当前环境可能默认打开powershell或gitbash或cmd，执行命令前最好加以鉴别以免无效操作
9. 分析代码判断哪些逻辑可能对人类阅读不友好，标注todo 在todo中附上备注文件并在备注文件中使用uml图给出现有逻辑和可能优化的方案