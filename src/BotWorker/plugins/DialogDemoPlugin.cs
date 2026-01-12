namespace BotWorker.Plugins
{
    [BotPlugin(
        Id = "demo.dialog",
        Name = "对话演示插件",
        Description = "展示多轮对话、意图识别和危险操作确认的参考实现",
        Version = "1.0.0",
        Author = "BotMatrix Team"
    )]
    public class DialogDemoPlugin : IPlugin
    {
        private IRobot? _robot;
        private readonly ILogger<DialogDemoPlugin>? _logger;

        public DialogDemoPlugin()
        {
        }

        public DialogDemoPlugin(ILogger<DialogDemoPlugin> logger)
        {
            _logger = logger;
        }

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;

            // 1. 注册一个带有“意图识别”的技能：意见反馈
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "Feedback",
                Description = "收集用户意见反馈",
                Commands = new[] { "/反馈", "反馈" },
                Intents = new List<Intent>
                {
                    new Intent { Name = "FeedbackIntent", Regex = ".*(建议|反馈|吐槽).*" }
                }
            }, HandleFeedback);

            // 2. 注册一个带有“危险确认”的技能：重置数据
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "ResetData",
                Description = "模拟重置用户数据（高危操作）",
                Commands = new[] { "/reset" }
            }, HandleResetData);
        }

        public Task StopAsync() => Task.CompletedTask;

        /// <summary>
        /// 处理意见反馈（多轮对话示例）
        /// </summary>
        private async Task<string> HandleFeedback(IPluginContext ctx, string[] args)
        {
            // 步骤 2：收到反馈内容
            if (ctx.SessionAction == "Feedback" && ctx.SessionStep == "WaitContent")
            {
                var content = ctx.Message;
                // 这里可以存入数据库
                _logger?.LogInformation("收到来自 {User} 的反馈: {Content}", ctx.UserId, content);
                
                await _robot!.Sessions.ClearSessionAsync(ctx.UserId, ctx.GroupId);
                return "✅ 感谢您的反馈！我们已记录。";
            }

            // 步骤 1：发起对话
            await _robot!.Sessions.StartDialogAsync(
                ctx.UserId, ctx.GroupId, "demo.dialog", "Feedback", "WaitContent");
            
            return "📝 请输入您的建议或反馈内容：";
        }

        /// <summary>
        /// 处理数据重置（危险操作确认示例）
        /// </summary>
        private async Task<string> HandleResetData(IPluginContext ctx, string[] args)
        {
            // 状态 B：用户已输入正确的验证码
            if (ctx.IsConfirmed && ctx.SessionAction == "ResetData")
            {
                // 执行真正的重置逻辑
                return "💣 [危险操作] 用户数据已成功重置！";
            }

            // 状态 A：发起确认请求
            var code = await _robot!.Sessions.StartConfirmationAsync(
                ctx.UserId, ctx.GroupId, "demo.dialog", "ResetData");

            return $"⚠️ 您正在尝试重置所有数据，该操作不可逆！\n请输入验证码【{code}】确认执行，或发送“取消”退出。";
        }
    }
}
