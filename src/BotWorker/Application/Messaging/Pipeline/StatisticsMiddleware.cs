using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.Common;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 统计中间件：记录用户发言次数等数�?    /// </summary>
    public class StatisticsMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 发言次数统计
                if (botMsg.GroupId != botMsg.BotInfo.MonitorGroupUin)
                {
                    if (Bots.Groups.GroupMsgCount.Update(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name) == -1)
                    {
                        // 记录日志但不终止管道
                        // Logger.Error("更新发言统计数据时出错�?);
                    }
                }
            }

            await next(context);
        }
    }
}


