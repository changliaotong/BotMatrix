using System;
using System.Threading.Tasks;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Application.Messaging.Pipeline
{
    public delegate Task RequestDelegate(IPluginContext context);

    /// <summary>
    /// 机器人消息处理中间件接口
    /// </summary>
    public interface IMiddleware
    {
        Task InvokeAsync(IPluginContext context, RequestDelegate next);
    }
}


