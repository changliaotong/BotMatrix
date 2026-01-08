using BotWorker.Bots.Entries;
using BotWorker.Bots.Groups;
using BotWorker.Bots.Public;
using BotWorker.Bots.Users;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Core.Database;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 处理消息
        public async Task HandleMessageAsync()
        {
            // 触发中间件管道执行
            // 管道内部包含：预处理、黑名单、统计、VIP检查、内置指令、插件分发等所有逻辑
            await ExecutePipelineAsync();
        }
        }
    }
}
