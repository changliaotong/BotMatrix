namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public async Task<string> GetTurnOnAsync(string cmdName, string cmdPara)
        {
            cmdPara = cmdPara.ToUpper();
            string fieldName = cmdPara switch
            {
                //个人设置
                "闭嘴" => "IsShutup",
                "闭嘴模式" => "IsShutup",
                "默认提示" => "DefaultHint",
                "私链" => "IsBlock",
                "GPT4" => "IsGpt4",

                //群设置
                "积分系统" => "IsCreditSystem",
                "阅后即焚" => "IsRecall",   
                "语音回复" => "IsVoiceReply",
                "撤回回复" => "IsReplyRecall",
                "图片回复" => "IsReplyImage",
                "权限提示" => "IsRightHint",
                "云黑名单" => "IsCloudBlack",
                "管理加白" => "IsWhite",
                "进群确认" => "IsConfirmNew",                
                "欢迎语" => "IsWelcomeHint",
                "道具系统" => "IsProp",
                "宠物系统" => "IsPet",
                "群管系统" => "IsWarn",
                "敏感词" => "IsWarn",
                "敏感词系统" => "IsWarn",
                "改名提示" => "IsChangeHint",
                "进群改名" => "IsChangeEnter",
                "进群禁言" => "IsMuteEnter",
                "退群提示" => "IsExitHint",
                "退群拉黑" => "IsBlackExit",
                "被踢提示" => "IsKickHint",
                "被踢拉黑" => "IsBlackKick",
                "踢出拉黑" => "IsBlackKick",
                "命令前缀" => "IsRequirePrefix",
                "群链" => "IsBlock",                              
                             
                "功能提示" => "IsHintClose",
                "AI" => "IsAI",
                "群主付" => "IsOwnerPay",
                "邀请统计" => "IsInvite",
                "多人互动" => "IsMultAI",
                "自动签到" => "IsAutoSignIn",
                "知识库" => "IsUseKnowledgebase",

                _ => throw new NotImplementedException()
            };

            if (cmdPara == "进群确认")
            {
                if (!GroupVip.IsYearVIP(GroupId) | (cmdName == "开启"))
                    return YearOnlyMsg;
            }

            int isOpen = cmdName == "开启" ? 1 : 0;
            int i = cmdPara.In("闭嘴", "闭嘴模式", "默认提示", "私链", "GPT4")
                ? await UserInfo.SetValueAsync(fieldName, isOpen, UserId)
                : await GroupInfo.SetValueAsync(fieldName, isOpen, GroupId);

            return i == -1 ? RetryMsg : $"✅ {cmdPara}已{cmdName}";
        }

        public string GetTurnOn(string cmdName, string cmdPara)
            => GetTurnOnAsync(cmdName, cmdPara).GetAwaiter().GetResult();

        //全局开启关闭机器人某项功能
        public async Task<string> GetCloseAllAsync()
        {
            string res = "";

            //机器人管理员才能使用此命令
            if (!BotInfo.IsAdmin(SelfId, UserId))
                return "";

            if (CmdPara == "")
                res = "";
            else
            {
                if ((CmdPara == "全局关闭") | (CmdPara == "全局开启"))
                    return "不能关闭此功能。";

                if (CmdPara == "成语接龙")
                    CmdPara = "接龙";

                //判断参数是否有效
                CmdPara = await BotCmdService.GetCmdNameAsync(CmdPara);
                if (CmdPara != "")
                {
                    bool is_cmd_close = await BotCmdService.IsCmdCloseAllAsync(CmdPara);
                    if (((CmdName == "全局开启") & (!is_cmd_close)) | ((CmdName == "全局关闭") & is_cmd_close))
                        res = CmdPara + "功能已" + CmdName;
                    else
                    {
                        int set_close = 0;
                        if (CmdName == "全局关闭")
                            set_close = 1;
                        _ = await BotCmdService.SetCmdCloseAllAsync(CmdPara, set_close);
                        res = CmdPara + "功能" + CmdName + "成功\n";
                    }
                }
            }

            return res + "\n已全局关闭：\n" + await BotCmdService.GetClosedCmdAsync();
        }

        public string GetCloseAll() => GetCloseAllAsync().GetAwaiter().GetResult();


        // 开启关闭 刷屏/图片/网址/脏话/广告
        // 撤回/扣分/警告/禁言/踢出/拉黑
        public async Task<string> GetTurnOnAsync(string cmdName, string cmdPara, string cmdPara2)
        {
            string keyField = GroupWarn.GetFieldName(cmdPara2);
            string keyword = await GroupInfo.GetValueAsync(keyField, GroupId);
            List<string> keys = [.. keyword.Split('|')];
            if (cmdName == "开启")
            {
                if (keys.Contains(cmdPara))
                    return $"{cmdPara}{cmdPara2}未关闭，无需开启";
                else
                    keys.Add(cmdPara);
            }
            else if (cmdName == "关闭")
            {
                if (!keys.Remove(cmdPara))
                    return $"{cmdPara}{cmdPara2}未开启，无需关闭";
            }
            keyword = string.Join(" ", [.. keys]).Trim().Replace(" ", "|");

            return await GroupInfo.SetValueAsync(keyField, keyword, GroupId) == -1
                ? $"{cmdPara}{cmdPara2}{cmdName}{RetryMsg}"
                : $"✅ {cmdPara}{cmdPara2}已{cmdName}";
        }

        public string GetTurnOn(string cmdName, string cmdPara, string cmdPara2)
            => GetTurnOnAsync(cmdName, cmdPara, cmdPara2).GetAwaiter().GetResult();
    }
}
