using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.Common.Exts;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 设置与管理中间件：处理机器人开关、黑白名单管理、敏感词设置等
    /// </summary>
    public class SetupMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (!context.IsMessage)
            {
                await next(context);
                return;
            }

            var message = context.RawMessage.Trim();

            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 处理 开启/关闭 指令
                if (message == "开启" || message == "#开启")
                {
                    botMsg.Answer = await botMsg.GetOpenAsync(true);
                    return;
                }
                if (message == "关闭" || message == "#关闭")
                {
                    botMsg.Answer = await botMsg.GetOpenAsync(false);
                    return;
                }

                // 处理 拉黑/敏感词等管理逻辑 (对应原 HandleSetupAsync)
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
