namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        public async Task<string> GetConfirmNew()
        {
            string res = "";
            if (!Group.IsConfirmNew) return res;

            if (!GroupMember.GetBool("IsConfirm", RealGroupId, UserId))
            {
                string confirmCode = GroupMember.GetValue("ConfirmCode", RealGroupId, UserId);
                if (confirmCode.IsNull())
                {
                    confirmCode = C.RandomInt(100, 999).ToString();
                    int i = AddGroupMember(50, confirmCode);
                    if (i == -1)
                        return res;
                }
                else if (confirmCode == Message.Trim())
                {
                    return GroupMember.SetValue("IsConfirm", true, RealGroupId, UserId) == -1
                        ? RetryMsg
                        : $"[@:{UserId}]\n✅ 确认真人身份成功";
                }
                if (SelfPerm < UserPerm && SelfPerm < 2)
                {
                    //超过15分钟未确认的踢人
                    if (GroupMember.GetInt("ISNULL(ABS(DATEDIFF(MINUTE, GETDATE(), ConfirmDate)), 0)", RealGroupId, UserId) > 15)
                    {
                        await KickOutAsync(SelfId, GroupId, UserId);
                        return $"[@:{UserId}] 超过15分钟未确认，将被踢出群";
                    }
                }
                res = $"[@:{UserId}]，请回复【{confirmCode}】确认真人，否则T飞";
            }
            return res;
        }

        // 询问确认执行重要指令，例如删除数据，清空数据等，或者确认扣分
        public string ConfirmMessage(string confirmInfo)
        {
            IsCancelProxy = true;

            int confirmPassword = RandomInt(100, 499);
            int cancelPassword = confirmPassword + 499;

            return Confirm.Insert([
                new Cov("GroupId", RealGroupId),
                new Cov("GroupName", GroupName),
                new Cov("UserId", UserId),
                new Cov("Username", Name),
                new Cov("CmdName", CmdName),
                new Cov("CmdPara", CmdPara),
                new Cov("ConfirmInfo", confirmInfo),
                new Cov("ConfirmPassword", confirmPassword),
                new Cov("CancelPassword", cancelPassword),
            ]) == -1
                ? RetryMsg
                : $"{confirmInfo}\n回【{confirmPassword}】确认，回【{cancelPassword}】取消";
        }

        // 确认执行命令
        public async Task ConfirmCmdAsync()
        {
            string id = Confirm.GetWhere($"Id", $"GroupId = {RealGroupId} AND UserId = {UserId}", $"Id DESC");
            if (id == "")
                return;

            string confirmPassword = Confirm.GetValue("ConfirmPassword", id);
            string cancelPassword = Confirm.GetValue("CancelPassword", id);

            if (!Message.In(confirmPassword, cancelPassword))
                return;

            IsCancelProxy = true;

            CmdName = Confirm.GetValue("CmdName", id);
            CmdPara = Confirm.GetValue("CmdPara", id);

            switch (Confirm.Delete(id))
            {
                case -1:
                    Answer = RetryMsg;
                    return;
                default:
                    if (Message == confirmPassword)
                    {                        
                        Message = $"{CmdName}{CmdPara}";
                        IsConfirm = true;
                        await GetCmdResAsync();
                    }
                    else
                        Answer = "✅ 命令取消成功";
                    return;
            }
        }
}
