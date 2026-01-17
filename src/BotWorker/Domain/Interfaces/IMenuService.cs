using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Interfaces
{
    public interface IMenuService
    {
        Task<string> HandleCommandAsync(IPluginContext ctx, string[] args);
        void BuildDynamicMenuTree();
        Task<string> GetRankingsDisplayAsync();
        string GetMonitorDisplay();
    }
}
