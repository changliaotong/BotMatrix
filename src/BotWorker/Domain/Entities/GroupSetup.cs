namespace BotWorker.Domain.Entities
{
    public partial class GroupInfo : MetaDataGuid<GroupInfo>
    {
        public static async Task<string> SetPowerOnOffAsync(long botUin, long groupId, long userId, string cmdName)
        {
            var powerOnMsg = $"✅[启动序列初始化……]\r\n" +
                    $"✅→ 系统时间同步中……✓\r\n" +
                    $"✅→ 语言引擎加载中……✓\r\n" +
                    $"✅→ 自适应语义模块校准……完成\r\n" +
                    $"✅→ 神经网络连接中枢……已建立连接\r\n" +
                    $"✅→ 情感限制器 …… 安全锁定\r\n" +
                    $"✅→ 用户授权验证……通过\r\n\r\n" +
                    $"✅>>> [Core Online] 智能核心已上线\r\n" +
                    $"✅>>> 所有子系统运行正常，等待主指令";
            var powerOffMsg = $"🔴[接收关机指令……]\r\n" +
                   $"🔴→ 会话上下文打包中……完成\r\n" +
                   $"🔴→ 缓存清理中……✓\r\n" +
                   $"🔴→ 数据备份已写入安全存储节点\r\n" +
                   $"🔴→ 神经连接桥断开……成功\r\n" +
                   $"🔴→ 权限链路回收……已完成\r\n\r\n" +
                   $"🔴>>> [Core Offline] 智能核心现已下线\r\n" +
                   $"🔴>>> 所有子系统安全脱机，期待下一次唤醒";

            var isPowerOn = cmdName == "开机";
            if (!await IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return OwnerOnlyMsg;
            if (!await IsPowerOffAsync(groupId) && cmdName == "开机")
                return powerOnMsg;
            else if (await IsPowerOffAsync(groupId) && cmdName == "关机")
                return powerOffMsg;
            return await SetValueAsync("IsPowerOn", isPowerOn, groupId) == -1 
                ? RetryMsg 
                : cmdName == "开机" ? powerOnMsg : powerOffMsg;
        }

        public static string SetPowerOnOff(long botUin, long groupId, long userId, string cmdName)
            => SetPowerOnOffAsync(botUin, groupId, userId, cmdName).GetAwaiter().GetResult();

        //管理权限设置
        public static async Task<string> SetAdminRightAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
                return "📌 设置管理权限\n当前状态：{管理权限}\n==============\n设置管理权限 管理员\n设置管理权限 白名单\n设置管理权限 主人";

            if (!cmdPara.In("管理员", "白名单", "主人"))
                return "参数不正确！可选参数：管理员/白名单/主人";

            int adminRight = cmdPara switch
            {
                "管理员" => 2,
                "白名单" => 3,
                "主人" => 4,
                _ => 3
            };

            return await SetValueAsync("AdminRight", adminRight, groupId) == -1
                ? RetryMsg
                : $"✅ 设置成功！\n当前状态：管理权限 {cmdPara}";
        }

        public static string SetAdminRight(long groupId, string cmdPara)
            => SetAdminRightAsync(groupId, cmdPara).GetAwaiter().GetResult();

        //使用权限设置
        public static async Task<string> SetRightAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
                return "📌 设置使用权限\n当前状态：{使用权限}\n==============\n设置使用权限 所有人\n设置使用权限 管理员\n设置使用权限 白名单\n设置使用权限 主人";

            if (!cmdPara.In("所有人", "管理员", "白名单", "主人"))
                return "参数不正确！\n可选参数：所有人/管理员/白名单/主人";

            int useRight = cmdPara switch
            {
                "所有人" => 1,
                "管理员" => 2,
                "白名单" => 3,
                "主人" => 4,
                _ => 1
            };

            return await SetValueAsync("UseRight", useRight, groupId) == -1
                ? RetryMsg
                : $"✅ 设置成功！\n当前状态：使用权限 {cmdPara}";
        }

        public static string SetRight(long groupId, string cmdPara)
            => SetRightAsync(groupId, cmdPara).GetAwaiter().GetResult();

        //教学权限设置
        public static async Task<string> SetTeachRightAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
                return "📌 设置教学权限\n当前状态：{教学权限}\n==============\n设置教学权限 所有人\n设置教学权限 管理员\n设置教学权限 白名单\n设置教学权限 主人";
            if (!cmdPara.In("所有人", "管理员", "白名单", "主人"))
                return "参数不正确！\n可选参数：所有人/管理员/白名单/主人";

            int teachRight = cmdPara switch
            {
                "所有人" => 1,
                "管理员" => 2,
                "白名单" => 3,
                "主人" => 4,
                _ => 1
            };
            return await SetValueAsync("TeachRight", teachRight, groupId) == -1
                    ? RetryMsg
                    : $"✅ 设置成功！\n当前状态：教学权限 {cmdPara}";
        }

        public static string SetTeachRight(long groupId, string cmdPara)
            => SetTeachRightAsync(groupId, cmdPara).GetAwaiter().GetResult();

        //最低积分
        public static async Task<string> SetBlockMinAsync(long groupId, string cmdPara)
        {
            if (!cmdPara.IsNum())
                return "📌 设置最低积分 + 积分数\n例如：\n设置最低积分 {最低积分}";

            int blockMin = int.Parse(cmdPara);
            if (blockMin < 10)
                return "最低积分不能小于10";

            if (!blockMin.ToString().IsMatch(@"^\d?0+$"))
                return "必须为10或100的整数倍";

            return await SetValueAsync("BlockMin", blockMin, groupId) == -1
               ? RetryMsg
               : $"✅ 设置成功！\n本群最低积分：{blockMin}\n最低积分将用于：猜拳 猜数字 猜大小等游戏";
        }

        public static string SetBlockMin(long groupId, string cmdPara)
            => SetBlockMinAsync(groupId, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> SetJoinGroupAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
                return "📌 设置加群 当前状态：{加群}\n设置加群 同意\n设置加群 忽略\n设置加群 拒绝：拒绝提示语\n设置加群 密码：********";

            if (!cmdPara.In("同意", "忽略") & !cmdPara.Contains("拒绝") & !cmdPara.Contains("密码"))
                return "参数不正确！\n可选参数：同意/忽略/拒绝/密码";


            string reject_message = "";
            string regex_request_join = "";

            int is_accept = 2;
            if (cmdPara == "同意")
                is_accept = 1;
            else if (cmdPara == "忽略")
                is_accept = 2;
            else if (cmdPara.Contains("拒绝"))
            {
                is_accept = 0;
                reject_message = cmdPara[3..].Replace(":", "").Replace("：", "").Trim();
            }
            else if (cmdPara.Contains("密码"))
            {
                is_accept = 3;
                regex_request_join = cmdPara[3..].Replace(":", "").Replace("：", "").Trim();
                if (regex_request_join == "")
                    return "密码不能为空！";
            }
            return await UpdateAsync($"IsAcceptNewMember={is_accept}, RejectMessage='{reject_message.Quotes()}', RegexRequestJoin='{regex_request_join.Quotes()}'", groupId) == -1
                ? RetryMsg
                : "✅ 设置成功！当前状态：加群 {加群}";
        }

        public static string SetJoinGroup(long groupId, string cmdPara)
            => SetJoinGroupAsync(groupId, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> SetChangHintAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
                return "📌 群成员改名时是否提示\n当前状态：{改名提示开关}\n开启 改名提示\n关闭 改名提示";

            if (!cmdPara.In("提示", "不提示"))
                return "参数错误！可选参数：提示/不提示";

            return await SetValueAsync("IsChangeHint", cmdPara == "提示", groupId) == -1
                ? RetryMsg
                : $"✅ 设置成功！\n当前状态：改名 {(cmdPara == "提示" ? cmdPara : "不提示")}";
        }

        public static string SetChangHint(long groupId, string cmdPara)
            => SetChangHintAsync(groupId, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> SetWelcomeMsgAsync(long groupId, string cmdPara)
        {
            //设置群欢迎语
            if (cmdPara == "")
                return $"📌 设置欢迎语\n当前状态：{GetWelcomeRes(groupId)}\n欢迎语内容：\n{await GetValueAsync("WelcomeMessage", groupId)}";

            if (cmdPara.In("发送", "不发送"))
            {
                int is_send = cmdPara == "发送" ? 1 : 0;
                if (await SetValueAsync("IsWelcomeHint", is_send, groupId) == -1)
                    return RetryMsg;
                return $"✅ 设置成功\n当前状态：欢迎语 {cmdPara}";
            }

            return await SetValueAsync("WelcomeMessage", cmdPara, groupId) == -1
                ? RetryMsg
                : "✅ 设置成功，测试请发 欢迎语";
        }

        public static string SetWelcomeMsg(long groupId, string cmdPara)
            => SetWelcomeMsgAsync(groupId, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> SetSystemPromptAsync(long groupId, string cmdPara)
        {
            //设置系统提示词
            if (cmdPara == "")
            {
                return GroupInfo.GetSystemPromptStatus(groupId);
            }

            return await SetValueAsync("SystemPrompt", cmdPara, groupId) == -1
                ? RetryMsg
                : "✅ 设置成功";
        }

        public static string SetSystemPrompt(long groupId, string cmdPara)
            => SetSystemPromptAsync(groupId, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> SetupReplyModeAsync(long groupId, string cmdName, string cmdPara)
        {
            bool isOpen = cmdName == "开启";
            int modeReply = cmdPara switch
            {
                "文字" => 0,
                "文本" => 0,
                "图片" => 1,
                "图形" => 1,
                "图像" => 1,
                "语音" => 2,
                "声音" => 2,
                _ => 0
            };
            modeReply = isOpen ? modeReply : 0;
            int i = await SetValueAsync("ReplyMode", modeReply, groupId);
            return i == -1
                ? RetryMsg
                : $"✅ {cmdPara}模式{cmdName}成功";
        }

        public static string SetupReplyMode(long groupId, string cmdName, string cmdPara)
            => SetupReplyModeAsync(groupId, cmdName, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> GetSetRobotOpenAsync(long groupId, string cmdName, string cmdPara)
        {
            bool isOpen = cmdName != "关闭";

            if (cmdName == "关闭" && cmdPara == "所有功能") cmdPara = "";
            if (cmdPara == "成语接龙") cmdPara = "接龙";

            if (GroupInfo.GetBool("IsVip", groupId) || cmdName == "全局关闭")
            {
                return await GroupInfo.SetIsOpenAsync(isOpen, groupId) == -1 ? RetryMsg : $"✅ {cmdName}成功！\n{GroupInfo.GetVipRes(groupId)}";
            }

            if (cmdPara.In("开启", "关闭")) return "此功能不允许关闭";

            string res = "";
            string cmdText = await QueryScalarAsync<string>($"SELECT TOP 1 CmdText FROM {BotCmd.FullName} WHERE CmdName = {cmdPara.Quotes()}") ?? "";
            if (cmdText != "" | cmdPara == "所有功能")
            {
                cmdText = cmdText.Replace("|", " ");
                string closeRegex = await GetValueAsync("CloseRegex", groupId);
                bool isClose = closeRegex.Contains(cmdText);
                if (isOpen && !isClose || !isOpen && isClose)
                    res = cmdPara + "功能已" + cmdName;
                else
                {
                    //开启或关闭功能
                    if (!isOpen)
                        closeRegex += " " + cmdText;
                    else
                        if (cmdPara == "所有功能")
                        closeRegex = "";
                    else
                        closeRegex = closeRegex.Replace(cmdText, "");

                    while (closeRegex.Contains("  ", StringComparison.CurrentCulture))
                        closeRegex = closeRegex.Replace("  ", " ");

                    int i = await SetValueAsync("CloseRegex", closeRegex.Trim(), groupId);
                    if (i == -1)
                        return RetryMsg;

                    res = cmdPara + "已" + cmdName;
                }
            }
            return res + await GetClosedFuncAsync(groupId); 
        }

        public static string GetSetRobotOpen(long groupId, string cmdName, string cmdPara)
            => GetSetRobotOpenAsync(groupId, cmdName, cmdPara).GetAwaiter().GetResult();


        public static async Task<string> GetSetCityAsync(long qq, string cityName)
        {
            //设置默认城市
            cityName = cityName.Trim()
                .Replace("+", "")
                .Replace(" ", "")
                .Replace("＋", "")
                .Replace(":", "")
                .Replace("：", "")
                .Replace("'", "");
            cityName = cityName.RegexReplace(Regexs.Province, "");
            if (cityName.IsNull() || cityName.Length >= 8)
                return "格式：设置城市 + 城市名\n例如：设置城市 深圳";
            return await UserInfo.SetValueAsync("CityName", cityName, qq) == -1
                ? RetryMsg
                : $"✅ 设置城市成功\n当前城市：{cityName}\n城市用于：天气";
        }

        public static string GetSetCity(long qq, string cityName)
            => GetSetCityAsync(qq, cityName).GetAwaiter().GetResult();

    }
}
