using Microsoft.EntityFrameworkCore;

namespace BotWorker.Domain.Interfaces
{
    public interface IAppModuleDbContext
    {
        /// <summary>
        /// 允许模块向主上下文注册自己的实体类型映射配置
        /// </summary>
        void RegisterModels(ModelBuilder builder);
    }
}


