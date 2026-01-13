# BotMatrix 自动化测试方案

## 1. 背景与目标
为了确保 BotMatrix 系统（BotNexus + BotWorker）的稳定性和功能正确性，本方案利用 Python 编写的虚拟机器人脚本进行自动化集成测试。
目标是实现对现有功能（AI 聊天、系统指令、技能插件）的回归测试，并能快速适配未来新增的功能模块。

## 2. 测试架构
- **测试主体**：`bot_sim.py` (基于 Python `websockets` 库)
- **模拟对象**：OneBot v11 标准机器人
- **交互流程**：
  1. `bot_sim.py` 连接到 `BotNexus` (ws/bots)。
  2. 模拟用户发送消息事件 (PostType: message)。
  3. `BotNexus` 将消息存入 Redis 队列。
  4. `BotWorker` 消费消息并处理逻辑。
  5. `BotWorker` 发起 Action (如 `send_msg`) 回传给 `BotNexus`。
  6. `BotNexus` 转发 Action 到 `bot_sim.py`。
  7. `bot_sim.py` 断言结果并生成报告。

## 3. 测试用例定义 (`test_config.json`)
测试用例采用 JSON 格式配置，支持以下断言类型：
- `expected`: 精确匹配回复内容。
- `expected_contains`: 回复内容包含指定关键字列表。
- `expected_not_empty`: 回复内容不为空。
- `timeout`: 单次测试的超时时间（秒）。

### 核心测试类别：
1. **AI 交互测试**：验证 LLM 接口和 AI 插件是否工作。
2. **系统指令测试**：验证内置指令（如帮助、状态切换）是否正常。
3. **技能插件测试**：验证第三方或动态加载的 Skill 是否匹配成功。
4. **数据库集成测试**：通过指令执行结果验证数据库读写（如黑名单管理）。

## 4. 自动化执行流程
1. **环境准备**：
   - 启动 Redis。
   - 启动 BotNexus (`run.bat`)。
   - 启动 BotWorker (`dotnet run`)。
2. **执行测试**：
   - 运行 `python bot_sim.py`。
3. **结果审计**：
   - 检查生成的 `TEST REPORT`。
   - 若出现 `FAIL`，查看控制台日志中的 `Reason` 和 BotWorker 的实时日志。

## 5. 未来扩展方案
- **多机器人模拟**：在 `server` 配置中增加多实例支持，测试并发处理。
- **性能监控**：在测试报告中记录 `CostTime`，监控响应耗时趋势。
- **CI/CD 集成**：将脚本集成到构建流水线，确保每次提交代码后系统仍可正常编译并运行核心链路。
- **Mock 模式**：为复杂的外部 API（如天气、搜索）增加 Mock 逻辑。

## 6. 当前已知问题与待办
- **数据库截断错误**：观察到 `SendMessage` 表写入时存在字符串截断，需检查字段长度定义。
- **指令响应率**：部分指令受限于环境配置（如权限、群组状态），测试时需确保虚拟机器人具有足够权限。
