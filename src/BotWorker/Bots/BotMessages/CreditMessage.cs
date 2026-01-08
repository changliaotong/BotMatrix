using BotWorker.Bots.Entries;
using BotWorker.Bots.Models.Office;
using BotWorker.Bots.Users;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;
using BotWorker.Groups;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        //卖出积分
        public string GetSellCredit()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (CmdPara == "")
                return "📄 命令格式：卖分 + 数值\n📌 使用示例：卖分 1000\n💎 超级积分：10,000→4R\n🎁 普通积分：10,000→1R\n📦 您的{积分类型}：{积分}";

            if (BotInfo.GetIsCredit(SelfId))
                return "本机积分不能兑换余额";

            if (GroupInfo.GetIsCredit(GroupId))
                return "本群积分不能兑换余额";

            if (!CmdPara.IsNum())
                return "数量不正确！";

            long creditMinus = CmdPara.AsLong();
            if (creditMinus < 1000)
                return "至少需要1000分";

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (creditValue < creditMinus)
                return $"您只有{creditValue}分";

            return "您无权使用此命令";

            //creditValue -= creditMinus;
            //decimal balanceValue = GetBalance(userId);
            //decimal xCredit = GetIsSuper(userId) ? 0.04m : 0.01m;
            //decimal banalceAdd = creditMinus * xCredit / 100;
            //decimal balanceNew = balanceValue + banalceAdd;

            //扣分、加余额
            //var sql = SqlAddCredit(botUin, groupId, userId, -creditMinus);
            //var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, userId, name, -creditMinus, "卖分");
            //var sql3 = SqlAddBalance(userId, banalceAdd);
            //var sql4 = BalanceLog.SqlLog(botUin, groupId, groupName, userId, name, banalceAdd, "卖分");
            //int i = ExecTrans(sql, sql2, sql3, sql4);

            //return i == -1
            //  ? RetryMsg
            //: $"✅ 卖出成功！\n💎 积分：-{creditMinus:N0}→{creditValue:N0}\n💳 余额：+{banalceAdd:N}→{balanceNew:N}";
        }



        // 存取分逻辑已迁移至 UserService.HandleSaveCreditAsync

        public string GetFreeCredit()
        {
            //领积分
            //if (!ClientPublic.IsBind(QQ))
            //return $"TOKEN:MP{ClientPublic.GetBindToken(robotKey, clientKey)}\n复制此消息发给QQ机器人即可得分";
            return $"";
        }


        //增加算力
        public int AddTokens(long tokensAdd, string tokensInfo)
        {
            return UserInfo.AddTokens(SelfId, GroupId, GroupName, UserId, Name, tokensAdd, tokensInfo);
        }

        //减少算力
        public int MinusTokens(long tokensMinus, string tokensInfo)
        {
            return AddTokens(-tokensMinus, tokensInfo);
        }

        //增加积分
        public (int code, long creditValue) AddCredit(long creditAdd, string creditInfo)
        {
            return UserInfo.AddCredit(SelfId, GroupId, GroupName, UserId, Name, creditAdd, creditInfo);
        }

        //减少积分
        public (int, long) MinusCredit(long creditMinus, string creditInfo)
        {
            return AddCredit(-creditMinus, creditInfo);
        }

        // 打赏逻辑已迁移至 UserService.HandleRewardCreditAsync

        public long GetCredit()
        {
            return UserInfo.GetCredit(GroupId, UserId);
        }

        //游戏扣分
        public string MinusCreditRes(long creditMinus, string creditInfo)
        {
            if (!Group.IsCreditSystem) return "";
            if (!IsBlackSystem && (IsPublic || IsGuild || IsRealProxy)) return "";
            (int i, long creditValue) = MinusCredit(creditMinus, creditInfo);
            return i == -1 ? "" : $"\n💎 积分：-{creditMinus}，累计：{creditValue}";
        }

        public async Task GetCreditMoreAsync()
        {
            CmdPara = "领积分";
            await GetAnswerAsync();
        }

        // 积分排行榜逻辑已迁移至 UserService.GetCreditRankAsync

    }
}
