namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage: MetaData<BotMessage>
{
        public string GetHeadCQ() => IsQQ && IsGroup ? UserInfo.GetHeadCQ(UserId) : "";

        public string GetHintInfo()
        {
            if (User.IsDefaultHint)
            {
                string funcDefault = UserInfo.GetStateRes(User.State);
                if (funcDefault != "闲聊")
                    return $"\n退出{funcDefault}请发 结束";
            }
            return "";
        }

        public bool IsRobotOwner(long userId) => Group.RobotOwner == userId;

        public bool IsRobotOwner() => Group.RobotOwner == UserId;

        public async Task<int> AddGroupMemberAsync(long groupCredit = 50, string confirmCode = "")
        {
            return await GroupMember.AppendAsync(GroupId, UserId, Name, DisplayName);
        }

        public int AddGroupMember(long groupCredit = 50, string confirmCode = "")
        {
            return AddGroupMemberAsync(groupCredit, confirmCode).GetAwaiter().GetResult();
        }

        public async Task<int> AddClientAsync(long qqRef)
        {
            int i = await UserInfo.AppendAsync(SelfId, GroupId, UserId, Name, qqRef);
            if (i == -1)
                return i;

            if (Group.IsCredit)
            {
                i = await AddGroupMemberAsync();
                if (i == -1)
                    return i;
            }

            if (SelfInfo.IsCredit)
            {
                i = await Friend.AppendAsync(SelfId, UserId, Name);
                if (i == -1)
                    return i;
            }

            return i;
        }

        public int AddClient()
        {
            return AddClientAsync(GroupInfo.GetGroupOwner(GroupId)).GetAwaiter().GetResult();
        }

        public int AddClient(long qqRef)
        {
            return AddClientAsync(qqRef).GetAwaiter().GetResult();
        }
}
