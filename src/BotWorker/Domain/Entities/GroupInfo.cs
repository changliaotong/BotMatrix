using System.Text;
using System.Text.RegularExpressions;
using Newtonsoft.Json;
using BotWorker.Common.Data;

namespace BotWorker.Domain.Entities
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
        [JsonIgnore]
        [HighFrequency]
        public DateTime LastDate { get; set; }
        [JsonIgnore]
        [HighFrequency]
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
        [JsonIgnore]
        [HighFrequency]
        public string LastAnswer { get; set; } = string.Empty;
        [JsonIgnore]
        [HighFrequency]
        public string LastChengyu { get; set; } = string.Empty;
        [JsonIgnore]
        [HighFrequency]
        public DateTime LastChengyuDate { get; set; }
        [DbIgnore]
        public DateTime TrialStartDate { get; set; }
        [DbIgnore]
        public DateTime TrialEndDate { get; set; }
        [DbIgnore]
        [JsonIgnore]
        [HighFrequency]
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
            => GetIsCreditAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> GetIsCreditAsync(long groupId)
        {
            return groupId != 0 && await GetBoolAsync("IsCredit", groupId);
        }

        public static bool GetIsPet(long groupId)
            => GetIsPetAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> GetIsPetAsync(long groupId)
        {
            return groupId != 0 && await GetBoolAsync("IsPet", groupId);
        }

        // 关机
        public static int SetPowerOff(long groupId)
            => SetPowerOffAsync(groupId).GetAwaiter().GetResult();

        public static async Task<int> SetPowerOffAsync(long groupId)
        {
            return await SetValueAsync("IsPowerOn", false, groupId);
        }

        /// 开机
        public static int SetPowerOn(long groupId)
            => SetPowerOnAsync(groupId).GetAwaiter().GetResult();

        public static async Task<int> SetPowerOnAsync(long groupId)
        {
            return await SetValueAsync("IsPowerOn", true, groupId);
        }

        // 是否开机
        public static bool GetPowerOn(long groupId, string groupName = "")
            => GetPowerOnAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> GetPowerOnAsync(long groupId)
        {
            return await GetBoolAsync("IsPowerOn", groupId);
        }

        // 是否关机
        public static bool IsPowerOff(long groupId)
            => !GetPowerOn(groupId);
       
        public static long GetGroupOwner(long groupId, long def = 0)
        {
            return GetDef("GroupOwner", groupId, def);
        }

        public static long GetRobotOwner(long groupId, long def = 0)
        {
            return GetDef("RobotOwner", groupId, def);
        }

        public static bool IsOwner(long groupId, long userId)
        {
            return userId == GetRobotOwner(groupId);
        }

        public static bool GetIsValid(long groupId)
        {
            return GetBool("IsValid", groupId);
        }

        public static string GetRobotOwnerName(long groupId)
        {
            string res = GetValue("isnull(RobotOwnerName,'')", groupId);
            if (res == "")
            {
                res = GetRobotOwner(groupId).ToString();
                res = $"[@:{res}]";
            }
            return res;
        }

        public static string GetRobotOwnerName(long groupId, string botName)
        {
            string res = GetValue("isnull(RobotOwnerName,'')", groupId);
            if (res == "")
            {
                res = GetRobotOwner(groupId).ToString();
                res = $"[@:{res}]";
            }
            return res;
        }

        // 判断该群是否还可以体验
        public static bool IsCanTrial(long groupId)
        {
            if (GroupVip.IsVipOnce(groupId))
                return false;

            //体验超过180天可再次体验一次
            int days = GetInt($"ABS({SqlDateDiff("DAY", SqlDateTime, "TrialStartDate")})", groupId);
            if (days >= 180)
            {
                int trialDays = 7;
                Update($"IsValid = 1, TrialStartDate = {SqlDateTime}, TrialEndDate = {SqlDateAdd("day", trialDays, SqlDateTime)}", groupId);
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
            return GetInt($"ABS({SqlDateDiff("SECOND", "LastExitHintDate", SqlDateTime)})", groupId);
        }

        // 云问答
        public static async Task<int> CloudAnswerAsync(long groupId)
        {
            return await GetIntAsync("IsCloudAnswer", groupId);
        }

        ///云问答内容
        public static async Task<string> CloudAnswerResAsync(long groupId)
        {
            List<string> answers = ["闭嘴", "本群", "官方", "话痨", "终极", "AI"];
            int index = await CloudAnswerAsync(groupId);
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
        public static async Task<string> GetJoinResAsync(long groupId)
        {
            int joinRes = await GetIntAsync("IsAcceptNewmember", groupId);
            return joinRes switch
            {
                0 => "拒绝",
                1 => "同意",
                2 => "忽略",
                _ => "未设置",
            };
        }

        public static string GetJoinRes(long groupId) => GetJoinResAsync(groupId).GetAwaiter().GetResult();

        // 系统提示词
        public static async Task<string> GetSystemPromptAsync(long groupId)
        {
            return await GetValueAsync("SystemPrompt", groupId);
        }

        public static string GetSystemPrompt(long groupId) => GetSystemPromptAsync(groupId).GetAwaiter().GetResult();

        // 机器人管理权限 状态
        public static async Task<string> GetAdminRightResAsync(long groupId)
        {
            int adminRight = await GetIntAsync("AdminRight", groupId);
            return adminRight switch
            {
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "主人",
            };
        }

        public static string GetAdminRightRes(long groupId) => GetAdminRightResAsync(groupId).GetAwaiter().GetResult();

        /// 机器人使用权限 状态
        public static async Task<string> GetRightResAsync(long groupId)
        {
            return (await GetIntAsync("IsOpen", groupId)) switch
            {
                1 => "所有人",
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "已关闭",
            };
        }

        public static string GetRightRes(long groupId) => GetRightResAsync(groupId).GetAwaiter().GetResult();

        // 教学权限
        public static async Task<string> GetTeachRightResAsync(long groupId)
        {
            return (await GetIntAsync("TeachRight", groupId)) switch
            {
                1 => "所有人",
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "",
            };
        }

        public static string GetTeachRightRes(long groupId) => GetTeachRightResAsync(groupId).GetAwaiter().GetResult();

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
        public static async Task<string> GetRobotOwnerNameAsync(long groupId, BotData.Platform botType = BotData.Platform.QQ)
        {
            string res = await GetValueAsync(SqlIsNull("RobotOwnerName", "''"), groupId);
            if (res == "")
            {
                res = (await GetRobotOwnerAsync(groupId)).ToString();
                res = ((int)botType).In(2, 3) ? res : $"[@:{res}]";
            }
            return res;
        }

        // 群主
        public static async Task<long> GetGroupOwnerAsync(long groupId, long def = 0)
        {
            return await GetLongAsync("GroupOwner", groupId, def);
        }

        // 机器人主人
        public static async Task<long> GetRobotOwnerAsync(long groupId, long def = 0)
        {
            return await GetLongAsync("RobotOwner", groupId, def);
        }

        public static async Task<int> SetInGameAsync(int isInGame, long groupId)
        {
            return await SetValueAsync("IsInGame", isInGame, groupId);
        }

        //是否群机器人主人
        public static async Task<bool> IsOwnerAsync(long groupId, long userId)
        {
            return userId == await GetRobotOwnerAsync(groupId);
        }

        // 开始成语接龙 game
        public static async Task<int> StartCyGameAsync(int isInGame, string lastCy, long groupId)
        {
            return await UpdateAsync(new
            {
                IsInGame = isInGame,
                LastChengyu = lastCy,
                LastChengyuDate = DateTime.Now
            }, groupId);
        }

        // 获取最后一次成语接龙的时间间隔（分钟）
        public static async Task<int> GetChengyuIdleMinutesAsync(long groupId)
        {
            string sql = $"SELECT ABS({SqlDateDiff("MINUTE", "LastChengyuDate", SqlDateTime)}) FROM {FullName} WHERE Id = {groupId}";
            return await QueryScalarAsync<int>(sql);
        }

        public static async Task<bool> GetIsValidAsync(long groupId)
        {
            return await GetBoolAsync("IsValid", groupId);
        }

        public static async Task<string> GetGroupNameAsync(long groupId)
        {
            return await GetValueAsync("GroupName", groupId);
        }

        public static string GetGroupOwnerNickname(long groupId)
            => GetGroupOwnerNicknameAsync(groupId).GetAwaiter().GetResult();

        public static async Task<string> GetGroupOwnerNicknameAsync(long groupId)
        {
            return await GetValueAsync("GroupOwnerNickname", groupId);
        }

        public static bool GetIsAI(long groupId)
            => GetIsAIAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> GetIsAIAsync(long groupId)
        {
            return await GetBoolAsync("IsAI", groupId);
        }

        public static bool GetIsOwnerPay(long groupId)
            => GetIsOwnerPayAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> GetIsOwnerPayAsync(long groupId)
        {
            return await GetBoolAsync("IsOwnerPay", groupId);
        }

        public static int GetContextCount(long groupId)
            => GetContextCountAsync(groupId).GetAwaiter().GetResult();

        public static async Task<int> GetContextCountAsync(long groupId)
        {
            return await GetIntAsync("ContextCount", groupId);
        }

        public static bool GetIsMultAI(long groupId)
            => GetIsMultAIAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> GetIsMultAIAsync(long groupId)
        {
            return await GetBoolAsync("IsMultAI", groupId);
        }

        public static bool GetIsUseKnowledgebase(long groupId)
            => GetIsUseKnowledgebaseAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> GetIsUseKnowledgebaseAsync(long groupId)
        {
            return await GetBoolAsync("IsUseKnowledgebase", groupId);
        }

        // 添加新群
        public static async Task<int> AppendAsync(long group, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "")
        {
            return await ExistsAsync(group)
                ? await UpdateGroupAsync(group, name, selfId, groupOwner, robotOwner)
                : await InsertAsync(new
                {
                    Id = group,
                    GroupOpenid = openid,
                    GroupName = name,
                    GroupOwner = groupOwner,
                    RobotOwner = robotOwner,
                    BotUin = selfId,
                });
        }

        public static async Task<int> UpdateGroupAsync(long group, string name, long selfId, long groupOwner = 0, long robotOwner = 0)
        {
            var updateFields = new Dictionary<string, object?>();

            if (!name.IsNull()) updateFields["GroupName"] = name;

            if (groupOwner != 0 && GetGroupOwner(group) == 0 && !GroupVip.IsVip(group))
                updateFields["GroupOwner"] = groupOwner;

            if (robotOwner != 0 && GetRobotOwner(group) == 0 && !GroupVip.IsVip(group))
                updateFields["RobotOwner"] = robotOwner;

            updateFields["BotUin"] = selfId;
            updateFields["LastDate"] = DateTime.MinValue; // 自动处理为数据库当前时间

            return await UpdateAsync(updateFields, group);
        }

        // 添加新群
        public static int Append(long group, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "")
        {
            return AppendAsync(group, name, selfId, selfName, groupOwner, robotOwner, openid).GetAwaiter().GetResult();
        }

        public static int UpdateGroup(long group, string name, long selfId, long groupOwner = 0, long robotOwner = 0)
        {
            return UpdateGroupAsync(group, name, selfId, groupOwner, robotOwner).GetAwaiter().GetResult();
        }

        public static int SetRobotOwner(long groupId, long groupOwner)
        {            
            return SetValue("RobotOwner", groupOwner, groupId);
        }
    }
}
