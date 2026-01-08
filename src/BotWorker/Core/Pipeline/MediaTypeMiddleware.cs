using System.Threading.Tasks;
using BotWorker.Core.Plugin;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 媒体类型中间件：处理图片、文件、视频等非文本消息
    /// </summary>
    public class MediaTypeMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                if (botMsg.IsFile || botMsg.IsVideo || botMsg.IsXml || botMsg.IsJson || 
                    botMsg.IsKeyboard || botMsg.IsLightApp || botMsg.IsLongMsg || 
                    botMsg.IsMarkdown || botMsg.IsStream || botMsg.IsVoice || 
                    botMsg.IsMusic || botMsg.IsPoke)
                {
                    botMsg.Answer = botMsg.HandleOtherMessage();
                    botMsg.Reason += "[非文本]";
                    return; // 处理完毕，不再向下传递
                }
            }

            await next(context);
        }
    }
}
