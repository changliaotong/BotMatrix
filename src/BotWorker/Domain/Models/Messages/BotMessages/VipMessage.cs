using sz84.Bots.Entries;
using sz84.Bots.Games;
using sz84.Bots.Groups;
using sz84.Bots.Models.Office;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;
using sz84.Bots.Users;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 换群
        public string GetChangeGroup()
        {
            IsCancelProxy = true;

            if (!CmdPara.IsMatchQQ())
                return "群号不正确，请发命令\n换群 + 新群号";

            if (!GroupVip.IsVip(GroupId))
                return "体验版无需换群";

            if (!IsRobotOwner())
                return $"你无权换群，你不是群【{GroupId}】机器人主人，";

            long new_groupId = long.Parse(CmdPara);
            if (GroupVip.IsVip(new_groupId))
                return $"不能换到群【{new_groupId}】，该群已有机器人";

            if (!User.IsSuper)
                return $"非超级分用户不能自己换群，请联系客服QQ处理";

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (creditValue < 12000)
                return $"您的积分{creditValue}不足12000，换群需扣除12000积分";
            if (!IsConfirm)
                return ConfirmMessage("换群将扣除12000分");

            (int i, creditValue) = AddCredit(-12000, "换群扣分");
            if (i == -1)
                return RetryMsg;

            i = GroupVip.ChangeGroup(GroupId, new_groupId, UserId);
            if (i == -1)
                return RetryMsg;

            return $"✅ 换群成功！将机器人加入新群即可使用\n积分：-12000，累计：{creditValue}";
        }

        // 换主人
        public string GetChangeOwner()
        {
            IsCancelProxy = true;

            if (!IsRobotOwner())
                return $"您不是群【{GroupId}】机器人主人，无权换主人";

            if (!CmdPara.IsMatchQQ())
                return $"参数不正确，请发命令 #换主人 + QQ";

            if (!User.IsSuper)
                return $"非超级分用户不能自己换主人，请联系客服QQ处理";

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (creditValue < 12000)
                return $"换主人需扣除12000分，您的积分{creditValue}不足";

            (int i, creditValue) = AddCredit(-12000, "换主人扣分");
            if (i == -1)
                return RetryMsg;

            long newUserId = long.Parse(CmdPara);
            i = GroupInfo.SetValue("RobotOwner", newUserId, GroupId);
            if (i == -1)
                return RetryMsg;

            GroupVip.SetValue("UserId", newUserId, GroupId);

            return $"✅ 换主人成功！\n积分：-12000，累计：{creditValue}";
        }

        public string GetBuyRobot()
        {
            IsCancelProxy = true;

            string res = SetupPrivate();
            if (res != "")
                return res;

            if (!IsVip)
                return "本群没有开通VIP，余额仅可用于续费";

            if (!CmdPara.IsNum())
                return "📄 格式：续费 + 月数\n📌 例如：续费12\n🔹【续费1】1个月20元\n🔹【续费2】2个月35元\n🔹【续费3】3个月50元\n🔹【续费6】半年80元\n🔹【续费12】一年120元\n🔹【续费24】两年200元\n🔹【续费999】永久498元\n💳 您的余额：{余额}";

            int month = CmdPara.AsInt();
            decimal robotPrice = Price.GetRobotPrice(month);
            decimal balance = UserInfo.GetBalance(UserId);
            if (balance < robotPrice)
                return $"您的余额{balance:N}不足{robotPrice:N}";

            var sql = UserInfo.SqlAddBalance(UserId, -robotPrice);
            var sql2 = BalanceLog.SqlLog(SelfId, GroupId, GroupName, UserId, Name, -robotPrice, $"群{GroupId}续费{month}个月");
            var sql3 = Income.SqlInsert(GroupId, month, "机器人", 0, "余额", "", $"余额支付:{robotPrice}", UserId, BotInfo.SystemUid);
            var sql4 = GroupVip.SqlBuyVip(GroupId, GroupName, UserId, month, robotPrice, "使用余额续费");
            int i = ExecTrans(sql, sql2, sql3, sql4);
            return i == -1
                ? RetryMsg
                : $"✅ 群{GroupId}续费{month}个月\n💳 余额：-{robotPrice:N}，累计：{{余额}}\n{{VIP}}";
        }

        // 购买 买入命令分类 买分 买道具 购买一切 根据不同参数调用不同的函数
        public string GetBuyRes()
        {
            if (CmdPara.Contains("积分"))
            {
                CmdPara = CmdPara.Replace("积分", "").Replace("jf", "").Trim();
                return UserInfo.GetBuyCredit(this, SelfId, GroupId, GroupName, UserId, Name, CmdPara);
            }
            else if ((CmdPara == "禁言卡") | (CmdPara == "飞机票") | (CmdPara == "道具"))
                return GroupProps.GetBuyRes(SelfId, GroupId, GroupName, UserId, Name, CmdPara);
            else
                return PetOld.GetBuyPet(SelfId, GroupId, GroupId, GroupName, UserId, Name, CmdPara);
        }

        // 兑换礼品
        public string GetGoodsCredit()
        {
            if (!User.IsSuper)
                return $"仅超级积分可兑换礼品，你的积分类型：{{积分类型}}";

            long creditValue = UserInfo.GetCredit(GroupId, UserId);

            if (CmdPara == "")
                return "红富士苹果包邮12斤：\n 24个装（中果）：119,520分\n换中果发送【兑换礼品 119520】\n您的{积分类型}：{积分}";

            if (CmdPara != "119520")
                return "参数不正确";

            if (creditValue < 119520)
                return $"您的积分{creditValue}不足119,520";

            if (!IsConfirm)
                return ConfirmMessage("119520分换苹果一箱24个装");

            if (MinusCredit(44160, "兑换礼品 苹果一箱24个装（中果）").Item1 == -1)
                return RetryMsg;

            return "✅ 兑换苹果一箱24个装（中果）成功，请联系客服QQ为您安排发货";
        }

        // 升级为超级分 
        public string GetUpgrade()
        {
            if (!CmdPara.IsMatchQQ())
                return "命令格式：\n升级 + QQ\n例如：\n升级 {客服QQ}";

            if (Partner.IsNotPartner(UserId))
                return "非合伙人无权使用此命令";

            long upgradeQQ = CmdPara.GetAtUserId();
            if (UserInfo.GetIsSuper(upgradeQQ))
                return "已为超级积分，无需升级";

            long creditValue = UserInfo.GetTotalCredit(upgradeQQ);
            if (creditValue > 1000)
                return $"该用户有{creditValue}分，升级前请先将原有积分清零";

            int res = UserInfo.Update($"is_super=1, super_date=getdate(), ref_qq={UserId}", upgradeQQ); ;
            if (res == -1)
                return RetryMsg;

            return $"✅ {upgradeQQ}升级超级积分成功！";
        }

        // 降级为普通分
        public string GetCancelSuper()
        {
            if (CmdPara != "")
                return "";

            if (!User.IsSuper)
                return "普通积分无需降级";

            if (IsConfirm && UserInfo.GetCredit(UserId) <= 1000)
            {
                int i = UserInfo.SetValue("IsSuper", false, UserId);
                return i == -1 ? RetryMsg : "降级成功";
            }
            else
                return ConfirmMessage("确认降级为普通积分");
        }


        // 版本及有效期
        public string GetVipRes()
        {
            IsCancelProxy = true;

            string res;

            if (GroupId == 0 || IsPublic)
            {
                string sql = $"select top 5 GroupId, abs(datediff(day, getdate(), EndDate)) as res from {GroupVip.FullName} where UserId = {UserId} order by EndDate";
                res = QueryRes(sql, "{0} 有效期：{1}天\n");
                return res;
            }

            string version;

            if (GroupVip.Exists(GroupId))
            {
                if (GroupVip.IsYearVIP(GroupId))
                    version = "年费版";
                else
                    version = "VIP版";
                int valid_days = GroupVip.RestDays(GroupId);
                if (valid_days >= 1850)
                    res = "『永久版』";
                else
                    res = $"『{version}』有效期：{valid_days}天";
            }
            else
            {
                if (GroupVip.IsVipOnce(GroupId))
                    return "已过期，请及时续费";
                else
                    version = "体验版";
                res = $"『{version}』";
            }

            return res;
        }
    }
}
