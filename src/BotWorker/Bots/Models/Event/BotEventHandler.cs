using BotWorker.Bots.Models.Achievement;

namespace BotWorker.Bots.Models.Event
{
    public class BotEventHandler
    {
        private readonly AchievementService _achievementService;

        public BotEventHandler(AchievementService achievementService)
        {
            _achievementService = achievementService;
        }

        // 用户签到事件
        public async Task UserSignedInAsync(long userId)
        {
            // 触发成就系统签到检测
            await _achievementService.OnUserSignedIn(userId);

            // 这里可以额外做签到逻辑，比如发欢迎消息
        }

        // 用户发言事件
        public async Task UserSentMessageAsync(long userId, string message)
        {
            // 触发发言次数成就检测
            await _achievementService.OnUserSentMessage(userId);

            // 额外处理消息内容，比如指令解析
        }

        // 用户揍群主事件
        public async Task UserPokedBossAsync(long userId)
        {
            // 触发揍群主成就检测
            await _achievementService.OnUserPokedBoss(userId);

            // 额外业务逻辑
        }
    }


}
