using System.Text.RegularExpressions;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        //禁言
        public async Task<string> GetMuteResAsync()
        {
            IsCancelProxy = true;

            //判断权限
            if (!IsGuild && SelfPerm == 1)
                return "我不是管理无权禁言";

            if (!IsGuild && SelfPerm == 2)
                return "我不是管理无权禁言";

            //判断权限 使用道具卡
            bool isUseProp = false;
            if (!HaveSetupRight())
            {
                //道具系统是否开启
                if (Group.IsProp)
                    return "您无权使用此命令";

                if (!GroupProps.HaveProp(GroupId, UserId, 1))
                    return "您没有【禁言卡】无权禁言";
                else
                    isUseProp = true;
            }

            string regex = CmdName == "禁言" ? Regexs.Mute : Regexs.UnMute;

            //分析命令
            string res = "";
            foreach (Match match in Message.Matches(regex))
            {
                var targetUin = match.Groups["UserId"].Value.AsLong();
                var dissayTime = 0;
                if (CmdName == "取消禁言")
                {
                    res = "✅ 收到！马上取消";
                }
                else
                {
                    dissayTime = match.Groups["time"].Value.AsInt(10);
                    var unit = match.Groups["unit"].Value.Trim();
                    if (isUseProp)
                    {
                        dissayTime = 10 * 60;
                        int i = GroupProps.UseProp(GroupId, UserId, 1, targetUin);
                        if (i == -1)
                            return "使用【禁言卡】道具失败，请稍后重试";
                        else
                            res = "✅ 成功使用【禁言卡】";
                    }
                    else
                    {
                        if (unit.In("分", "m", "M", "分钟", ""))
                            dissayTime *= 60;
                        else if (unit.In("时", "h", "H", "小时"))
                            dissayTime = dissayTime * 60 * 60;
                        else if (unit.In("日", "d", "D", "天"))
                            dissayTime = dissayTime * 60 * 60 * 24;
                        res = "✅ 收到！马上禁言";
                    }
                    
                }

                await MuteAsync(SelfId, RealGroupId, targetUin, dissayTime);
            }
            return res;
        }
    }
}
