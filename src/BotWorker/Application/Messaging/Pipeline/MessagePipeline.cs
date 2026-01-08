using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Bots.BotMessages;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 消息处理管道执行�?    /// </summary>
    public class MessagePipeline
    {
        private readonly List<IMiddleware> _middlewares = new();
        private readonly IServiceProvider _serviceProvider;

        public MessagePipeline(IServiceProvider serviceProvider)
        {
            _serviceProvider = serviceProvider;
        }

        /// <summary>
        /// 添加中间件实�?        /// </summary>
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

            int index = 0;

            // 递归定义 next 委托
            async Task<bool> Next()
            {
                if (index < _middlewares.Count)
                {
                    var middleware = _middlewares[index++];
                    return await middleware.InvokeAsync(context, Next);
                }
                return true; // 管道执行完毕
            }

            return await Next();
        }
    }
}


