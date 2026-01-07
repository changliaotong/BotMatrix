using Microsoft.EntityFrameworkCore;

namespace sz84.Bots.Plugins
{
    public interface IAppModuleDbContext
    {
        /// <summary>
        /// 允许模块向主上下文注册自己的实体类型映射配置
        /// </summary>
        void RegisterModels(ModelBuilder builder);
    }

}
