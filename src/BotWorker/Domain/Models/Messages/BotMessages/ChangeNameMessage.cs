using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{

        // 改名
        public async Task<string> GetChangeName()
        {
            IsCancelProxy = true;

            if (IsPublic)
                return "公众号无此功能";

            if (!IsGroup)
                return "改名功能仅限群内使用";

            if (!IsGroupBound)
                return "仅官机不能改名，需配合小号机器人使用";

            if (SelfPerm == 2)            
                return IsRealProxy ? $"升机器人({SelfId})为管理才能改名" : "升我为管理才能改名";            

            GetPrefix();

            if (CmdPara.IsMatch(Regexs.UserPara))
            {
                Answer = SetupPrivate(true);
                if (Answer != "")
                    return Answer;

                foreach (Match match in CmdPara.Matches(Regexs.UserPara))
                {
                    var targetUin = match.Groups["UserId"].Value.AsLong();
                    var targetName = match.Groups["cmdPara"].Value.Trim();

                    Answer += $"✅ [@:{targetUin}]({targetUin}) 群名片将改为：{targetName.ReplaceInvalid() ?? "空（显示为QQ昵称）"}";
                    
                    var (prefixBoy, prefixGirl, prefixAdmin) = GetPrefix();
                    await ChangeNameAsync(SelfId, GroupId, targetUin, targetName, prefixBoy, prefixGirl, prefixAdmin);
                }
            }
            else
            {     
                var (prefixBoy, prefixGirl, prefixAdmin) = GetPrefix();
                await ChangeNameAsync(SelfId, GroupId, UserId, CmdPara.Trim(), prefixBoy, prefixGirl, prefixAdmin);
            }

            Answer += $"✅ [@:{UserId}]({UserId}) 群名片将改为：{CmdPara.ReplaceInvalid() ?? "空（显示为QQ昵称）"}";
            IsSend = GroupInfo.GetBool("IsChangeHint", GroupId);

            return Answer;
        }

        public (string prefixBoy, string prefixGirl, string prefixAdmin) GetPrefix()
        {            
            return (Group.CardNamePrefixBoy, Group.CardNamePrefixGirl, Group.CardNamePrefixManager);
        }

        // 一键改名
        public async Task<string> GetChangeNameAllAsync()
        {
            IsCancelProxy = true;

            if (IsPublic)
                return "公众号无此功能";

            if (!IsGroup)
                return "改名功能仅限群内使用";

            if (!IsGroupBound)
                return "仅官机不能改名，需配合小号机器人使用";

            if (SelfPerm == 2)
                return IsRealProxy ? $"升机器人({SelfId})为管理才能改名" : "升我为管理才能改名";

            if (!GroupVip.IsYearVIP(GroupId))
                return YearOnlyMsg;

            var (prefixBoy, prefixGirl, prefixAdmin) = GetPrefix();

            //暂时关闭提示 需要特殊处理
            if (Group.IsChangeEnter)
                GroupInfo.SetValue("IsChangeHint", false, GroupId);

            await ChangeNameAllAsync(SelfId, GroupId, prefixBoy, prefixGirl, prefixAdmin);

            //恢复改名提示
            if (Group.IsChangeEnter)
                GroupInfo.SetValue("IsChangeHint", true, GroupId);

            return Answer;
        }
}
