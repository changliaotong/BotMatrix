using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Plugins;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Domain.Entities;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 统计中间件：记录用户发言次数等数据
    /// </summary>
    public class StatisticsMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 1. 发言次数统计
                if (botMsg.GroupId != BotInfo.MonitorGroupUin)
                {
                    // 记录基础统计
                    if (GroupMsgCount.Update(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name) == -1)
                    {
                        // 记录日志但不终止管道
                    }

                    // 接入成就系统指标上报
                    _ = Task.Run(async () => {
                        var unlocks = await BotWorker.Modules.Games.AchievementPlugin.ReportMetricAsync(botMsg.UserId.ToString(), "sys.msg_count", 1);
                        if (unlocks.Count > 0)
                        {
                            // 这里可以考虑通过 Robot 发送成就解锁通知
                        }
                    });
                }
            }

            await next(context);
        }
    }
}


