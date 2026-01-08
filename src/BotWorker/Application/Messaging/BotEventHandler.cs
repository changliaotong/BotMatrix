using System.Threading.Tasks;

namespace BotWorker.Application.Messaging
{
    public class BotEventHandler
    {
        public BotEventHandler()
        {
        }

        // 用户签到事件
        public async Task UserSignedInAsync(long userId)
        {
            // 这里可以额外做签到逻辑，比如发欢迎消息
            await Task.CompletedTask;
        }

        // 用户发言事件
        public async Task UserSentMessageAsync(long userId, string message)
        {
            // 额外处理消息内容，比如指令解析
            await Task.CompletedTask;
        }

        // 用户揍群主事件
        public async Task UserPokedBossAsync(long userId)
        {
            // 额外业务逻辑
            await Task.CompletedTask;
        }
    }
}
