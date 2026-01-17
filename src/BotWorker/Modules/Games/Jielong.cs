using System.Reflection;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.jielong",
        Name = "成语接龙",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "趣味成语接龙游戏，答对奖励积分，答错扣除积分",
        Category = "Games"
    )]
    public class JielongPlugin : IPlugin
    {
        private readonly IJielongService _jielongService;

        public JielongPlugin(IJielongService jielongService)
        {
            _jielongService = jielongService;
        }

        public BotPluginAttribute Metadata => GetType().GetCustomAttribute<BotPluginAttribute>()!;

        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(new SkillCapability("成语接龙", ["接龙", "jl"]), HandleJielongAsync);
            // 注册消息处理事件，用于在游戏中直接接龙
            await robot.RegisterEventAsync("message", HandleUserMessageAsync);
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleJielongAsync(IPluginContext ctx, string[] args)
        {
            var cmdPara = args.Length > 0 ? string.Join(" ", args) : "";
            return await _jielongService.GetJielongResAsync(ctx, cmdPara);
        }

        private async Task HandleUserMessageAsync(IPluginContext ctx)
        {
            if (ctx.GroupId == null) return;
            
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId);

            // 如果在游戏中，且消息看起来像成语（4个字）或者是强制接龙指令
            if (await _jielongService.InGameAsync(groupId, userId))
            {
                var msg = ctx.RawMessage.Trim();
                if (msg.Length == 4 || msg.StartsWith("接龙") || msg.StartsWith("jl"))
                {
                    var cmdPara = msg.StartsWith("接龙") ? msg[2..].Trim() : (msg.StartsWith("jl") ? msg[2..].Trim() : msg);
                    var res = await _jielongService.GetJielongResAsync(ctx, cmdPara);
                    if (!string.IsNullOrEmpty(res))
                    {
                        await ctx.ReplyAsync(res);
                    }
                }
            }
        }
    }

    public class Jielong
    {
        public long Id { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public string UserName { get; set; } = "";
        public string chengyu { get; set; } = "";
        public int GameNo { get; set; }
        public int Credit { get; set; }
        public DateTime InsertDate { get; set; } = DateTime.Now;
    }
}
