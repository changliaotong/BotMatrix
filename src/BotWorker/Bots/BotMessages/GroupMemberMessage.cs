using BotWorker.Bots.Entries;
using BotWorker.Bots.Games.Gift;
using BotWorker.Bots.Users;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;
using BotWorker.Groups;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 兑换本群积分/金币/紫币等
        public string ExchangeCoins(string cmdPara, string cmdPara2)
        {
            if (!cmdPara2.IsNum())
                return "数量不正确";

            long coinsValue = cmdPara2.AsLong();
            if (coinsValue < 10)
                return "数量最少为10";

            if ((cmdPara == "积分") | (cmdPara == "群积分"))
                cmdPara = "本群积分";

            int coinsType = CoinsLog.conisNames.IndexOf(cmdPara);
            long minusCredit = coinsValue * 120 / 100;

            long creditGroup = GroupId;

            if (coinsType == (int)CoinsLog.CoinsType.groupCredit)
            {
                if (!GroupInfo.GetIsCredit(GroupId))
                    return "未开启本群积分，无法兑换";
                creditGroup = 0;
            }

            long creditValue = UserInfo.GetCredit(creditGroup, UserId);

            if (UserInfo.GetIsSuper(UserId))
                minusCredit = coinsValue;

            string res = "";
            string saveRes = "";

            if (creditValue < minusCredit)
            {
                //兑换本群积分时，可直接扣已存积分
                long creditSave = UserInfo.GetSaveCredit(UserId);
                if ((cmdPara == "本群积分") & (creditSave >= minusCredit - creditValue))
                {
                    int i = WithdrawCredit(minusCredit - creditValue, ref creditValue, ref creditSave, ref res);
                    if (i == -1)
                        return res;
                    else
                        saveRes = $"\n取分：{minusCredit - creditValue}，累计：{creditSave}";
                }
                else
                    return $"您的积分{creditValue}不足{minusCredit}";
            }
            creditValue -= minusCredit;
            //扣分 记录积分记录 增加金币 记录金币变化记录
            var sqlAddCredit = UserInfo.SqlAddCredit(SelfId, creditGroup, UserId, -minusCredit);
            var sqlCreditHis = CreditLog.SqlHistory(SelfId, creditGroup, GroupName, UserId, Name, -minusCredit, $"兑换{cmdPara}*{coinsValue}");
            var sqlPlusCoins = GroupMember.SqlPlus(CoinsLog.conisFields[coinsType], coinsValue, GroupId, UserId);
            long coinsValue2 = 0;
            var sqlCoinsHis = CoinsLog.SqlCoins(SelfId, GroupId, GroupName, UserId, Name, coinsType, coinsValue, ref coinsValue2, $"兑换{cmdPara}*{coinsValue}");
            if (ExecTrans(sqlAddCredit, sqlCreditHis, sqlPlusCoins, sqlCoinsHis) == -1)
                res = RetryMsg;
            else
                res = $"兑换{cmdPara}：{coinsValue}，累计：{coinsValue2}{saveRes}\n{UserInfo.GetCreditType(creditGroup, UserId)}：-{minusCredit}，累计：{creditValue}";
            return res;
        }

        public string GetGiftRes(long userGift, string giftName, int giftCount = 1)
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (CmdPara == "")
                return $"{GroupGift.GiftFormat}\n\n{Gift.GetGiftList(GroupId, UserId)}";

            List<string> users = CmdPara.GetValueList(Regexs.Users);
            CmdPara = CmdPara.RegexReplace(Regexs.Users, "");
            List<string> NumList = CmdPara.GetValueList(@"\d{1,4}");
            CmdPara = CmdPara.RegexReplace(@"\d{1,4}", "");
            giftCount = NumList.Count == 0 ? 1 : NumList.First().AsInt();
            giftName = CmdPara;
            string res = "";

            foreach (string user in users)
            {
                userGift = user.AsLong();
                res += GroupGift.GetGiftRes(SelfId, GroupId, GroupName, UserId, Name, userGift, giftName, giftCount);
            }

            return res;
        }

        // 爱群主
        public async Task<string> GetLampRes()
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (!GroupGift.IsFans(GroupId, UserId))
            {
                Answer = GetBingFans("加团");
                if (!IsPublic)
                    await SendMessageAsync();
            }            

            var lampTime = GroupGift.LampMinutes(GroupId, UserId);
            if (lampTime < 10)
                return $"📌 粉丝灯牌已点亮！\n" +
                       $"🧊 冷却时间：{10 - lampTime}分钟\n" +
                       $"💖 亲密度值：{{亲密度值}}\n" +
                       $"🎖️ 粉丝排名：第{{粉丝排名}}名 LV{{粉丝等级}}\n";

            long creditMinus = IsGuild ? RandomInt(1, 1200) : 100;
            long creditAdd = creditMinus / 2;
            long groupOwner = GroupInfo.GetGroupOwner(GroupId);

            long creditOwner = UserInfo.GetCredit(GroupId, groupOwner);
            creditOwner += creditAdd;
            
            //送灯牌过程：更新灯牌时间、亲密值、积分记录、更新积分、主人积分更新
            if (UserId == creditOwner)
                creditOwner -= creditMinus;

            var sql = GroupGift.SqlLightLamp(GroupId, UserId);
            var sql2 = CreditLog.SqlHistory(SelfId, GroupId, GroupName, UserId, Name, creditMinus, "爱群主");
            var sql3 = UserInfo.SqlAddCredit(SelfId, GroupId, UserId, creditMinus);
            var sql4 = CreditLog.SqlHistory(SelfId, GroupId, GroupName, groupOwner, GroupInfo.GetRobotOwnerName(GroupId), creditAdd, "爱群主");
            var sql5 = UserInfo.SqlAddCredit(SelfId, GroupId, groupOwner, creditAdd);
            return ExecTrans(sql, sql2, sql3, sql4, sql5) == -1
                ? RetryMsg
                : $"🚀 成功点亮粉丝灯牌！\n" +
                  $"💖 亲密指数：+100→{{亲密度值}}\n" +
                  $"💎 群主积分：+{creditAdd}→{creditOwner:N0}\n" +
                  $"🎖️ 粉丝排名：第{{粉丝排名}}名 LV{{粉丝等级}}\n" +
                  $"🧊 冷却时间：10分钟\n" +
                  $"💎 积分：+{creditMinus}，累计：{{积分}}";
        }

        // 加入粉丝团
        public string GetBingFans(string cmdName)
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (cmdName == "加团")
            {
                if (GroupGift.IsFans(GroupId, UserId))
                    return "您已是粉丝团成员，无需再次加入";

                long creditMinus = 100;
                long creditValue = UserInfo.GetCredit(GroupId, UserId);
                if (creditValue < creditMinus)
                    return $"您的积分{creditValue}不足{creditMinus}加入粉丝团";

                //更新group_member/扣分/积分记录
                var sql = GroupGift.SqlBingFans(GroupId, UserId);
                var sql2 = CreditLog.SqlHistory(SelfId, GroupId, GroupName, UserId, Name, -creditMinus, "加团扣分");
                var sql3 = UserInfo.SqlAddCredit(SelfId, GroupId, UserId, -creditMinus);
                int i = ExecTrans(sql, sql2, sql3);
                return (i == -1)
                    ? RetryMsg
                    : $"✅ 恭喜您成为第{GroupGift.GetFansCount(GroupId)}名粉丝团成员\n亲密度值：+100，累计：{{亲密度值}}\n积分：-{creditMinus}，累计：{creditValue - creditMinus}";
            }
            if (cmdName == "退灯牌")
            {
                if (!GroupGift.IsFans(GroupId, UserId))
                    return "您尚未加入粉丝团";

                //退粉丝团
                if (Exec($"UPDATE {FullName} SET IsFans = 0, FansValue = 0, FansLevel = 0 WHERE GroupId = {GroupId} AND UserId = {UserId}") == -1)
                    return RetryMsg;
                return "✅ 成功退出粉丝团";
            }
            return "";
        }

        // 爱早喵
        public static async Task<string> GetLoveZaomiaoRes()
        {
            //todo 完善爱早喵功能
            return $"早喵也爱你，么么哒";
        }
    }
}
