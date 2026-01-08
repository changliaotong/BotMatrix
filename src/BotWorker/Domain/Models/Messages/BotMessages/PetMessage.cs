using BotWorker.Domain.Entities;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 赎身
        public string GetFreeMe()
        {            
            if (!Group.IsPet)
                return PetOld.InfoClosed;

            //以当前主人购买时的价格成交，对方只能得到80%，系统扣除20%
            long currMaster = PetOld.GetCurrMaster(Group.Id, UserId);
            if (currMaster == UserId)
                return "您已是自由身，无需赎身";

            long buyPrice = PetOld.GetBuyPrice(Group.Id, UserId);
            long creditAdd = buyPrice;
            long creditMinus = buyPrice * 12 / 10;
            if (User.IsSuper)
                creditMinus = creditMinus * 22 / 10;
            long creditValue = UserInfo.GetCredit(Group.Id, UserId);
            if (creditValue < creditMinus)
                return $"您的积分{creditValue}不足{creditMinus}";

            if (!IsConfirm)
                return ConfirmMessage($"赎身需扣分：-{creditMinus}");

            long credit_value2 = 0;
            return PetOld.DoFreeMe(SelfId, GroupId, GroupName, UserId, Name, currMaster, creditMinus, creditAdd) == -1
                ? RetryMsg
                : $"✅ 赎身成功！\n[@:{currMaster}]积分：+{creditAdd}，累计：{credit_value2}\n您的积分：-{creditMinus}，累计：{{积分}}";
        }
    }
}
