namespace BotWorker.Modules.AI.Interfaces
{
    public interface IDevWorkflowManager
    {
        /// <summary>
        /// 启动自动化开发任务
        /// </summary>
        /// <param name="requirementDoc">原始需求文档</param>
        /// <param name="projectPath">项目生成路径</param>
        Task<bool> StartDevProjectAsync(string requirementDoc, string projectPath);
    }
}
