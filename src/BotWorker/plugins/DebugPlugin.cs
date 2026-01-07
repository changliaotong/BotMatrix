using System;
using System.Threading.Tasks;
using BotWorker.Core.Plugin;

namespace BotWorker.Plugins
{
    public class DebugPlugin : IPlugin
    {
        public string Name => "DebugPlugin";
        public string Description => "用于调试上下文信息的插件";

        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "debug",
                Description = "显示当前上下文信息",
                Commands = new[] { ".debug" }
            }, HandleDebug);

            // 注册一个通用事件处理器，例如成员增加事件
            await robot.RegisterEventAsync("EventNoticeGroupIncrease", HandleMemberIncrease);
            await robot.RegisterEventAsync("MemberJoinedEvent", HandleMemberIncrease);
        }

        private async Task HandleMemberIncrease(IPluginContext ctx)
        {
            await ctx.ReplyAsync($"[Debug] 欢迎新成员 {ctx.UserId} 加入群组 {ctx.GroupId}！");
        }

        private async Task<string> HandleDebug(IPluginContext ctx, string[] args)
        {
            var userStr = ctx.User != null ? $"{ctx.User.Name}({ctx.User.Id}, 积分:{ctx.User.Credit})" : "null";
            var groupStr = ctx.Group != null ? $"{ctx.Group.GroupName}({ctx.Group.Id})" : "私聊";
            var memberStr = ctx.Member != null ? $"角色:{ctx.Member.Role}" : "n/a";
            var botStr = ctx.Bot != null ? $"{ctx.Bot.BotName}({ctx.Bot.BotUin})" : "null";

            // 测试 AI 服务
            var aiResponse = await ctx.AI.ChatAsync("你好");

            // 测试主动回复
            await ctx.ReplyAsync("正在获取调试信息...");

            return $"[Debug 信息]\n" +
                   $"用户: {userStr}\n" +
                   $"群组: {groupStr}\n" +
                   $"成员: {memberStr}\n" +
                   $"机器人: {botStr}\n" +
                   $"AI 测试: {aiResponse}\n" +
                   $"语言: {ctx.I18n.GetString("hello")}";
        }
    }
}
