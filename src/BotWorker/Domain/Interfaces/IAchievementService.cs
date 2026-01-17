using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IAchievementService
    {
        /// <summary>
        /// 报告指标并检查成就
        /// </summary>
        /// <param name="userId">用户ID</param>
        /// <param name="key">指标键</param>
        /// <param name="delta">变化量或绝对值</param>
        /// <param name="isAbsolute">是否为绝对值</param>
        /// <returns>新解锁的成就名称列表</returns>
        Task<List<string>> ReportMetricAsync(string userId, string key, double delta, bool isAbsolute = false);
    }
}
