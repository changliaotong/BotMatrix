namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
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
                    confirmCode = RandomInt(100, 999).ToString();
                    int i = await AddGroupMemberAsync(50, confirmCode);
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
        public async Task<string> ConfirmMessage(string confirmInfo)
        {
            IsCancelProxy = true;

            var sessionManager = PluginManager?.Sessions;
            if (sessionManager == null) return RetryMsg;

            // 使用 SessionManager 生成验证码并存储会话
            var code = await sessionManager.StartConfirmationAsync(UserId.ToString(), RealGroupId.ToString(), "system", CmdName, new
            {
                CmdName,
                CmdPara,
                ConfirmInfo = confirmInfo
            });

            return $"{confirmInfo}\n请输入验证码【{code}】以确认，或发送“取消”退出。";
        }

        // 确认执行命令
        public async Task ConfirmCmdAsync()
        {
            var sessionManager = PluginManager?.Sessions;
            if (sessionManager == null) return;

            var session = await sessionManager.GetSessionAsync(UserId.ToString(), RealGroupId.ToString());
            if (session == null || string.IsNullOrEmpty(session.ConfirmationCode))
                return;

            if (Message == "取消")
            {
                await sessionManager.ClearSessionAsync(UserId.ToString(), RealGroupId.ToString());
                Answer = "✅ 已取消当前操作。";
                IsCancelProxy = true;
                return;
            }

            if (Message != session.ConfirmationCode)
                return;

            IsCancelProxy = true;

            var data = session.GetData<dynamic>();
            if (data != null)
            {
                CmdName = data.GetProperty("CmdName").GetString();
                CmdPara = data.GetProperty("CmdPara").GetString();
            }

            await sessionManager.ClearSessionAsync(UserId.ToString(), RealGroupId.ToString());
            
            Message = $"{CmdName}{CmdPara}";
            IsConfirm = true;
            await GetCmdResAsync();
        }
}
