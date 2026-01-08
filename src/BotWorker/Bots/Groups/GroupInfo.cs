using System.Text;
using System.Text.RegularExpressions;
using BotWorker.Bots.Groups;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Entries
{
    public partial class GroupInfo : MetaDataGuid<GroupInfo>
    {
        public override string TableName => "Group";
        public override string KeyField => "Id";

        public const long groupMin = 990000000000;
        [DbIgnore]
        public string GroupOpenId { get; set; } = string.Empty;
        [DbIgnore]
        public long TargetGroup { get; set; }
        public string GroupName { get; set; } = string.Empty;
        [DbIgnore]
        public bool IsValid { get; set; }
        [DbIgnore]
        public bool IsProxy { get; set; }
        public string GroupMemo { get; set; } = string.Empty;
        [DbIgnore]
        public long GroupOwner { get; set; }
        public string GroupOwnerName { get; set; } = string.Empty;
        public string GroupOwnerNickname { get; set; } = string.Empty;
        public int GroupType { get; set; }
        [DbIgnore]
        public long RobotOwner { get; set; }
        public string RobotOwnerName { get; set; } = string.Empty;
        public string WelcomeMessage { get; set; } = string.Empty;
        public int GroupState { get; set; }
        [DbIgnore]
        public long BotUin { get; set; }
        public string BotName { get; set; } = string.Empty;
        public DateTime LastDate { get; set; }
        [DbIgnore]
        public int IsInGame { get; set; }
        public bool IsOpen { get; set; }
        public int UseRight { get; set; }
        public int TeachRight { get; set; } //教学权限 1：所有人；2：管理员；3：白名单；4：主人
        public int AdminRight { get; set; } //管理权限 1：所有人；2：管理员；3：白名单；4：主人
        [DbIgnore]
        public DateTime QuietTime { get; set; }
        public bool IsCloseManager { get; set; }
        public int IsAcceptNewMember { get; set; }
        [DbIgnore]
        public string CloseRegex { get; set; } = string.Empty;
        public string RegexRequestJoin { get; set; } = string.Empty;
        public string RejectMessage { get; set; } = string.Empty;
        public bool IsWelcomeHint { get; set; }
        public bool IsExitHint { get; set; }
        public bool IsKickHint { get; set; }
        public bool IsChangeHint { get; set; }
        public bool IsRightHint { get; set; } 
        public bool IsCloudBlack { get; set; }
        public int IsCloudAnswer { get; set; }
        public bool IsRequirePrefix { get; set; }
        public bool IsSz84 { get; set; }
        public bool IsWarn { get; set; }
        public bool IsBlackExit { get; set; } //退群拉黑
        public bool IsBlackKick { get; set; } //踢人拉黑
        public bool IsBlackShare { get; set; } //分享拉黑
        public bool IsChangeEnter { get; set; }
        public bool IsMuteEnter { get; set; }
        public bool IsChangeMessage { get; set; }
        [DbIgnore]
        public bool IsSaveRecord { get; set; }
        [DbIgnore]
        public bool IsPause { get; set; }
        public string RecallKeyword { get; set; } = string.Empty;
        public string WarnKeyword { get; set; } = string.Empty;
        public string MuteKeyword { get; set; } = string.Empty;
        public string KickKeyword { get; set; } = string.Empty;
        public string BlackKeyword { get; set; } = string.Empty;
        public int MuteEnterCount { get; set; }
        public int MuteKeywordCount { get; set; }
        public int KickCount { get; set; }
        public int BlackCount { get; set; }
        public long ParentGroup { get; set; }
        public string CardNamePrefixBoy { get; set; } = string.Empty;
        public string CardNamePrefixGirl { get; set; } = string.Empty;
        public string CardNamePrefixManager { get; set; } = string.Empty;
        [DbIgnore]
        public string LastAnswer { get; set; } = string.Empty;
        [DbIgnore]
        public string LastChengyu { get; set; } = string.Empty;
        [DbIgnore]
        public DateTime LastChengyuDate { get; set; }
        [DbIgnore]
        public DateTime TrialStartDate { get; set; }
        [DbIgnore]
        public DateTime TrialEndDate { get; set; }
        [DbIgnore]
        public DateTime LastExitHintDate { get; set; }
        [DbIgnore]
        public string BlockRes { get; set; } = string.Empty;
        [DbIgnore]
        public int BlockType { get; set; }        
        public int BlockMin { get; set; }
        [DbIgnore]
        public int BlockFee { get; set; }
        [DbIgnore]
        public Guid GroupGuid { get; set; }
        public bool IsBlock { get; set; }
        public bool IsWhite { get; set; }
        public string CityName { get; set; } = string.Empty;
        public bool IsMuteRefresh { get; set; }
        public int MuteRefreshCount { get; set; }
        public bool IsProp { get; set; }
        public bool IsPet { get; set; }
        public bool IsBlackRefresh { get; set; }
        public string FansName { get; set; } = string.Empty;
        public bool IsConfirmNew { get; set; }
        public string WhiteKeyword { get; set; } = string.Empty;
        public string CreditKeyword { get; set; } = string.Empty;
        public bool IsCredit { get; set; }
        public bool IsPowerOn { get; set; }
        public bool IsHintClose { get; set; }
        
        public int RecallTime { get; set; }
        public bool IsInvite { get; set; }
        public int InviteCredit { get; set; }
        public bool IsReplyImage { get; set; }
        public bool IsReplyRecall { get; set; }
        public bool IsVoiceReply { get; set; }
        public string VoiceId { get; set; } = string.Empty;
        public bool IsAI { get; set; }
        public string SystemPrompt { get; set; } = string.Empty;
        public bool IsOwnerPay { get; set; }
        public int ContextCount { get; set; }
        public bool IsMultAI { get; set; }
        public bool IsAutoSignin { get; set; }
        public bool IsUseKnowledgebase { get; set; }
        [DbIgnore]
        public DateTime InsertDate { get; set; }   
        public bool IsSendHelpInfo { get; set; }
        public bool IsRecall { get; set; }
        public bool IsCreditSystem { get; set; }

        public static long GetSourceGroupId(long botUin, long groupId)
        {
            return GetWhere("Id", $"BotUin = {botUin} and TargetGroup = {groupId}").AsLong();
        }

        //本群积分
        public static bool GetIsCredit(long groupId)
        {
            return groupId != 0 && GetBool("IsCredit", groupId);
        }
        
        // 关机
        public static int SetPowerOff(long groupId)
        {
            return SetValue("IsPowerOn", false, groupId);
        }
 
        /// 开机
        public static int SetPowerOn(long groupId)
        {
            return SetValue("IsPowerOn", true, groupId);
        }

        // 是否开机
        public static bool GetPowerOn(long groupId, string groupName = "")
        {
            return GetBool("IsPowerOn", groupId);
        }

        // 是否关机
        public static bool IsPowerOff(long groupId)
        {
            return !GetPowerOn(groupId);
        }
       
        // 判断该群是否还可以体验
        public static bool IsCanTrial(long groupId)
        {
            if (GroupVip.IsVipOnce(groupId))
                return false;

            //体验超过180天可再次体验一次
            int days = GetInt("ABS(DATEDIFF(DAY, GETDATE(), TrialStartDate))", groupId);
            if (days >= 180)
            {
                int trialDays = 7;
                Update($"IsValid = 1, TrialStartDate = GETDATE(), TrialEndDate = GETDATE() + {trialDays}", groupId);
                return true;
            }
            return GetIsValid(groupId);
        }   

        // 取消体验
        public static int SetInvalid(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0)
        {
            Append(groupId, groupName, BotInfo.BotUinDef, BotInfo.BotNameDef, groupOwner, robotOwner);
            if (GroupVip.IsVip(groupId))
                return -1;
            else
                return SetValue("IsValid", false, groupId);
        }

        public static int SetHintDate(long groupId)
        {
            return SetNow("LastExitHintDate", groupId);
        }    

        // 管理员白名单
        public static bool GetIsWhite(long groupId)
        {
            return GetBool("IsWhite", groupId);
        }

        // 群链开关状态
        public static string GetIsBlockRes(long groupId)
        {
            return GetIsBlock(groupId) ? "已开启" : "已关闭";
        }

        // 群链开关
        public static bool GetIsBlock(long groupId)
        {
            return GetBool("IsBlock", groupId);
        }

        // 是否开启
        public static int GetIsOpen(long groupId)
        {
            return GetInt("IsOpen", groupId);
        }        

        // 提示语间隔秒数 包含欢迎语、退群提示、改名提示等信息
        public static int GetLastHintTime(long groupId)
        {
            return GetInt("ABS(DATEDIFF(SECOND, LastExitHintDate, GETDATE()))", groupId);
        }

        // 云问答
        public static int CloudAnswer(long groupId)
        {
            return GetInt("IsCloudAnswer", groupId);
        }

        ///云问答内容
        public static string CloudAnswerRes(long groupId)
        {
            List<string> answers = ["闭嘴", "本群", "官方", "话痨", "终极", "AI"];
            int index = CloudAnswer(groupId);
            if (index >= 0 && index < answers.Count)
                return answers[index];
            else
                return string.Empty;
        }

        /// 退群拉黑
        public static bool GetIsBlackExit(long groupId)
        {
            return GetBool("IsBlackExit", groupId);
        }

        // 被踢拉黑
        public static bool GetIsBlackKick(long groupId)
        {
            return GetBool("IsBlackKick", groupId);
        }

        // 关闭的功能
        public static string GetClosedFunc(long groupId)
        {
            string res = QueryRes($"SELECT CmdName FROM {BotCmd.FullName}", "{0}\n");
            string closeRegex = GetValue("CloseRegex", groupId);
            if (closeRegex.IsNull())  return "";

            StringBuilder sb = new("\n已关闭：");
            foreach (Match match in res.Matches(@"(?<CmdName>" + closeRegex.Replace(" ", "|").Trim() + ")"))
            {
                string cmdName = match.Groups["CmdName"].Value;
                sb.Append($"{cmdName} ");
            }

            return sb.ToString();
        }

        // 关闭的功能 regex
        public static string GetClosedRegex(long groupId)
        {
            string res = GetValue("CloseRegex", groupId);
            if (res != "")
                res = @"^[#＃﹟]{0,1}(?<cmd>(" + res.Trim().Replace(" ", "|") + @"))[+]*(?<cmdPara>[\s\S]*)";
            return res;
        }

        // 退群提示
        public static bool GetIsExitHint(long groupId)
        {
            return GetBool("IsExitHint", groupId);
        }

        // 被踢提示
        public static bool GetIsKickHint(long groupId)
        {
            return GetBool("IsKickHint", groupId);
        }

        // 命令前缀
        public static bool GetIsRequirePrefix(long groupId)
        {
            return GetBool("IsRequirePrefix", groupId);
        }


        // 是否自动审核加群 状态
        public static string GetJoinRes(long groupId)
        {
            int joinRes = GetInt("IsAcceptNewmember", groupId);
            return joinRes switch
            {
                0 => "拒绝",
                1 => "同意",
                2 => "忽略",
                _ => "未设置",
            };
        }

        // 机器人管理权限 状态
        public static string GetAdminRightRes(long groupId)
        {
            int adminRight = GetInt("AdminRight", groupId);
            return adminRight switch
            {
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "主人",
            };
        }

        /// 机器人使用权限 状态
        public static string GetRightRes(long groupId)
        {
            return GetIsOpen(groupId) switch
            {
                1 => "所有人",
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "已关闭",
            };
        }

        // 教学权限
        public static string GetTeachRightRes(long groupId)
        {
            return GetInt("TeachRight", groupId) switch
            {
                1 => "所有人",
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "",
            };
        }

        // 设置机器人开关状态
        public static int SetIsOpen(bool isOpen, long groupId)
        {
            return SetValue("IsOpen", isOpen, groupId);
        }

        // 是否发送欢迎语
        public static string GetWelcomeRes(long groupId)
        {
            return GetBool("IsWelcomeHint", groupId) ? "发送" : "不发送";
        }

        // 机器人主人名称
        public static string GetRobotOwnerName(long groupId, BotWorker.Common.Data.BotData.Platform botType = BotWorker.Common.Data.BotData.Platform.NapCat)
        {
            string res = GetValue("isnull(RobotOwnerName,'')", groupId);
            if (res == "")
            {
                res = GetRobotOwner(groupId).ToString();
                res = ((int)botType).In(2,3) ? res : $"[@:{res}]";
            }
            return res;
        }

        // 群主
        public static long GetGroupOwner(long groupId, long def = 0)
        {
            return GetDef("GroupOwner", groupId, def);
        }

        // 机器人主人
        public static long GetRobotOwner(long groupId, long def = 0)
        {
            return GetDef("RobotOwner", groupId, def);
        }

        public static int SetInGame(int isInGame, long groupId)
        {
            return SetValueSync("IsInGame", isInGame, groupId);
        }

        //是否群机器人主人
        public static bool IsOwner(long groupId, long userId)
        {
            return userId == GetRobotOwner(groupId);
        }

        // 开始成语接龙游戏
        public static int StartCyGame(int isInGame, string lastCy, long groupId)
        {
            SetValueSync("IsInGame", isInGame, groupId);
            return SetValueSync("LastChengyu", lastCy, groupId);
        }

        public static bool GetIsValid(long groupId)
        {
            return GetBool("IsValid", groupId);
        } 

        // 宠物系统是否开启
        public static bool GetIsPet(long groupId)
        {
            return GetBool("IsPet", groupId);
        }

        // 添加新群
        public static int Append(long group, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "")
        {
            return Exists(group)
                ? UpdateGroup(group, name, selfId, groupOwner, robotOwner) 
                : Insert([
                            new Cov("Id", group),
                            new Cov("GroupOpenid", openid),
                            new Cov("GroupName", name),
                            new Cov("GroupOwner", groupOwner),
                            new Cov("RobotOwner", robotOwner),
                            new Cov("BotUin", selfId),
                        ]);
        }

        public static int UpdateGroup(long group, string name, long selfId, long groupOwner = 0, long robotOwner = 0)
        {
            string udpName = name.IsNull() ? "" : $"GroupName = {name.Quotes()}, ";
            string udpGroupOwner = groupOwner == 0 || GetGroupOwner(group) != 0 || GroupVip.IsVip(group) ? "" : $"GroupOwner = {groupOwner},";
            string udpRobotOwner = robotOwner == 0 || GetRobotOwner(group) != 0 || GroupVip.IsVip(group) ? "" : $"RobotOwner = {robotOwner},";
            string udpRobot = $"BotUin = {selfId},";
            return UpdateNoCache($"{udpName} {udpGroupOwner} {udpRobotOwner} {udpRobot} LastDate = GETDATE() ", group);            
        }

        public static int SetRobotOwner(long groupId, long groupOwner)
        {            
            return SetValue("RobotOwner", groupOwner, groupId);
        }
    }
}
