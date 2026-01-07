using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Config;
using Microsoft.Extensions.Options;

namespace BotWorker.Services
{
    public interface IAuthService
    {
        bool IsSuperAdmin(long userId);
        Task<bool> HasPermissionAsync(long userId, long groupId, string permission);
    }

    public class AuthService : IAuthService
    {
        private readonly WorkerConfig _config;
        private readonly HashSet<long> _superAdmins;

        public AuthService(WorkerConfig config)
        {
            _config = config;
            // 假设配置中有超级管理员列表
            _superAdmins = new HashSet<long> { 12345678 }; // 示例，实际应从配置读取
        }

        public bool IsSuperAdmin(long userId)
        {
            return _superAdmins.Contains(userId);
        }

        public async Task<bool> HasPermissionAsync(long userId, long groupId, string permission)
        {
            if (IsSuperAdmin(userId)) return true;
            
            // 基础权限逻辑，后续根据数据库扩展
            return await Task.FromResult(true);
        }
    }
}
