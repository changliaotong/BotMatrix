using System.Text.RegularExpressions;
using BotWorker.Bots.Groups;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 禁言逻辑已迁移至 GroupService.MuteMemberAsync
    }
}
