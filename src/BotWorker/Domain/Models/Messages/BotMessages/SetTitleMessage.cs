
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public async Task<string> GetSetTitleResAsync()
        {
            IsCancelProxy = true;

            if (CmdPara.IsNull())
                return $"🏷️ 命令格式：\n" +
                       $"我要头衔 + 头衔名称\n" +
                       $"设置头衔 + QQ + 头衔名称";

            return await GetSetTitleAsync(UserId, CmdPara);
        }

        public async Task<string> GetSetTitleAsync(long? qq = null, string? title = null)
        {
            IsCancelProxy = true;

            if (!IsGroup)
                return "头衔功能仅限群内使用";

            if (!IsGuild && SelfPerm != 0)
                return "我不是群主不能设置头衔";

            qq ??= UserId;
            title ??= CmdPara?.Trim();

            if (string.IsNullOrWhiteSpace(title))
                return "头衔不能为空";

            if (qq != UserId && !HaveSetupRight())
                return "你无权限授予他人头衔";          

            await SetTitleAsync(SelfId, RealGroupId, qq ?? 0, title);

            //Answer = $"✅ 好的，立即给你头衔";

            return Answer;
        }
    }
}
