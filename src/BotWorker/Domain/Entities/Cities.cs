using System;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Entities
{
    /// <summary>
    /// 中国城市数据
    /// </summary>
    public partial class Cities
    {
        private static ICityRepository? _repository;
        private static ICityRepository Repository => _repository ??= BotMessage.ServiceProvider?.GetRequiredService<ICityRepository>() ?? throw new InvalidOperationException("ICityRepository not registered");
    }
}

