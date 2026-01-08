using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Interfaces;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Modules.Plugins;
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

            var pluginContext = new PluginContext(
                new BotMessageEvent(context),
                context.Platform,
                context.SelfId.ToString(),
                aiService,
                i18nService,
                context.User,
                context.Group,
                null, // Member property seems missing in BotMessage
                context.SelfInfo,
                async msg => { context.Answer = msg; await context.SendMessageAsync(); }
            );

            int index = 0;

            // 递归定义 next 委托
            async Task Next(IPluginContext ctx)
            {
                if (index < _middlewares.Count)
                {
                    var middleware = _middlewares[index++];
                    await middleware.InvokeAsync(ctx, Next);
                }
            }

            await Next(pluginContext);
            return true;
        }
    }
}


