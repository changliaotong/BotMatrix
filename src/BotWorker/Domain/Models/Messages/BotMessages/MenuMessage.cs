using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        //💎金币 👢踢我 　猜拳    礼物 活动 私链 群链 📜对联 👤升级 👤客服  农历 流水 🔁换群 🚪秒进群 🤔猜谜 📊发言榜 🎮逗你玩  
        public static string GetMenuRes()
        {
            return $"📜发送关键词使用指令~\r\n" +
                @"📌常用功能
📜菜单 🗓️签到 ☁️天气 📖说明书
🎁抽奖 🀄️三公 🤑梭哈 🎲猜大小

🔍查询
💎积分 🏦存分 💸取分 🏆积分榜
🛒买分 💵卖分 🎁打赏 🎈领积分
⚡算力 💳余额 🔄续费 💗亲密度

🎲趣味互动
🌅早安 ☀️午安 🌙晚安 💖爱群主
🔮运势 🎴抽签 🔑解签 🌈彩虹屁
📚成语 🧩接龙 😂笑话 👻鬼故事
🌍翻译 🎵点歌 ⏰报时 ⏳倒计时
🧮计算 ✊猜拳 🤔猜谜 👪粉丝团

🛡️群管功能
👥本群 🗣️话唠 🔥终极 🤖智能体
🤫闭嘴 🧠调校 👽变身 👋欢迎语
🤬脏话 📢广告 🔗网址 🚫敏感词
⛔禁言 👢踢出 ⚫拉黑 📛黑名单
💬刷屏 ↩️撤回 🧹清屏 📃白名单
🏷️头衔 🧾名片 🔤前缀 🤐禁言我

⚙️系统控制
🖥️后台 ⚙️设置 👑主人 🔀换主人
✅开启 ❌关闭 🗣️发言 📝提示词";
        }

        public static string GetMenuSimple()
        {
            return $"菜单";                
        }
    }
}
