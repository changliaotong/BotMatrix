using System.Text;
using System.Text.RegularExpressions;
using Newtonsoft.Json;

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
        public string CreditName => "积分";

        //本群积分
        public static async Task<bool> GetIsCreditAsync(long groupId)
        {
            return groupId != 0 && await GetBoolAsync("IsCredit", groupId);
        }

        public static async Task<bool> GetIsPetAsync(long groupId)
        {
            return groupId != 0 && await GetBoolAsync("IsPet", groupId);
        }

        // 关机
        public static async Task<int> SetPowerOffAsync(long groupId)
        {
            return await SetValueAsync("IsPowerOn", false, groupId);
        }

        /// 开机
        public static async Task<int> SetPowerOnAsync(long groupId)
        {
            return await SetValueAsync("IsPowerOn", true, groupId);
        }

        // 是否开机
        public static async Task<bool> GetPowerOnAsync(long groupId)
        {
            return await GetBoolAsync("IsPowerOn", groupId);
        }

        public static async Task<long> GetGroupOwnerAsync(long groupId, long def = 0)
        {
            return await GetLongAsync("GroupOwner", groupId, def);
        }

        public static int SetRobotOwner(long groupId, long ownerId) => SetRobotOwnerAsync(groupId, ownerId).GetAwaiter().GetResult();

        public static async Task<int> SetRobotOwnerAsync(long groupId, long ownerId)
        {
            return await SetValueAsync("RobotOwner", ownerId, groupId);
        }

        public static long GetRobotOwner(long groupId) => GetRobotOwnerAsync(groupId).GetAwaiter().GetResult();

        public static async Task<long> GetRobotOwnerAsync(long groupId, long def = 0)
        {
            return await GetLongAsync("RobotOwner", groupId, def);
        }

        public static bool IsOwner(long groupId, long userId) => IsOwnerAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<bool> IsOwnerAsync(long groupId, long userId)
        {
            return userId == await GetRobotOwnerAsync(groupId);
        }

        public static bool IsPowerOff(long groupId) => IsPowerOffAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> IsPowerOffAsync(long groupId)
        {
            return !await GetPowerOnAsync(groupId);
        }

        public static async Task<bool> GetIsValidAsync(long groupId)
        {
            return await GetBoolAsync("IsValid", groupId);
        }

        public static async Task<string> GetRobotOwnerNameAsync(long groupId)
        {
            string res = await GetValueAsync(SqlIsNull("RobotOwnerName", "''"), groupId);
            if (res == "")
            {
                res = (await GetRobotOwnerAsync(groupId)).ToString();
                res = $"[@:{res}]";
            }
            return res;
        }

        public static async Task<string> GetRobotOwnerNameAsync(long groupId, string botName)
        {
            string res = await GetValueAsync(SqlIsNull("RobotOwnerName", "''"), groupId);
            if (res == "")
            {
                res = (await GetRobotOwnerAsync(groupId)).ToString();
                res = $"[@:{res}]";
            }
            return res;
        }



        public static bool IsCanTrial(long groupId) => IsCanTrialAsync(groupId).GetAwaiter().GetResult();

        // 判断该群是否还可以体验
        public static async Task<bool> IsCanTrialAsync(long groupId)
        {
            if (await GroupVip.IsVipOnceAsync(groupId))
                return false;

            //体验超过180天可再次体验一次
            int days = await GetIntAsync($"ABS({SqlDateDiff("DAY", SqlDateTime, "TrialStartDate")})", groupId);
            if (days >= 180)
            {
                int trialDays = 7;
                await UpdateAsync($"IsValid = 1, TrialStartDate = {SqlDateTime}, TrialEndDate = {SqlDateAdd("day", trialDays, SqlDateTime)}", groupId);
                return true;
            }
            return await GetIsValidAsync(groupId);
        }

        public static async Task<int> SetInvalidAsync(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0)
        {
            await AppendAsync(groupId, groupName, BotInfo.BotUinDef, BotInfo.BotNameDef, groupOwner, robotOwner);
            if (await GroupVip.IsVipAsync(groupId))
                return -1;
            else
                return await SetValueAsync("IsValid", false, groupId);
        }

        public static async Task<int> SetHintDateAsync(long groupId)
        {
            return await SetNowAsync("LastExitHintDate", groupId);
        }

        public static async Task<bool> GetIsWhiteAsync(long groupId)
        {
            return await GetBoolAsync("IsWhite", groupId);
        }

        public static async Task<string> GetIsBlockResAsync(long groupId)
        {
            return await GetIsBlockAsync(groupId) ? "已开启" : "已关闭";
        }

        public static async Task<bool> GetIsBlockAsync(long groupId)
        {
            return await GetBoolAsync("IsBlock", groupId);
        }

        public static async Task<int> GetIsOpenAsync(long groupId)
        {
            return await GetIntAsync("IsOpen", groupId);
        }

        public static async Task<int> GetLastHintTimeAsync(long groupId)
        {
            return await GetIntAsync($"ABS({SqlDateDiff("SECOND", "LastExitHintDate", SqlDateTime)})", groupId);
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

        public static async Task<bool> GetIsBlackExitAsync(long groupId)
        {
            return await GetBoolAsync("IsBlackExit", groupId);
        }

        public static async Task<bool> GetIsBlackKickAsync(long groupId)
        {
            return await GetBoolAsync("IsBlackKick", groupId);
        }

        public static string GetClosedFunc(long groupId) => GetClosedFuncAsync(groupId).GetAwaiter().GetResult();

        // 关闭的功能
        public static async Task<string> GetClosedFuncAsync(long groupId)
        {
            string res = await QueryResAsync($"SELECT CmdName FROM {BotCmd.FullName}", "{0}\n");
            string closeRegex = await GetValueAsync("CloseRegex", groupId);
            if (closeRegex.IsNull())  return "";

            StringBuilder sb = new("\n已关闭：");
            foreach (Match match in res.Matches(@"(?<CmdName>" + closeRegex.Replace(" ", "|").Trim() + ")"))
            {
                string cmdName = match.Groups["CmdName"].Value;
                sb.Append($"{cmdName} ");
            }

            return sb.ToString();
        }

        public static string GetClosedRegex(long groupId) => GetClosedRegexAsync(groupId).GetAwaiter().GetResult();

        // 关闭的功能 regex
        public static async Task<string> GetClosedRegexAsync(long groupId)
        {
            string res = await GetValueAsync("CloseRegex", groupId);
            if (res != "")
                res = @"^[#＃﹟]{0,1}(?<cmd>(" + res.Trim().Replace(" ", "|") + @"))[+]*(?<cmdPara>[\s\S]*)";
            return res;
        }

        public static async Task<bool> GetIsExitHintAsync(long groupId)
        {
            return await GetBoolAsync("IsExitHint", groupId);
        }

        public static async Task<bool> GetIsKickHintAsync(long groupId)
        {
            return await GetBoolAsync("IsKickHint", groupId);
        }

        public static async Task<bool> GetIsRequirePrefixAsync(long groupId)
        {
            return await GetBoolAsync("IsRequirePrefix", groupId);
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

        // 系统提示词
        public static async Task<string> GetSystemPromptAsync(long groupId)
        {
            return await GetValueAsync("SystemPrompt", groupId);
        }

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

        public static async Task<int> SetInGameAsync(int isInGame, long groupId)
        {
            return await SetValueAsync("IsInGame", isInGame, groupId);
        }

        public static async Task<int> StartCyGameAsync(int state, string lastChengyu, long groupId)
        {
            return await UpdateAsync(new { IsInGame = state, LastChengyu = lastChengyu }, groupId);
        }

        public static async Task<int> StartCyGameAsync(long groupId)
        {
            return await SetValueAsync("IsCyGame", true, groupId);
        }

        public static async Task<int> GetChengyuIdleMinutesAsync(long groupId)
        {
            return await GetIntAsync(SqlDateDiff("MINUTE", "LastChatDate", SqlDateTime), groupId);
        }

        public static int SetPowerOn(long groupId) => SetPowerOnAsync(groupId).GetAwaiter().GetResult();
        public static int SetPowerOff(long groupId) => SetPowerOffAsync(groupId).GetAwaiter().GetResult();
        public static int SetInGame(int isInGame, long groupId) => SetInGameAsync(isInGame, groupId).GetAwaiter().GetResult();
        public static int StartCyGame(int state, string lastChengyu, long groupId) => StartCyGameAsync(state, lastChengyu, groupId).GetAwaiter().GetResult();
        public static int StartCyGame(long groupId) => StartCyGameAsync(groupId).GetAwaiter().GetResult();
        public static int GetChengyuIdleMinutes(long groupId) => GetChengyuIdleMinutesAsync(groupId).GetAwaiter().GetResult();

        public static int SetPowerOn(bool isOpen, long groupId)
        {
            return SetValue("IsPowerOn", isOpen, groupId);
        }

        public static int GetLastHintTime(long groupId) => GetLastHintTimeAsync(groupId).GetAwaiter().GetResult();
        public static int SetHintDate(long groupId) => SetHintDateAsync(groupId).GetAwaiter().GetResult();
        public static int SetInvalid(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0) => SetInvalidAsync(groupId, groupName, groupOwner, robotOwner).GetAwaiter().GetResult();
        public static bool IsVip(long groupId) => GroupVip.IsVipAsync(groupId).GetAwaiter().GetResult();

        // 设置机器人开关状态
        public static int SetIsOpen(bool isOpen, long groupId)
        {
            return SetValue("IsOpen", isOpen, groupId);
        }

        public static string GetWelcomeRes(long groupId) => GetWelcomeResAsync(groupId).GetAwaiter().GetResult();

        public static async Task<string> GetWelcomeResAsync(long groupId)
        {
            return await GetBoolAsync("IsWelcomeHint", groupId) ? "发送" : "不发送";
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

        public static async Task<string> GetGroupNameAsync(long groupId)
        {
            return await GetValueAsync("GroupName", groupId);
        }

        public static async Task<string> GetGroupOwnerNicknameAsync(long groupId)
        {
            return await GetValueAsync("GroupOwnerNickname", groupId);
        }

        public static async Task<bool> GetIsAIAsync(long groupId)
        {
            return await GetBoolAsync("IsAI", groupId);
        }

        public static async Task<bool> GetIsOwnerPayAsync(long groupId)
        {
            return await GetBoolAsync("IsOwnerPay", groupId);
        }

        public static async Task<int> GetContextCountAsync(long groupId)
        {
            return await GetIntAsync("ContextCount", groupId);
        }

        public static async Task<bool> GetIsMultAIAsync(long groupId)
        {
            return await GetBoolAsync("IsMultAI", groupId);
        }

        public static async Task<bool> GetIsUseKnowledgebaseAsync(long groupId)
        {
            return await GetBoolAsync("IsUseKnowledgebase", groupId);
        }

        public static int Append(long group, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "")
            => AppendAsync(group, name, selfId, selfName, groupOwner, robotOwner, openid).GetAwaiter().GetResult();

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

            if (groupOwner != 0 && await GetGroupOwnerAsync(group) == 0 && !await GroupVip.IsVipAsync(group))
                updateFields["GroupOwner"] = groupOwner;

            if (robotOwner != 0 && await GetRobotOwnerAsync(group) == 0 && !await GroupVip.IsVipAsync(group))
                updateFields["RobotOwner"] = robotOwner;

            updateFields["BotUin"] = selfId;
            updateFields["LastDate"] = DateTime.MinValue; // 自动处理为数据库当前时间

            return await UpdateAsync(updateFields, group);
        }

        public static async Task<bool> GetIsNoLogAsync(long groupId) => await GetBoolAsync("IsNoLog", groupId);
        public static async Task<bool> GetIsNoCheckAsync(long groupId) => await GetBoolAsync("IsNoCheck", groupId);
        public static async Task<bool> GetIsHintCloseAsync(long groupId) => await GetBoolAsync("IsHintClose", groupId);
        public static async Task<long> GetSourceGroupIdAsync(long groupId) => await GetLongAsync("SourceGroupId", groupId);
        public static async Task<long> GetSourceGroupIdAsync(long botUin, long groupId) => await GetLongAsync("SourceGroupId", groupId);

        public static long GetSourceGroupId(long groupId) => GetSourceGroupIdAsync(groupId).GetAwaiter().GetResult();
        public static long GetSourceGroupId(long botUin, long groupId) => GetSourceGroupIdAsync(botUin, groupId).GetAwaiter().GetResult();

        public static bool GetBool(string fieldName, long groupId) => GetBoolAsync(fieldName, groupId).GetAwaiter().GetResult();

        public static async Task<int> SetIsOpenAsync(bool isOpen, long groupId)
        {
            return await SetValueAsync("IsOpen", isOpen, groupId);
        }

        public static async Task<string> GetSystemPromptStatusAsync(long groupId)
        {
            string prompt = await GetSystemPromptAsync(groupId);
            if (string.IsNullOrEmpty(prompt)) prompt = "未设置";
            return $"📌 设置系统提示词\n内容：\n{prompt}";
        }

        public static string GetSystemPromptStatus(long groupId) => GetSystemPromptStatusAsync(groupId).GetAwaiter().GetResult();

        public static async Task<string> GetVipResAsync(long groupId)
        {
            string version;
            string res;

            if (await GroupVip.ExistsAsync(groupId))
            {
                if (await GroupVip.IsYearVIPAsync(groupId))
                    version = "年费版";
                else
                    version = "VIP版";
                int valid_days = await GroupVip.RestDaysAsync(groupId);
                if (valid_days >= 1850)
                    res = "『永久版』";
                else
                    res = $"『{version}』有效期：{valid_days}天";
            }
            else
            {
                if (await GroupVip.IsVipOnceAsync(groupId))
                    return "已过期，请及时续费";
                else
                    version = "体验版";
                res = $"『{version}』";
            }

            return res;
        }

        public static string GetVipRes(long groupId) => GetVipResAsync(groupId).GetAwaiter().GetResult();
    }
}
