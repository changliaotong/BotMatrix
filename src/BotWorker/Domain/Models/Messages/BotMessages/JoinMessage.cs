using Microsoft.Data.SqlClient;
using sz84.Bots.Entries;
using sz84.Bots.Extensions;
using sz84.Bots.Groups;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;
using sz84.Groups;
using sz84.Bots.Users;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {        
        
        // 是否允许加入群 同意返回“1”，不同意返回：“0:拒绝原因”
        public (int, string) GetRequestJoinGroup()
        {
            if (IsBlack)
                return (0, "黑名单禁入");
            else
            {
                if (GroupId.In(28981482, 81741884, 184705328))
                {
                    return GroupVip.IsClientVip(UserId) ? (1, "") : (0, "非VIP禁入");
                }
                else
                {
                    int res = GroupInfo.GetInt("IsAcceptNewmember", GroupId); // 0-拒绝 1-通过 2-忽略 3-密码验证
                    return res == 3
                        ? Message.IsMatch(GroupInfo.GetValue("RegexRequestJoin", GroupId)) ? (1, "") : (0, "密码错误")
                        : (res, GroupInfo.GetValue("RejectMessage", GroupId));
                }
            }
        } 

        //退群或被踢时减邀请人数
        public void SubInviteCount()
        {
            long InvitorUserId = GroupMember.GetLong("InvitorUserId", GroupId, UserId);
            if (InvitorUserId > 0)
                GroupMember.Plus("InviteExitCount", 1, GroupId, InvitorUserId);
        }

        // 机器人加群成功
        public void GetJoinedRes()
        {
            //处理加入群的信息
            GroupInfo.Append(GroupId, GroupName, SelfId, SelfName, InvitorQQ);

            BotEventLog.Append(SelfId, "加群成功", GroupId, GroupName, SelfId, SelfName);

            Answer = "我来了";

            if (Group.IsValid || IsGuild)
            {
                //自动开机
                GroupInfo.SetPowerOn(GroupId);

                //关闭状态自动开启
                if (Group.IsOpen)
                    GroupInfo.SetValue("IsOpen", true, GroupId);

                //加群后提示设置管理员
                if (!GroupVip.IsVip(GroupId) && Group.IsSz84)
                    Answer = "我来了，设置我为管理开启功能";
            }
            else
            {
                Answer = GroupVip.IsVipOnce(GroupId)
                    ? $"本群机器人已过期，如需继续使用请联系客服续费。客服QQ：{{客服QQ}}"
                    : $"本群机器人已过体验期，如需继续使用请联系客服购买。客服QQ：{{客服QQ}}";
            }
            IsCancelProxy = true;
        }

        // 获取欢迎语
        public string GetWelcomeRes(string para = "")
        {
            if (para != "")
            {
                if (para.IsMatchQQ())
                    TargetUin = para.GetAtUserId();
                else
                    return "";
            }
            else
                TargetUin = UserId;

            // 默认欢迎语列表
            string[] defaultWelcomes =
            [
                "👏 鼓掌欢迎新朋友！呱唧呱唧～ 欢迎加入大家庭！",
                "欢迎欢迎，热烈欢迎～✨ 大家鼓个掌撒个花🌸！",
                "新朋友上线！请大家排队鼓掌👏 欢迎TA闪亮登场～",
                "哟吼～来了位靓仔靓妹！掌声在哪里？🔥",
                "🎊 欢迎新朋友，咱们群今天多了颗闪亮的小星星✨",
                "🥳 欢迎加入！请收下这份来自全群的关爱～",
                "叮咚~ 新成员驾到！全体注意，准备欢迎仪式🎉",
                "🚪刚推开门的你，已经被我们盯上了：呱唧呱唧～🤗",
                "🌟 新人加入啦，整个群都跟着闪亮了起来！",
                "📢 欢迎新同学！请上讲台自我介绍（开玩笑的哈哈）~",
                "我们一直在等你，现在终于等到了！🙌 欢迎加入！",
                "🎈新朋友加入，群主开心到原地转圈圈！",
                "欢迎欢迎～愿你在这里收获欢笑、友谊与快乐！",
                "🎵 欢迎曲已奏响，请新朋友上场，大家鼓掌👏",
                "新朋友上线～请带上好心情一起嗨吧！🎉",
                "🏡 新人进群如归家，欢迎加入这个温暖的大家庭～",
                "请大家掌声欢迎！这位朋友可是大人物，我们发财靠他啦💰🤣",
                "嘿嘿，新人别害羞，我们群超友好，欢迎欢迎～🤩",
                "🐣 一只小萌新破壳而出，大家快来围观欢迎！"
            ];

            string res = Group.WelcomeMessage;

            if (res.IsNull())
                res = defaultWelcomes.RandomOne() ?? "";

            res = SelfInfo.BotType == 8 
                ? $"{res}" 
                : $"{(IsOnebot ? $"[CQ:image,file=https://q1.qlogo.cn/g?b=qq&nk={UserId}&s=100]" : "")}[@:{UserId}] ({UserId})\n{res} " + $"{(Group.IsWelcomeHint ? "" : "\n欢迎语已设置为不发送，开启请发【开启 欢迎语】")}";
            return res;
        }

        // 新成员加入群
        public async Task GetMemberJoinedAsync()
        {
            if (IsBlack && SelfInfo.BotType != 8)
            {
                await KickOutAsync(SelfId, RealGroupId, UserId);
                Answer = "黑名单成员溜进群将被T飞";
                return;
            }

            //欢迎语
            if (Group.IsWelcomeHint)
            {
                IsCancelProxy = true;
                //欢迎语为空、其它机器人、短期大量进群的不发送
                Answer = GetWelcomeRes();
                IsSend = SelfInfo.BotType == 8 || (!BotInfo.IsRobot(UserId) && GroupInfo.GetLastHintTime(GroupId) >= 10);
                if (IsSend) 
                    GroupInfo.SetHintDate(GroupId);
            }

            if (SelfInfo.BotType == 8) return;

            //邀请统计
            await InviteGetCreditAsync();

            //需要管理权限的功能
            if (SelfPerm < UserPerm)
            {
                //进群改名
                if (Group.IsChangeEnter)
                {
                    string prefix = Group.CardNamePrefixBoy;
                    if (prefix != "")                                                                 
                        await ChangeNameAsync(SelfId, RealGroupId, UserId, prefix + Name, prefix, "", "");                    
                }

                //进群禁言                    
                if (Group.IsMuteEnter)
                    await MuteAsync(SelfId, RealGroupId, UserId, Group.MuteEnterCount * 60);
                
                //进群确认
                Answer = await GetConfirmNew();
                if (!Answer.IsNull())
                    await SendMessageAsync();
            }
        }

        // 邀请统计、邀请送分
        public async Task InviteGetCreditAsync()
        {
            if (InvitorQQ > 0)
            {
                try
                {
                    int i = UserInfo.AppendUser(SelfId, GroupId, UserId, Name);
                    int j = UserInfo.AppendUser(SelfId, GroupId, InvitorQQ, InvitorName);
                    if (i >= 0 && j >= 0)
                    {
                        var sql = GroupMember.SqlUpdate("InvitorUserId", InvitorQQ, GroupId, UserId);
                        var sql2 = GroupMember.SqlPlus("InviteCount", 1, GroupId, InvitorQQ);

                        (string, SqlParameter[]) sql3 = ("", []);
                        (string, SqlParameter[]) sql4 = ("", []);

                        if (Group.InviteCredit > 50)
                        {
                            long minusCredit = Group.InviteCredit - 50;
                            long ownerCredit = UserInfo.GetCredit(GroupId, Group.RobotOwner);
                            if (ownerCredit >= minusCredit)
                            {
                                sql3 = UserInfo.SqlAddCredit(SelfId, GroupId, Group.RobotOwner, -minusCredit);
                                sql4 = CreditLog.SqlHistory(SelfId, GroupId, GroupName, Group.RobotOwner, Group.RobotOwnerName, -minusCredit, $"邀人送分:{InvitorQQ}邀请{UserId}");
                            }
                            else
                                Group.InviteCredit = 50;
                        }
                        long creditInvitor = UserInfo.GetCredit(GroupId, InvitorQQ);
                        var sql5 = UserInfo.SqlAddCredit(SelfId, GroupId, InvitorQQ, Group.InviteCredit);
                        var sql6 = CreditLog.SqlHistory(SelfId, GroupId, GroupName, UserId, Name, Group.InviteCredit, $"邀人送分:邀请{UserId}进群{GroupId}");                                              

                        i = ExecTrans(sql, sql2, sql3, sql4, sql5, sql6);
                        if (i != -1)
                        {
                            Answer = $"[@:{InvitorQQ}] 邀请 [@:{UserId}]进群\n累计已邀请{GroupMember.GetInt("InviteCount", GroupId, InvitorQQ)+1}人";
                            Answer += $"\n积分：+{Group.InviteCredit}，累计：{creditInvitor + Group.InviteCredit}";
                        }
                        
                        IsSend = Group.IsInvite;
                        IsCancelProxy = true;
                        await SendMessageAsync();
                        Answer = "";                        
                    }
                }
                catch (Exception ex)
                {
                    DbDebug("InviteGetCredit", ex.Message);
                }
            }
        }
    }
}
