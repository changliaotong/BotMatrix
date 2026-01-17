using System;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Infrastructure.Persistence;
using BotWorker.Domain.Repositories;
using BotWorker.Common;

namespace BotWorker.Modules.Games
{
    public class GroupService : IGroupService
    {
        private readonly IGroupRepository _groupRepository;
        private readonly IUserRepository _userRepository;
        private readonly IBotCmdRepository _botCmdRepository;
        private readonly IBotRepository _botRepository;
        private readonly IGroupOfficalRepository _groupOfficalRepository;

        public GroupService(
            IGroupRepository groupRepository, 
            IUserRepository userRepository,
            IBotCmdRepository botCmdRepository,
            IBotRepository botRepository,
            IGroupOfficalRepository groupOfficalRepository)
        {
            _groupRepository = groupRepository;
            _userRepository = userRepository;
            _botCmdRepository = botCmdRepository;
            _botRepository = botRepository;
            _groupOfficalRepository = groupOfficalRepository;
        }

        private const string RetryMsg = "⚠️ 操作失败，请稍后再试";
        private const string OwnerOnlyMsg = "❌ 只有群主或系统管理员可以执行此操作。";

        public async Task<string> SetPowerOnOffAsync(long botUin, long groupId, long userId, string cmdName)
        {
            return await TransactionWrapper.ExecuteAsync(async (wrapper) =>
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
                if (!await _groupRepository.IsOwnerAsync(groupId, userId, wrapper.Transaction) && !await _botRepository.IsAdminAsync(botUin, userId))
                    return OwnerOnlyMsg;
                if (!await _groupRepository.IsPowerOffAsync(groupId, wrapper.Transaction) && cmdName == "开机")
                    return powerOnMsg;
                else if (await _groupRepository.IsPowerOffAsync(groupId, wrapper.Transaction) && cmdName == "关机")
                    return powerOffMsg;
                
                return await _groupRepository.UpdateIsPowerOnAsync(groupId, isPowerOn, wrapper.Transaction) == -1
                    ? RetryMsg
                    : cmdName == "开机" ? powerOnMsg : powerOffMsg;
            });
        }

        public async Task<string> SetAdminRightAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
            {
                string current = await _groupRepository.GetAdminRightResAsync(groupId);
                return $"📌 设置管理权限\n当前状态：{current}\n==============\n设置管理权限 管理员\n设置管理权限 白名单\n设置管理权限 主人";
            }

            if (!cmdPara.In("管理员", "白名单", "主人"))
                return "参数不正确！可选参数：管理员/白名单/主人";

            int adminRight = cmdPara switch
            {
                "管理员" => 2,
                "白名单" => 3,
                "主人" => 4,
                _ => 3
            };

            return await _groupRepository.UpdateAdminRightAsync(groupId, adminRight) == -1
                ? RetryMsg
                : $"✅ 设置成功！\n当前状态：管理权限 {cmdPara}";
        }

        public async Task<string> SetRightAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
            {
                string current = await _groupRepository.GetRightResAsync(groupId);
                return $"📌 设置使用权限\n当前状态：{current}\n==============\n设置使用权限 所有人\n设置使用权限 管理员\n设置使用权限 白名单\n设置使用权限 主人";
            }

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

            return await _groupRepository.UpdateUseRightAsync(groupId, useRight) == -1
                ? RetryMsg
                : $"✅ 设置成功！\n当前状态：使用权限 {cmdPara}";
        }

        public async Task<string> SetTeachRightAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
            {
                string current = await _groupRepository.GetTeachRightResAsync(groupId);
                return $"📌 设置教学权限\n当前状态：{current}\n==============\n设置教学权限 所有人\n设置教学权限 管理员\n设置教学权限 白名单\n设置教学权限 主人";
            }
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
            return await _groupRepository.UpdateTeachRightAsync(groupId, teachRight) == -1
                    ? RetryMsg
                    : $"✅ 设置成功！\n当前状态：教学权限 {cmdPara}";
        }

        public async Task<string> SetBlockMinAsync(long groupId, string cmdPara)
        {
            if (!cmdPara.IsNum())
            {
                int current = await _groupRepository.GetBlockMinAsync(groupId);
                return "📌 设置最低积分 + 积分数\n" +
                       $"当前最低积分：{current}\n" +
                       "例如：\n设置最低积分 100";
            }

            int blockMin = int.Parse(cmdPara);
            if (blockMin < 10)
                return "最低积分不能小于10";

            if (!blockMin.ToString().IsMatch(@"^\d?0+$"))
                return "必须为10或100的整数倍";

            return await _groupRepository.UpdateBlockMinAsync(groupId, blockMin) == -1
               ? RetryMsg
               : $"✅ 设置成功！\n本群最低积分：{blockMin}\n最低积分将用于：猜拳 猜数字 猜大小等游戏";
        }

        public async Task<string> SetJoinGroupAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
            {
                string current = await _groupRepository.GetJoinResAsync(groupId);
                return $"📌 设置加群 当前状态：{current}\n设置加群 同意\n设置加群 忽略\n设置加群 拒绝：拒绝提示语\n设置入群审批 密码：********";
            }

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
            return await _groupRepository.UpdateJoinGroupSettingsAsync(groupId, is_accept, reject_message.Quotes(), regex_request_join.Quotes()) == -1
                ? RetryMsg
                : "✅ 设置成功！当前状态：加群 {加群}";
        }

        public async Task<string> SetChangHintAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
            {
                string current = await _groupRepository.GetIsChangeHintResAsync(groupId);
                return $"📌 群成员改名时是否提示\n当前状态：{current}\n开启 改名提示\n关闭 改名提示";
            }

            if (!cmdPara.In("提示", "不提示"))
                return "参数错误！可选参数：提示/不提示";

            return await _groupRepository.UpdateIsChangeHintAsync(groupId, cmdPara == "提示") == -1
                ? RetryMsg
                : $"✅ 设置成功！\n当前状态：改名 {(cmdPara == "提示" ? cmdPara : "不提示")}";
        }

        public async Task<string> SetWelcomeMsgAsync(long groupId, string cmdPara)
        {
            //设置群欢迎语
            if (cmdPara == "")
            {
                var welcomeRes = await _groupRepository.GetWelcomeResAsync(groupId);
                var welcomeMessage = await _groupRepository.GetValueAsync("WelcomeMessage", groupId);
                return $"📌 设置欢迎语\n当前状态：{welcomeRes}\n欢迎语内容：\n{welcomeMessage}";
            }

            if (cmdPara.In("发送", "不发送"))
            {
                bool is_send = cmdPara == "发送";
                if (await _groupRepository.UpdateIsWelcomeHintAsync(groupId, is_send) == -1)
                    return RetryMsg;
                return $"✅ 设置成功\n当前状态：欢迎语 {cmdPara}";
            }

            return await _groupRepository.UpdateWelcomeMessageAsync(groupId, cmdPara) == -1
                ? RetryMsg
                : "✅ 设置成功，测试请发 欢迎语";
        }

        public async Task<string> SetSystemPromptAsync(long groupId, string cmdPara)
        {
            //设置系统提示词
            if (cmdPara == "")
            {
                return await _groupRepository.GetSystemPromptStatusAsync(groupId);
            }

            return await _groupRepository.UpdateSystemPromptAsync(groupId, cmdPara) == -1
                ? RetryMsg
                : "✅ 设置成功";
        }

        public async Task<string> SetupReplyModeAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
            {
                string current = await _groupRepository.GetReplyModeResAsync(groupId);
                return $"📌 设置回复模式\n当前状态：{current}\n设置：文字/图片/语音";
            }
            // Note: In the previous implementation, cmdName was used to determine isOpen.
            // But usually this is called as part of a command like "开启 文字模式".
            // For now, assume it's always "开启" if this method is called.
            bool isOpen = true; 
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
            int i = await _groupRepository.UpdateReplyModeAsync(groupId, modeReply);
            return i == -1
                ? RetryMsg
                : $"✅ {cmdPara}模式开启成功";
        }

        public async Task<string> GetSetRobotOpenAsync(long groupId, string cmdPara)
        {
            // Similar to SetupReplyModeAsync, we need to know if it's open or close.
            // Usually cmdPara contains the command.
            bool isOpen = !cmdPara.StartsWith("关闭");
            string actualCmdPara = cmdPara.Replace("开启", "").Replace("关闭", "").Trim();
            string cmdName = isOpen ? "开启" : "关闭";

            if (cmdName == "关闭" && actualCmdPara == "所有功能") actualCmdPara = "";
            if (actualCmdPara == "成语接龙") actualCmdPara = "接龙";

            if (actualCmdPara == "")
            {
                return await _groupRepository.SetIsOpenAsync(isOpen, groupId) == -1 ? RetryMsg : $"✅ {cmdName}成功！\n{await _groupRepository.GetVipResAsync(groupId)}";
            }

            if (actualCmdPara.In("开启", "关闭")) return "此功能不允许关闭";

            string res = "";
            string cmdText = await _botCmdRepository.GetCmdTextAsync(actualCmdPara);
            if (cmdText != "" | actualCmdPara == "所有功能")
            {
                cmdText = cmdText.Replace("|", " ");
                string closeRegex = await _groupRepository.GetValueAsync<string>("CloseRegex", groupId) ?? "";
                bool isClose = closeRegex.Contains(cmdText);
                if (isOpen && !isClose || !isOpen && isClose)
                    res = actualCmdPara + "功能已" + cmdName;
                else
                {
                    //开启或关闭功能
                    if (!isOpen)
                        closeRegex += " " + cmdText;
                    else
                        if (actualCmdPara == "所有功能")
                        closeRegex = "";
                    else
                        closeRegex = closeRegex.Replace(cmdText, "");

                    while (closeRegex.Contains("  ", StringComparison.CurrentCulture))
                        closeRegex = closeRegex.Replace("  ", " ");

                    int i = await _groupRepository.UpdateCloseRegexAsync(groupId, closeRegex.Trim());
                    if (i == -1)
                        return RetryMsg;

                    res = actualCmdPara + "已" + cmdName;
                }
            }
            return res + await _groupRepository.GetClosedFuncAsync(groupId); 
        }

        public async Task<bool> GetBoolAsync(string field, long groupId)
        {
            return await _groupRepository.GetValueAsync<bool>(field, groupId);
        }

        public async Task<string> GetSetCityAsync(long userId, string cmdPara)
        {
            //设置默认城市
            cmdPara = cmdPara.Trim()
                .Replace("+", "")
                .Replace(" ", "")
                .Replace("市", "");

            if (cmdPara == "") return "请输入城市名称";

            var city = await _groupRepository.GetCityAsync(cmdPara);
            if (city == null) return "未找到该城市";

            return await _userRepository.UpdateCityAsync(userId, city.CityName) == -1
                ? RetryMsg
                : $"✅ 默认城市已设置为：{city.CityName}";
        }

        public async Task<bool> GetBoolAsync(string field, long groupId)
        {
            var val = await _groupRepository.GetValueAsync<string>(field, groupId);
            if (string.IsNullOrEmpty(val)) return false;
            return val == "true";
        }


        public async Task<string> SetCloudAnswerAsync(long groupId, string cmdPara)
        {
            if (cmdPara == "")
            {
                string current = await _groupRepository.CloudAnswerResAsync(groupId);
                return $"📌 设置云端词库\n当前状态：{current}\n设置：闭嘴/本群/官方/话痨/终极/AI";
            }

            string[] answers = { "闭嘴", "本群", "官方", "话痨", "终极", "AI" };
            int index = Array.IndexOf(answers, cmdPara);
            if (index == -1)
                return "参数不正确！可选参数 闭嘴/本群/官方/话痨/终极/AI";

            return (await _groupRepository.UpdateIsCloudAnswerAsync(groupId, index) == -1)
                ? RetryMsg
                : $"✅ 设置成功！\n当前状态：云端词库 {cmdPara}";
        }

        public async Task<string> SetExitGroupAsync(long groupId, string cmdPara, GroupInfo group)
        {
            if (cmdPara == "")
            {
                string hintStr = group.IsExitHint ? "提示" : "不提示";
                string blackStr = group.IsBlackExit ? "拉黑" : "不拉黑";
                return $"📌 设置退群\n当前状态：{hintStr} {blackStr}\n设置退群 提示/不提示/拉黑/不拉黑";
            }

            cmdPara = cmdPara.Replace("加黑", "拉黑");
            string[] validParams = { "提示", "不提示", "拉黑", "不拉黑" };
            if (!validParams.Any(p => cmdPara.Contains(p)))
                return "参数不正确！可选参数 提示/不提示/拉黑/不拉黑";

            bool isExitHint = cmdPara.Contains("提示") && !cmdPara.Contains("不提示");
            bool isBlackExit = cmdPara.Contains("拉黑") && !cmdPara.Contains("不拉黑");

            if (await _groupRepository.UpdateExitGroupSettingsAsync(groupId, isExitHint, isBlackExit) == -1)
                return RetryMsg;

            group.IsExitHint = isExitHint;
            group.IsBlackExit = isBlackExit;

            string resHint = isExitHint ? "提示" : "不提示";
            string resBlack = isBlackExit ? "拉黑" : "不拉黑";
            return $"✅ 设置成功！\n当前状态：有人退群时 {resHint} {resBlack}";
        }

        public async Task<string> SetKickBlackAsync(long groupId, string cmdPara, GroupInfo group)
        {
            if (cmdPara == "")
            {
                string hintStr = group.IsKickHint ? "提示" : "不提示";
                string blackStr = group.IsBlackKick ? "拉黑" : "不拉黑";
                return $"📌 设置被踢\n当前状态：{hintStr} {blackStr}\n设置被踢：提示/不提示/拉黑/不拉黑";
            }

            cmdPara = cmdPara.Replace("加黑", "拉黑");
            string[] validParams = { "提示", "不提示", "拉黑", "不拉黑" };
            if (!validParams.Any(p => cmdPara.Contains(p)))
                return "参数不正确！可选参数 提示/不提示/拉黑/不拉黑";

            bool isKickHint = cmdPara.Contains("提示") && !cmdPara.Contains("不提示");
            bool isBlackKick = cmdPara.Contains("拉黑") && !cmdPara.Contains("不拉黑");

            if (await _groupRepository.UpdateKickBlackSettingsAsync(groupId, isKickHint, isBlackKick) == -1)
                return RetryMsg;

            group.IsKickHint = isKickHint;
            group.IsBlackKick = isBlackKick;

            string resHint = isKickHint ? "提示" : "不提示";
            string resBlack = isBlackKick ? "拉黑" : "不拉黑";
            return $"✅ 设置成功！\n当前状态：有人被踢时 {resHint} {resBlack}";
        }

        public async Task<string> SetDefaultGroupAsync(long userId, long groupId, bool isGroup, string cmdPara, string botUinDef)
        {
            if (string.IsNullOrEmpty(cmdPara))
            {
                if (isGroup)
                {
                    cmdPara = groupId.ToString();
                }
                else
                {
                    var ownedGroups = await _groupRepository.GetOwnedGroupsAsync(userId);
                    StringBuilder sb = new();
                    foreach (var g in ownedGroups)
                    {
                        sb.Append($"\n{g.GroupName}({g.GroupId})");
                    }
                    string res = sb.ToString();
                    if (res != "")
                        res = $"您是主人的群：{res}";

                    return $"设置群 + 群号 例如：\n设置群 123456\n{res}";
                }
            }

            if (!long.TryParse(cmdPara, out _))
                return $"群号不正确\n设置群 + 群号 例如：\n设置群 123456";

            string defaultGroup = cmdPara;
            if (defaultGroup == botUinDef)
                defaultGroup = "null";

            return (await _userRepository.SetValueAsync("DefaultGroup", defaultGroup, userId) == -1)
                ? RetryMsg
                : $"✅ 您的群设置为：{cmdPara}\n默认群用于私聊时：\n设置 教学 闲聊 逗你玩";
        }

        public async Task<bool> GetBoolAsync(string field, long groupId)
        {
            var val = await _groupRepository.GetValueAsync<string>(field, groupId);
            return val == "1" || val?.ToLower() == "true";
        }

        public async Task<int> SetValueAsync(string field, object value, long groupId)
        {
            return await _groupRepository.SetValueAsync(field, value, groupId);
        }

        public async Task<(long groupId, bool isNew)> GetGroupIdAsync(string groupOpenid, string groupName, long userId, long botUin = 0, string botName = "")
        {
            var groupId = await _groupOfficalRepository.GetTargetGroupAsync(groupOpenid);
            if (groupId != 0)
                return (groupId, false);

            groupId = await _groupOfficalRepository.GetMaxGroupIdAsync();
            int i = await _groupRepository.AppendAsync(groupId, groupName, botUin, botName, userId, userId, groupOpenid);
            return i == -1 ? (0, false) : (groupId, true);
        }

        public async Task<bool> GetIsCreditAsync(long groupId)
        {
            return groupId != 0 && await GetBoolAsync("IsCredit", groupId);
        }

        public async Task<int> SetPowerOffAsync(long groupId)
        {
            return await SetValueAsync("IsPowerOn", false, groupId);
        }

        public async Task<int> SetPowerOnAsync(long groupId)
        {
            return await SetValueAsync("IsPowerOn", true, groupId);
        }

        public async Task<bool> GetPowerOnAsync(long groupId)
        {
            return await GetBoolAsync("IsPowerOn", groupId);
        }

        public async Task<bool> IsPowerOffAsync(long groupId)
        {
            return !await GetPowerOnAsync(groupId);
        }

        public async Task<bool> IsCanTrialAsync(long groupId)
        {
            // TODO: Implement full logic including GroupVip check
            return await GetBoolAsync("IsValid", groupId);
        }

        public async Task<int> SetInvalidAsync(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0)
        {
            // TODO: Implement full logic
            return await SetValueAsync("IsValid", false, groupId);
        }
    }
}
