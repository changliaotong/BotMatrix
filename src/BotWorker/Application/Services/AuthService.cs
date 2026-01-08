using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IAuthService
    {
        bool IsSuperAdmin(long userId);
        Task<bool> HasPermissionAsync(long userId, long groupId, string permission);
    }

    public class AuthService : IAuthService
    {
        private readonly HashSet<long> _superAdmins;

        public AuthService()
        {
            // 示例，实际应从配置读取
            _superAdmins = new HashSet<long> { 12345678 }; 
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
