using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Services
{
    public interface IImageGenerationService
    {
        /// <summary>
        /// 根据提示词生成图像
        /// </summary>
        /// <param name="prompt">提示词</param>
        /// <param name="refinePrompt">是否进行提示词优化</param>
        /// <returns>生成的图像 URL 或 CQ 码</returns>
        Task<string> GenerateImageAsync(string prompt, bool refinePrompt = false);

        /// <summary>
        /// 优化提示词
        /// </summary>
        /// <param name="prompt">原始提示词</param>
        /// <returns>优化后的提示词</returns>
        Task<string> RefinePromptAsync(string prompt);
    }
}
