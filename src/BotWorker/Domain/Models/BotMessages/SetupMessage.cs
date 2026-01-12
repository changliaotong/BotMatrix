using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public async Task<string> GetShutupResAsync()
        {
            if (IsRobotOwner())
                return await Task.FromResult("");
            else
            {
                return await Task.FromResult("");
            }
        }

        // 机器人是否开启状态 机器人、群、使用人
        public bool HaveUseRight()
        {           
            return Group.UseRight switch
            {
                1 => true,
                2 => IsRobotOwner() || UserPerm < 2,
                3 => IsRobotOwner() || IsWhiteList(),
                4 => IsRobotOwner(),
                _ => false
            };
        }

        public async Task<string> GetSetupUrlAsync()
        {
            if (!IsPublic && IsGroup)
                return "安全起见，请私聊使用此功能";

            (int i, var token) = Token.Append(UserId);
            if (i == -1)
                return RetryMsg;

            string loginMethod = "登录方法：\n1. 点击下方链接直接进入\n2. 或在登录页面输入您的QQ号和TOKEN";

            if (IsGuild)
            {
                var OldAnswer = Answer;
                var OldDelayMs = DelayMs;
                Answer = $"sz84点com";
                DelayMs = 3000;
                await SendMessageAsync();
                Answer = $"{token}";
                DelayMs = 5000;
                await SendMessageAsync();
                Answer = OldAnswer;
                DelayMs = OldDelayMs;
                return $"早喵机器人后台地址：sz84点com\n{loginMethod}\n您的登录TOKEN（请勿转发他人）：{token}";
            }
            else
            {
                return $"早喵机器人后台地址：{SetupUrl}\n{loginMethod}\n以下地址可直接进入后台（请勿转发他人）\n{SetupUrl}/login?t={token}";
            }
        }

        public string SetupPrivate(bool adminRight = false, bool teachRight = false)
        {
            if ((!IsGroup) && (RealGroupId == BotInfo.GroupCrm) && (!IsSuperAdmin))
                return "请先设置默认群\n命令格式：\n设置群 + 群号码\n例如：\n设置群 " + BotInfo.GroupIdDef;

            if (adminRight && (!HaveSetupRight()))
                return "您无权修改本群设置！";

            if (teachRight && (!HaveTeachRight()))
                return GroupId == 0
                    ? $"您无权在群({RealGroupId})教我说话"
                    : "您无权在本群教我说话";

            return "";
        }

        public async Task<string> SetupPrivateAsync(bool adminRight = false, bool teachRight = false)
        {
            return await Task.FromResult(SetupPrivate(adminRight, teachRight));
        }

        public const string RegexDissayTime = @"^(?<dissay_time>\d+)(?<time_unit>(小时|分钟))$";
        public const string RegexCmdPara = @"^[#＃﹟]{0,1}(?<cmdName>("
                                  + @"glqx|管理权限|"
                                  + @"syqx|使用权限|"
                                  + @"tjqx|jxqx|调教权限|教学权限|调校权限|"
                                  + @"ltms|聊天模式|wd|问答|wdms|问答模式|lt|聊天|"
                                  + @"zdjf|最低积分|"
                                  + @"hyy|欢迎语|"
                                  + @"tq|退群|"
                                  + @"bt|被踢|tc|踢出|"
                                  + @"jq|加群|"
                                  + @"gm|改名|"
                                  + @"qz|前缀|"
                                  + @"ql|群链|ai|tsc|提示词|ai提示词|系统提示词|ai系统提示词|"
                                  + @"cs|城市|mrcs|默认城市|"
                                  + @"sl|私链|"
                                  + @"q|群|mrq|默认群|"
                                  + @"yy|语音|yl|音聊|sy|声音"
                                  + @"))[ \\/+]*(?<cmdPara>[\s\S]*)";
        public async Task<string> SetupResAsync()
        {            
            var (cmdName, cmdPara) = GetCmdPara(CmdPara, RegexCmdPara);
            string res;

            if (cmdName == "")
            {
                if (SelfId != 3889494926)
                {
                    res = "⚙️===群设置===\n";
                    if (HaveSetupRight())
                        res += "设置管理权限 {管理权限}\n" +
                               "设置使用权限 {使用权限}\n" +
                               "设置教学权限 {教学权限}\n" +
                               "设置聊天模式 {聊天模式}\n" +
                               "设置最低积分 {最低积分}\n" +
                               "设置提示词\n" +
                              $"设置欢迎语 {(Group.IsWelcomeHint ? "发送" : "不发送")}\n" +
                              $"设置语音 {VoiceMapUtil.GetVoiceName(Group.VoiceId ?? "")}\n" +
                               "设置退群 {退群提示} {退群拉黑}\n" +
                               "设置被踢 {被踢提示} {被踢拉黑}\n";
                    else
                        res += "管理权限 {管理权限}\n" +
                               "使用权限 {使用权限}\n" +
                               "教学权限 {教学权限}\n" +
                               "聊天模式 {聊天模式}\n" +
                               "最低积分 {最低积分}\n" +
                               "退群 {退群提示} {退群拉黑}\n" +
                               "被踢 {被踢提示} {被踢拉黑}\n";

                    res += $"自动签到 {(Group.IsAutoSignin ? "已开启" : "已关闭")}\n" +
                           $"👤======个人设置======\n" +
                           $"设置城市 {User.CityName}\n" +
                           $"{(User.IsShutup ? "闭嘴模式 已开启\n" : "")}";
                }
                else
                {
                    res = $"👤======个人设置======\n" +
                          $"设置城市 {User.CityName}\n" +
                          $"{(User.IsShutup ? "闭嘴模式 已开启\n" : "")}";
                }
                
                return res;
            }

            if (cmdName.In("默认提示", "mrts"))
                return GetTurnOn(cmdName, cmdPara);

            if (cmdName.In("城市", "默认城市", "mrcs", "cs"))
                return GroupInfo.GetSetCity(UserId, cmdPara);

            if (cmdName.In("私链", "sl"))
                return cmdPara.In("开启", "关闭")
                    ? GetTurnOn(cmdPara, cmdName)
                    : "私链开关：{私链开关}\n设置格式：\n开启 私链\n关闭 私链";

            if (cmdName.In("群", "默认群", "mrq", "q"))
                return SetDefaultGroup(cmdPara);

            //前面为个人设置，后面群设置需要权限
            res = SetupPrivate(true, false);
            if (res != "")
                return res;

            if (cmdName.In("语音", "yy"))
                return await SetGroupVoiceAsync(cmdPara);

            if (cmdName.In("欢迎语", "hhy"))
                return GroupInfo.SetWelcomeMsg(GroupId, cmdPara);

            if (cmdName.In("ai", "tsc", "提示词", "ai提示词", "系统提示词", "ai系统提示词"))
                return GroupInfo.SetSystemPrompt(GroupId, cmdPara);

            if (cmdName.In("管理权限", "glqx"))
                return GroupInfo.SetAdminRight(GroupId, cmdPara);

            if (cmdName.In("使用权限", "syqx"))
                return GroupInfo.SetRight(GroupId, cmdPara);

            if (cmdName.In("调教权限", "教学权限", "tjqx", "jxqx"))
                return GroupInfo.SetTeachRight(GroupId, cmdPara);

            if (cmdName.In("聊天模式", "问答", "聊天", "问答模式", "wd", "lt", "wdms", "ltms"))
                return SetCloudAnswer(GroupId, UserId, cmdPara);

            if (cmdName.In("最低积分", "zdjf"))
                return GroupInfo.SetBlockMin(GroupId, cmdPara);

            if (cmdName.In("加群", "jq"))
                return GroupInfo.SetJoinGroup(GroupId, cmdPara);

            if (cmdName.In("退群", "tq"))
                return SetExitGroup(GroupId, cmdPara);

            if (cmdName.In("被踢", "踢出", "bt", "tc"))
                return SetKickBlack(GroupId, cmdPara);

            if (cmdName.In("改名", "gm"))
                return GroupInfo.SetChangHint(GroupId, cmdPara);

            if (cmdName.In("群链", "ql"))
                return (cmdPara.Trim() == "")
                    ? "群链：{私链开关}\n开启 私链\n关闭 私链"
                    : GetTurnOn(cmdPara, cmdName);

            return HaveSetupRight()
                ? "参数错误\n可选参数：\n管理权限/使用权限/教学权限/聊天模式/欢迎语/提示词/加群/退群/被踢/改名/城市/私链/群"
                : "参数错误\n可选参数：城市/私链/群";
        }

        

        public async Task<string> SetGroupVoiceAsync(string input)
        {
            // 1. 无输入：显示语音列表（分组 + 编号）
            if (string.IsNullOrWhiteSpace(input))
            {
                var curId = Group?.VoiceId;
                var list = VoiceMapUtil.BuildVoiceList(curId ?? "");
                return list + "\n发送：设置语音 + 名称 / 编号\n例如：设置语音 8";
            }

            input = input.Trim();

            // 2. 支持数字编号
            if (int.TryParse(input, out int num))
            {
                var hit = VoiceMapUtil.FindByIndex(num);
                if (hit == null)
                    return "❌ 语音编号不存在";
                return await SaveVoice(hit.Value.Id, hit.Value.Name);
            }

            // 3. 精准匹配名称
            if (VoiceMapUtil.NameToId.TryGetValue(input, out var exactId))
            {
                return await SaveVoice(exactId, input);
            }

            // 4. 模糊匹配（自动选第一个，无状态友好）
            var like = VoiceMapUtil.All
                .FirstOrDefault(v => v.Name.Contains(input, StringComparison.OrdinalIgnoreCase));

            if (like != null)
                return await SaveVoice(like.Id, like.Name) + "（模糊匹配）";

            // 5. 特殊快捷指令
            if (input.Equals("随机", StringComparison.OrdinalIgnoreCase))
            {
                var all = VoiceMapUtil.All;
                var v = all[Random.Shared.Next(all.Count)];
                return await SaveVoice(v.Id, v.Name) + "（随机）";
            }

            return "❌ 未找到语音，请发送：设置语音";
        }

        private async Task<string> SaveVoice(string voiceId, string voiceName)
        {
            int i = GroupInfo.SetValue("VoiceId", voiceId, GroupId);
            if (i == -1) return RetryMsg;

            if (IsQQ)
            {
                // 找出所有分组
                var groupNames = VoiceMap.Categories
                    .Where(cat => cat.Items.Any(v => v.Id == voiceId))
                    .Select(cat => cat.Name)
                    .ToList();

                string categoryName = string.Join("、", groupNames);

                // 找试听 URL
                string url = VoiceMap.Categories
                    .SelectMany(cat => cat.Items)
                    .FirstOrDefault(v => v.Id == voiceId)?.PreviewUrl ?? "";

                Answer = $"[CQ:music,type=custom,url={url},title={voiceName},content={categoryName},audio={url},image={UserInfo.GetHead(UserId)}]";
                await SendMessageAsync();
            }

            return $"✅ 设置成功！{voiceName}";
        }

        public string SetExitGroup(long groupId, string cmdPara)
        {
            if (cmdPara == "")
                return "📌 设置退群\n当前状态：{退群提示} {退群拉黑}\n设置退群 提示/不提示/拉黑/不拉黑";

            cmdPara = cmdPara.Replace("加黑", "拉黑");

            if (!("提示 不提示 拉黑 不拉黑".Split(' ').Any(p => cmdPara.Contains(p))))
                return "参数不正确！可选参数 提示/不提示/拉黑/不拉黑";
            
            if (GroupInfo.SetValue("IsExitHint", Group.IsExitHint = cmdPara.Contains("提示") && !cmdPara.Contains("不提示"), groupId) == -1
             || GroupInfo.SetValue("IsBlackExit", Group.IsBlackExit = cmdPara.Contains("拉黑") && !cmdPara.Contains("不拉黑"), groupId) == -1)
                return RetryMsg;

            return "✅ 设置成功！\n当前状态：有人退群时 {退群提示} {退群拉黑}";
        }

        public string SetKickBlack(long groupId, string cmdPara)
        {
            if (cmdPara == "")
                return "📌 设置被踢\n当前状态：{被踢提示} {被踢拉黑}\n设置被踢：提示/不提示/拉黑/不拉黑";

            cmdPara = cmdPara.Replace("加黑", "拉黑");

            if (!("提示 不提示 拉黑 不拉黑".Split(' ').Any(p => cmdPara.Contains(p))))
                return "参数不正确！可选参数 提示/不提示/拉黑/不拉黑";

            if (GroupInfo.SetValue("IsExitHint", Group.IsKickHint = cmdPara.Contains("提示") && !cmdPara.Contains("不提示"), groupId) == -1
             || GroupInfo.SetValue("IsBlackExit", Group.IsBlackKick = cmdPara.Contains("拉黑") && !cmdPara.Contains("不拉黑"), groupId) == -1)
                return RetryMsg;

            return "✅ 设置成功！\n当前状态：有人被踢时 {被踢提示} {被踢拉黑}";
        }


        public string SetDefaultGroup(string cmdPara)
        {
            if (cmdPara == "")
            {
                if (IsGroup)
                    cmdPara = GroupId.ToString();
                else
                {
                    //私聊不加群号时显示该用户名下的所有群供参考
                    string res = QueryRes($"SELECT TOP 5 GroupId, GroupName FROM {FullName} WHERE GroupOwner = {UserId} and Valid = 1 ORDER BY GroupName",
                                           "\n{1}({0})");
                    if (res != "")
                        res = $"您是主人的群：{res}";

                    return $"设置群 + 群号 例如：\n设置群 {User.DefaultGroup}\n{res}";
                }
            }

            //设置默认群
            if (!cmdPara.IsNum())
                return $"群号不正确\n设置群 + 群号 例如：\n设置群 {User.DefaultGroup}";

            string defaultGroup = cmdPara;
            if (defaultGroup == BotInfo.GroupCrm.ToString())
                defaultGroup = "null";

            return (UserInfo.SetValue("DefaultGroup", defaultGroup, UserId) == -1)
                ? RetryMsg
                : $"✅ 您的群设置为：{cmdPara}\n默认群用于私聊时：\n设置 教学 闲聊 逗你玩";
        }

        //聊天模式设置
        public string SetCloudAnswer(long GroupId, long qq, string cmdPara)
        {
            if (cmdPara == "")
                return "💬 当前模式：{聊天模式}\n📌 可选模式：闭嘴/本群/官方/话唠/终极/AI/纯血AI\n💡 切换方法：开启 {聊天模式}";

            if (!cmdPara.In("闭嘴", "本群", "官方", "话唠", "话痨", "终极", "AI", "纯血AI"))
                return "模式不正确！\n可选模式：闭嘴/本群/官方/话唠/终极/AI/纯血AI";

            int isCloud = cmdPara.ToUpper() switch
            {
                "闭嘴" => 0,
                "本群" => 1,
                "官方" => 2,
                "话痨" => 3,
                "话唠" => 3,
                "终极" => 4,
                "AI" => 5,
                "纯血AI" => 6,
                _ => 0
            };

            if (isCloud >= 4 && !IsGuild && SystemSetting.IsCloudLimited && !GroupVip.IsForever(GroupId))
                return "非永久版不能使用终极模式";

            int i = GroupInfo.SetValue("IsCloudAnswer", isCloud, GroupId);
            if (i == -1)
                return RetryMsg;
            
            var res = $"✅ 设置成功！当前设置：{cmdPara.ToUpper()}";
            if (!IsGuild)
            {
                if (isCloud == 3 && !GroupVip.IsYearVIP(GroupId))
                    res += "\n本群只能体验【话唠模式】至凌晨4点，长期使用需升级为年费版";
                else if (isCloud == 4 && !GroupVip.IsForever(GroupId))
                    res += "\n本群只能体验【终极模式】至凌晨4点，长期使用需升级为永久版";
                if (isCloud == 5 && !GroupVip.IsForever(GroupId))
                    res += "\n本群只能体验【AI模式】至凌晨4点，长期使用需升级为永久版";
                else if (isCloud == 6 && !GroupVip.IsForever(GroupId))
                    res += "\n本群只能体验【纯血AI模式】至凌晨4点，长期使用需升级为永久版";
            }
            else if (GroupId > GroupOffical.MIN_GROUP_ID)
            {
                res += $"\n📌 本机器人需 @ 使用，如需免艾特权限，请联系客服升级";
            }

            return res;
        }

        public async Task GetShortcutSetAsync()
        {
            var cmdPara = CmdPara;
            if (CmdPara == "猜拳")
            {
                List<string> cmds = ["剪刀", "石头", "布"];
                foreach (var cmd in cmds)
                {
                    CmdPara = cmd;
                    await GetCmdResAsync();
                }
                Answer = $"✅ {cmdPara}已{CmdName}";
                return;
            }
            else if (CmdPara == "猜大小")
            {                
                List<string> cmds = ["押大", "押小", "押单", "押双", "押全围", "押点", "押对"];
                foreach (var cmd in cmds)
                {
                    CmdPara = cmd;
                    await GetCmdResAsync();
                }
                Answer = $"✅ {cmdPara}已{CmdName}";
                return;
            }

            CmdPara = CmdPara.Replace("话痨", "话唠").Replace("模式", "");

            int isOpen = -1;
            if (CmdName == "开启")
                isOpen = 1;

            if (isOpen == -1)
            {
                switch (CmdPara)
                {
                    case "聊天":
                        CmdPara = "问答闭嘴";
                        break;

                    default:
                        var downgradeMap = new Dictionary<string, string>(StringComparer.OrdinalIgnoreCase)
                        {
                            { "纯血AI", "问答AI" },
                            { "AI", "问答终极" },
                            { "终极", "问答话唠" },
                            { "话唠", "问答官方" },
                            { "官方", "问答本群" },
                            { "本群", "问答闭嘴" },
                            { "闭嘴", "问答话唠" }
                        };

                        if (downgradeMap.TryGetValue(CmdPara, out var newCmd))
                        {
                            CmdPara = newCmd;
                        }
                        else
                        {
                            CmdPara += "关闭";
                        }
                        break;
                }
            }
            else if (isOpen == 1)
            {
                if (CmdPara.In("闭嘴", "本群", "官方", "话唠", "终极", "AI", "纯血AI"))
                    CmdPara = "问答" + CmdPara;
                else
                    CmdPara += "开启";
            }
            CmdName = "设置";
            CurrentMessage = $"{CmdName}{CmdPara}";
            await GetCmdResAsync();
        }

        public async Task<string> GetWarnSetupAsync(string regexCmd)
        {
            if (!IsRobotOwner() && !BotInfo.IsAdmin(SelfId, UserId))
            {
                return OwnerOnlyMsg;
            }
            string cmdName = Message.RegexGetValue(regexCmd, "CmdName");
            _ = Message.RegexGetValue(regexCmd, "cmdPara");
            cmdName = GroupWarn.GetCmdName(cmdName);
            regexCmd = Regexs.WarnPara;
            if (Message.IsMatch(regexCmd))
            {
                var matches = Message.Matches(regexCmd);
                foreach (Match match in matches.Cast<Match>())
                {
                    string cmdPara = match.Groups["cmdPara"].Value;
                    string cmdPara2 = match.Groups["cmdPara2"].Value;
                    cmdPara = GroupWarn.GetCmdPara(cmdPara);
                    regexCmd = Regexs.WarnPara2;
                    if (cmdPara2.IsMatch(regexCmd))
                    {
                        var matches2 = cmdPara2.Matches(regexCmd);
                        foreach (var match2 in matches2.Cast<Match>())
                        {
                            cmdPara2 = match2.Groups["cmdPara2"].Value;
                            cmdPara2 = GroupWarn.GetCmdPara(cmdPara2);
                            Answer += "\n" + await GetTurnOnAsync(cmdName, cmdPara, cmdPara2);
                        }
                    }
                }
            }
            Answer = $"✅ 命令执行结果：{Answer}";
            Answer += GroupId == 0 ? "\n设置群 {默认群}" : "";
            return Answer;
        }

        public void GetWarnSetup(string regexCmd)
        {
            _ = GetWarnSetupAsync(regexCmd).GetAwaiter().GetResult();
        }

        //管理权限
        public bool HaveSetupRight()
        {
            if (UserPerm == 0 || BotInfo.IsAdmin(SelfId, UserId) || IsRobotOwner())
                return true;           

            return Group.AdminRight switch
            {
                2 => UserPerm < 2,
                3 => IsWhiteList(),
                4 => IsRobotOwner(),
                _ => false
            };
        }

        //教学权限
        public bool HaveTeachRight()
        {           
            if (!IsGroup || Group.TeachRight == 1 || IsRobotOwner())
                return true;

            return Group.TeachRight switch
            {
                2 => UserPerm < 2,
                3 => IsWhiteList(),
                4 => IsRobotOwner(),
                _ => false
            };
        }

        public async Task<string> SetPowerAsync(bool powerOn)
        {
            if (!HaveSetupRight()) return "您无权修改本群设置！";
            Group.IsPowerOn = powerOn;
            await Group.UpdateAsync();
            return powerOn ? "机器人已开机" : "机器人已关机";
        }

        public async Task<string> SetOpenAsync(bool isOpen)
        {
            if (!HaveSetupRight()) return "您无权修改本群设置！";
            Group.IsOpen = isOpen;
            await Group.UpdateAsync();
            return isOpen ? "机器人已开启" : "机器人已关闭";
        }

        public async Task<string> GetOpenAsync(bool open)
        {
            return await SetPowerAsync(open);
        }

        public async Task<string> HandleSetupAsync()
        {
            return await SetupResAsync();
        }
    }
}
