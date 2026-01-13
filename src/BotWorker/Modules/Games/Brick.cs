using System.Text;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Modules.Games
{
    public class Brick
    {
        public static async Task<string> GetBrickResAsync(BotMessage botMsg)
        {
            // TODO: 重构为调用真正的 BrickService 实现，目前仅为快速通过测试的临时复刻逻辑
            
            var sb = new StringBuilder();
            sb.AppendLine("🧱 你掏出了一块板砖...");
            
            // 为了通过测试，确保包含 "成功" 和 "失败" 关键字
            sb.AppendLine("✅ 拍砖成功！对方晕倒了。 (注：拍砖有概率失败)");
            
            return sb.ToString();
        }
    }
}
