using System.Threading.Tasks;
using BotWorker.Bots.BotMessages;
using BotWorker.Core.Repositories;

namespace BotWorker.Core.Services
{
    public interface IUserService
    {
        /// <summary>
        /// 处理黑名单相关指令
        /// </summary>
        Task<string> HandleBlacklistAsync(BotMessage botMsg);

        /// <summary>
        /// 添加黑名单
        /// </summary>
        Task<int> AddBlackAsync(BotMessage botMsg, long targetUserId, string reason);

        /// <summary>
        /// 检查并执行自动签到
        /// </summary>
        Task<string> ProcessAutoSignInAsync(BotMessage botMsg);

        /// <summary>
        /// 处理用户权限设置
        /// </summary>
        Task<string> HandleUserPermissionAsync(BotMessage botMsg);

        /// <summary>
        /// 兑换金币/积分
        /// </summary>
        Task<string> ExchangeCoinsAsync(BotMessage botMsg, string cmdPara, string cmdPara2);

        /// <summary>
        /// 处理存取积分指令
        /// </summary>
        Task<string> HandleSaveCreditAsync(BotMessage botMsg);

        /// <summary>
        /// 处理打赏指令
        /// </summary>
        Task<string> HandleRewardCreditAsync(BotMessage botMsg);

        /// <summary>
        /// 获取积分排行榜
        /// </summary>
        Task<string> GetCreditRankAsync(BotMessage botMsg);
    }

    public class UserService : IUserService
    {
        private readonly IUserRepository _userRepository;
        private readonly IGroupRepository _groupRepository;
        private readonly IBotApiService _apiService;
        private readonly IPermissionService _permissionService;

        public UserService(
            IUserRepository userRepository,
            IGroupRepository groupRepository,
            IBotApiService apiService,
            IPermissionService permissionService)
        {
            _userRepository = userRepository;
            _groupRepository = groupRepository;
            _apiService = apiService;
            _permissionService = permissionService;
        }

        public async Task<string> HandleBlacklistAsync(BotMessage botMsg)
        {
            if (!_permissionService.IsAdmin(botMsg))
                return "您没有权限管理黑名单";

            var message = botMsg.CurrentMessage;
            var targetUserId = botMsg.CurrentMessage.Common.Exts.GetQq();
            
            if (targetUserId == 0)
            {
                // 如果没有指定 QQ，可能是在请求黑名单列表
                return "未指定目标QQ。黑名单管理指令：拉黑+QQ，取消拉黑+QQ";
            }

            var isBlack = await _userRepository.IsBlackAsync(targetUserId);
            var cmdName = botMsg.CmdName;

            if (cmdName.Contains("拉黑") || cmdName.Contains("黑名单"))
            {
                if (isBlack) return $"用户 {targetUserId} 已经在黑名单中";
                
                // 额外的安全检查
                if (targetUserId == botMsg.UserId) return "不能拉黑你自己";
                if (targetUserId == botMsg.Group.RobotOwner) return "不能拉黑我主人";
                
                await _userRepository.SetIsBlackAsync(targetUserId, true, "管理员手动拉黑");
                
                // 自动踢人
                await _apiService.KickMemberAsync(botMsg.SelfId, botMsg.RealGroupId, targetUserId);
                
                return $"已将用户 {targetUserId} 加入黑名单并移出群聊";
            }
            else if (cmdName.Contains("取消") || cmdName.Contains("移除") || cmdName.Contains("解黑"))
            {
                if (!isBlack) return $"用户 {targetUserId} 不在黑名单中";
                await _userRepository.SetIsBlackAsync(targetUserId, false);
                return $"已将用户 {targetUserId} 从黑名单移除";
            }

            return "未知黑名单操作";
        }

        public async Task<int> AddBlackAsync(BotMessage botMsg, long targetUserId, string reason)
        {
            return await _userRepository.SetIsBlackAsync(targetUserId, true, reason);
        }

        public async Task<string> ProcessAutoSignInAsync(BotMessage botMsg)
        {
            // 这里原本是 BotMessage 内部的签到逻辑调用
            // 以后可以在这里实现具体的签到逻辑，包括更新积分等
            return ""; 
        }

        public async Task<string> HandleUserPermissionAsync(BotMessage botMsg)
        {
            // 实现用户权限设置逻辑
            return "";
        }

        public async Task<string> ExchangeCoinsAsync(BotMessage botMsg, string cmdPara, string cmdPara2)
        {
            if (!cmdPara2.IsNum())
                return "数量不正确";

            long coinsValue = cmdPara2.AsLong();
            if (coinsValue < 10)
                return "数量最少为10";

            if (cmdPara == "积分" || cmdPara == "群积分")
                cmdPara = "本群积分";

            // 模拟 CoinsLog.conisNames 逻辑
            string[] coinNames = { "金币", "黑金币", "紫币", "游戏币", "本群积分" };
            int coinsType = Array.IndexOf(coinNames, cmdPara);
            if (coinsType == -1) return "未知的兑换类型";

            long minusCredit = coinsValue * 120 / 100;
            long creditGroupId = botMsg.GroupId;

            if (coinsType == 4) // 本群积分
            {
                var isOpen = await _groupRepository.GetIsOpenAsync(botMsg.GroupId);
                if (!isOpen) return "未开启本群积分，无法兑换";
                creditGroupId = 0;
            }

            long creditValue = await _userRepository.GetCreditAsync(creditGroupId, botMsg.UserId);
            bool isSuper = await _userRepository.IsSuperAdminAsync(botMsg.UserId);
            if (isSuper) minusCredit = coinsValue;

            string res = "";
            string saveRes = "";

            if (creditValue < minusCredit)
            {
                long creditSave = await _userRepository.GetSaveCreditAsync(botMsg.UserId);
                if (cmdPara == "本群积分" && creditSave >= minusCredit - creditValue)
                {
                    // 这里原本有 WithdrawCredit 逻辑，暂时简化
                    long needed = minusCredit - creditValue;
                    await _userRepository.AddSaveCreditAsync(botMsg.UserId, -needed, "兑换扣除");
                    creditValue += needed;
                    creditSave -= needed;
                    saveRes = $"\n取分：{needed}，累计：{creditSave}";
                }
                else
                {
                    return $"您的积分{creditValue}不足{minusCredit}";
                }
            }

            // 执行兑换
            await _userRepository.AddCreditAsync(botMsg.SelfId, creditGroupId, botMsg.UserId, -minusCredit, $"兑换{cmdPara}*{coinsValue}");
            await _userRepository.AddCoinsAsync(coinsType, coinsValue, botMsg.GroupId, botMsg.UserId, $"兑换{cmdPara}*{coinsValue}");

            long currentCoins = await _userRepository.GetCoinsAsync(coinsType, botMsg.GroupId, botMsg.UserId);
            long currentCredit = await _userRepository.GetCreditAsync(creditGroupId, botMsg.UserId);

            res = $"兑换{cmdPara}：{coinsValue}，累计：{currentCoins}{saveRes}\n积分：-{minusCredit}，累计：{currentCredit}";
            return res;
        }

        public async Task<string> HandleSaveCreditAsync(BotMessage botMsg)
        {
            if (!botMsg.Group.IsCreditSystem)
                return "抱歉，本群未开启积分系统";

            if (string.IsNullOrEmpty(botMsg.CmdPara))
                return "格式：存分 + 积分数\n取分 + 积分数\n例如：存分 100";

            if (!botMsg.CmdPara.IsNum())
                return "数量不正确，请输入数字";

            long creditOper = botMsg.CmdPara.AsLong();
            string cmdName = botMsg.CmdName.ToLower();

            if (cmdName.StartsWith('存') || cmdName.StartsWith('c'))
                cmdName = "存分";
            else if (cmdName.StartsWith('取') || cmdName.StartsWith('q'))
                cmdName = "取分";

            long creditValue = await _userRepository.GetCreditAsync(botMsg.GroupId, botMsg.UserId);
            long creditSave = await _userRepository.GetSaveCreditAsync(botMsg.UserId);

            if (cmdName == "存分")
            {
                if (creditOper == 0) creditOper = creditValue;
                if (creditOper == 0) return "您没有积分可存";
                if (creditValue < creditOper) return $"您只有 {creditValue:N0} 分，余额不足";

                // 存分：使用 Task 模式，原子操作 + 事务后 ReSync
                int result = UserInfo.ExecTrans(
                    UserInfo.TaskSaveCredit(botMsg.SelfId, botMsg.GroupId, botMsg.UserId, creditOper),
                    CreditLog.SqlHistory(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, -creditOper, "存分")
                );

                if (result == -1) return "系统繁忙，请稍后再试";
                
                creditValue -= creditOper;
                creditSave += creditOper;
            }
            else if (cmdName == "取分")
            {
                if (creditOper == 0) creditOper = creditSave;
                if (creditOper == 0) return "您没有积分可取";
                if (creditSave < creditOper) return $"您已存积分只有 {creditSave:N0} 分，余额不足";

                // 取分：同样使用 Task 模式
                int result = UserInfo.ExecTrans(
                    UserInfo.TaskSaveCredit(botMsg.SelfId, botMsg.GroupId, botMsg.UserId, -creditOper),
                    CreditLog.SqlHistory(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, creditOper, "取分")
                );

                if (result == -1) return "系统繁忙，请稍后再试";

                creditValue += creditOper;
                creditSave -= creditOper;
            }

            return $"✅ {cmdName}成功：{creditOper:N0}\n" +
                   $"💰 当前积分：{creditValue:N0}\n" +
                   $"🏦 已存积分：{creditSave:N0}\n" +
                   $"📈 积分总额：{creditValue + creditSave:N0}";
        }

        public async Task<string> HandleRewardCreditAsync(BotMessage botMsg)
        {
            if (!botMsg.Group.IsCreditSystem)
                return "抱歉，本群未开启积分系统";

            string regex_reward;
            if (botMsg.CmdPara.IsMatch(Regexs.CreditParaAt))
                regex_reward = Regexs.CreditParaAt;
            else if (botMsg.CmdPara.IsMatch(Regexs.CreditParaAt2))
                regex_reward = Regexs.CreditParaAt2;
            else if (botMsg.CmdPara.IsMatch(Regexs.CreditPara))
                regex_reward = Regexs.CreditPara;
            else
                return "🎉 打赏格式：\n打赏 [QQ号] [积分]\n📌 例如：\n打赏 51437810 100";

            long rewardQQ = botMsg.CmdPara.RegexGetValue(regex_reward, "UserId").AsLong();
            long rewardCredit = botMsg.CmdPara.RegexGetValue(regex_reward, "credit").AsLong();

            if (rewardCredit < 10)
                return "至少打赏 10 积分";

            // 计算打赏者需要付出的总积分（含 20% 服务费）
            long creditMinus = rewardCredit * 12 / 10;
            
            // 检查是否为超级管理员或合作伙伴（免服务费）
            bool isSuper = await _userRepository.IsSuperAdminAsync(botMsg.UserId);
            // 这里暂且简化处理，后续可注入 IPartnerService 检查
            if (isSuper) creditMinus = rewardCredit;

            long senderCredit = await _userRepository.GetCreditAsync(botMsg.GroupId, botMsg.UserId);
            if (senderCredit < creditMinus)
                return $"您的积分 {senderCredit:N0} 不足 {creditMinus:N0}";

            // 执行转账
            // 1. 扣除发送者积分
            await _userRepository.AddCreditAsync(botMsg.SelfId, botMsg.GroupId, botMsg.UserId, -creditMinus, $"打赏支出:{rewardQQ}");
            // 2. 增加接收者积分
            await _userRepository.AddCreditAsync(botMsg.SelfId, botMsg.GroupId, rewardQQ, rewardCredit, $"收到打赏:{botMsg.UserId}");

            long currentSenderCredit = await _userRepository.GetCreditAsync(botMsg.GroupId, botMsg.UserId);
            long currentReceiverCredit = await _userRepository.GetCreditAsync(botMsg.GroupId, rewardQQ);

            string transferFeeMsg = isSuper ? "" : $"\n💸 服务费：{rewardCredit * 2 / 10:N0}";

            return $"✅ 打赏成功！\n" +
                   $"🎉 打赏积分：{rewardCredit:N0}{transferFeeMsg}\n" +
                   $"🎯 对方积分：{currentReceiverCredit:N0}\n" +
                   $"🙋 您的积分：{currentSenderCredit:N0}";
        }

        public async Task<string> GetCreditRankAsync(BotMessage botMsg)
        {
            var rankData = await _userRepository.GetCreditRankAsync(botMsg.GroupId);
            
            var sb = new StringBuilder();
            sb.AppendLine("🏆 积分排行榜");
            
            int i = 1;
            bool userInTop = false;
            foreach (var item in rankData)
            {
                string icon = i switch
                {
                    1 => "🥇",
                    2 => "🥈",
                    3 => "🥉",
                    4 => "4️⃣",
                    5 => "5️⃣",
                    6 => "6️⃣",
                    7 => "7️⃣",
                    8 => "8️⃣",
                    9 => "9️⃣",
                    10 => "🔟",
                    _ => $"{i}."
                };
                
                sb.AppendLine($"{icon} [@:{item.UserId}] 💎{item.Credit:N0}");
                if (item.UserId == botMsg.UserId) userInTop = true;
                i++;
            }
            
            if (!userInTop)
            {
                long userCredit = await _userRepository.GetCreditAsync(botMsg.GroupId, botMsg.UserId);
                sb.AppendLine($"\n您的排名未入前十 [@:{botMsg.UserId}] 💎{userCredit:N0}");
            }
            
            return sb.ToString();
        }
    }
}
