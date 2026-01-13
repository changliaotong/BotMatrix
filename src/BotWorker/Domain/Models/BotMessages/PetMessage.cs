namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 赎身
        public async Task<string> GetFreeMeAsync()
        {
            var checkResult = await CheckFreeMeAsync();
            if (checkResult.Error != null) return checkResult.Error;

            int res = await PetOld.DoFreeMeAsync(SelfId, GroupId, GroupName, UserId, Name, checkResult.CurrMaster, checkResult.CreditMinus, checkResult.CreditAdd);
            if (res == -1)
                return RetryMsg;

            long currentCredit = await UserInfo.GetCreditAsync(GroupId, UserId);
            long masterCredit = await UserInfo.GetCreditAsync(GroupId, checkResult.CurrMaster);
            return $"✅ 赎身成功！\n[@:{checkResult.CurrMaster}]积分：+{checkResult.CreditAdd}，累计：{masterCredit}\n您的积分：-{checkResult.CreditMinus}，累计：{currentCredit}";
        }

        private async Task<(string? Error, long CurrMaster, long CreditAdd, long CreditMinus)> CheckFreeMeAsync()
        {
            if (!Group.IsPet)
                return (PetOld.InfoClosed, 0, 0, 0);

            //以当前主人购买时的价格成交，对方只能得到80%，系统扣除20%
            long currMaster = PetOld.GetCurrMaster(Group.Id, UserId);
            if (currMaster == UserId)
                return ("您已是自由身，无需赎身", 0, 0, 0);

            long buyPrice = PetOld.GetBuyPrice(Group.Id, UserId);
            long creditAdd = buyPrice;
            long creditMinus = buyPrice * 12 / 10;
            if (User.IsSuper)
                creditMinus = creditMinus * 22 / 10;
            long creditValue = await UserInfo.GetCreditAsync(Group.Id, UserId);
            if (creditValue < creditMinus)
                return ($"您的积分{creditValue}不足{creditMinus}", 0, 0, 0);

            if (!IsConfirm)
                return (await ConfirmMessage($"赎身需扣分：-{creditMinus}"), 0, 0, 0);

            return (null, currMaster, creditAdd, creditMinus);
        }
    }
}
