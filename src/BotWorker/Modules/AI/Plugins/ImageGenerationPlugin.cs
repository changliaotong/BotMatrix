using System.ComponentModel;
using System.Threading.Tasks;
using Microsoft.SemanticKernel;
using BotWorker.Modules.AI.Services;
using BotWorker.Modules.AI.Tools;

namespace BotWorker.Modules.AI.Plugins
{
    public class ImageGenerationPlugin
    {
        private readonly IImageGenerationService _imageService;

        public ImageGenerationPlugin(IImageGenerationService imageService)
        {
            _imageService = imageService;
        }

        [KernelFunction]
        [Description("根据文字描述生成图像。适用于生成海报、配图、插画等场景。")]
        [ToolRisk(ToolRiskLevel.Low, "调用 AI 生成图像")]
        public async Task<string> GenerateImage(
            [Description("对想要生成的图像的详细文字描述")] string prompt,
            [Description("是否自动优化提示词以获得更好的生图效果，默认为 true")] bool refinePrompt = true
        )
        {
            var imageUrl = await _imageService.GenerateImageAsync(prompt, refinePrompt);
            if (string.IsNullOrEmpty(imageUrl))
            {
                return "❌ 图像生成失败，请稍后再试。";
            }
            
            // 返回带 CQ 码的图片，这样机器人可以直接发送
            return $"[CQ:image,file={imageUrl}]";
        }
    }
}
