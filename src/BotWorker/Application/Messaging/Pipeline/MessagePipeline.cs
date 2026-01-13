using BotWorker.Infrastructure.Communication.OneBot;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 消息处理管道执行器
    /// </summary>
    public class MessagePipeline
    {
        private readonly List<IMiddleware> _middlewares = new();
        private readonly IServiceProvider _serviceProvider;

        public MessagePipeline(IServiceProvider serviceProvider)
        {
            _serviceProvider = serviceProvider;
        }

        /// <summary>
        /// 添加中间件实例
        /// </summary>
        public void Use(IMiddleware middleware)
        {
            _middlewares.Add(middleware);
        }

        /// <summary>
        /// 执行管道
        /// </summary>
        public async Task<bool> ExecuteAsync(BotMessage context)
        {
            if (!_middlewares.Any()) return true;

            var aiService = _serviceProvider.GetRequiredService<IAIService>();
            var i18nService = _serviceProvider.GetRequiredService<II18nService>();

            var logger = _serviceProvider.GetRequiredService<ILogger<MessagePipeline>>();
            var pluginContext = new PluginContext(
                new BotMessageEvent(context),
                context.Platform,
                context.SelfId.ToString(),
                aiService,
                i18nService,
                logger,
                context.User,
                context.Group,
                null, // Member property seems missing in BotMessage
                context.SelfInfo,
                async msg => { context.Answer = msg; await context.SendMessageAsync(); },
                async (title, artist, jumpUrl, coverUrl, audioUrl) => { await context.SendMusicAsync(title, artist, jumpUrl, coverUrl, audioUrl); }
            );

            int index = 0;

            // 递归定义 next 委托
            async Task Next(IPluginContext ctx)
            {
                if (index < _middlewares.Count)
                {
                    var middleware = _middlewares[index++];
                    // logger.LogInformation("[Pipeline] Step {Step}/{Total}: Executing {MiddlewareName} for message {MessageId}", 
                    //    index, _middlewares.Count, middleware.GetType().Name, ctx.EventId);
                    
                    try
                    {
                        await middleware.InvokeAsync(ctx, Next);
                    }
                    catch (Exception ex)
                    {
                        logger.LogError(ex, "[Pipeline] Error in middleware {MiddlewareName} for message {MessageId}", 
                            middleware.GetType().Name, ctx.EventId);
                        throw;
                    }

                    // logger.LogInformation("[Pipeline] Step {Step}/{Total}: Completed {MiddlewareName} for message {MessageId}", 
                    //    index, _middlewares.Count, middleware.GetType().Name, ctx.EventId);
                }
            }

            await Next(pluginContext);

            // 如果管道执行完后有回答且尚未发送，则自动发送
            if (!string.IsNullOrEmpty(context.Answer) && !context.IsSent && context.IsSend)
            {
                // logger.LogInformation("[Pipeline] Auto-sending final answer for message {MessageId}", context.MsgId);
                await context.SendMessageAsync();
            }

            return true;
        }
    }
}


