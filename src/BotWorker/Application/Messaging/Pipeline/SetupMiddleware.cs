using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Plugins;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 设置与管理中间件：处理机器人开关、黑白名单管理、敏感词设置等
    /// </summary>
    public class SetupMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Infrastructure.Communication.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                var message = botMsg.CurrentMessage.Trim();

                // 处理 开机/关机/开启/关闭 指令
                if (message == "开机" || message == "#开机")
                {
                    botMsg.Answer = await botMsg.SetPowerAsync(true);
                    return;
                }
                if (message == "关机" || message == "#关机")
                {
                    botMsg.Answer = await botMsg.SetPowerAsync(false);
                    return;
                }
                if (message == "开启" || message == "#开启")
                {
                    botMsg.Answer = await botMsg.SetOpenAsync(true);
                    return;
                }
                if (message == "关闭" || message == "#关闭")
                {
                    botMsg.Answer = await botMsg.SetOpenAsync(false);
                    return;
                }

                // 处理 拉黑/敏感词等管理逻辑
                if (botMsg.IsAtMe || botMsg.IsCmd)
                {
                    var res = await botMsg.HandleSetupAsync();
                    if (!string.IsNullOrEmpty(res))
                    {
                        botMsg.Answer = res;
                        return;
                    }
                }
            }

            await next(context);
        }
    }
}
