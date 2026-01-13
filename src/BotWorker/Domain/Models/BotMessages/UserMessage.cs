namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage: MetaData<BotMessage>
{
        public string GetHeadCQ() => IsQQ && IsGroup ? UserInfo.GetHeadCQ(UserId) : "";

        public async Task<string> GetHintInfo()
        {
            if (User.IsDefaultHint)
            {
                string funcDefault = await UserInfo.GetStateResAsync(User.State);
                if (funcDefault != "闲聊")
                    return $"\n退出{funcDefault}请发 结束";
            }
            return "";
        }

        public bool IsRobotOwner(long userId) => Group.RobotOwner == userId;

        public bool IsRobotOwner() => Group.RobotOwner == UserId;

        public int AddGroupMember(long groupCredit = 50, string confirmCode = "") => AddGroupMemberAsync(groupCredit, confirmCode).GetAwaiter().GetResult();

        public async Task<int> AddGroupMemberAsync(long groupCredit = 50, string confirmCode = "")
        {
            return await GroupMember.AppendAsync(GroupId, UserId, Name, DisplayName);
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
}
