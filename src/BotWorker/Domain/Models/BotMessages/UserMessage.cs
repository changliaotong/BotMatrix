namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public string GetHeadCQ() => IsQQ && IsGroup ? UserService.GetHeadCQAsync(UserId).Result : "";

        public async Task<string> GetHintInfo()
        {
            if (User.IsDefaultHint)
            {
                string funcDefault = await UserService.GetStateResAsync(User.State);
                if (funcDefault != "闲聊")
                    return $"\n退出{funcDefault}请发 结束";
            }
            return "";
        }

        public bool IsRobotOwner(long userId) => Group.RobotOwner == userId;

        public bool IsRobotOwner() => Group.RobotOwner == UserId;

        public async Task<int> AddGroupMemberAsync(long groupCredit = 50, string confirmCode = "")
        {
            return await GroupMemberRepository.AppendAsync(GroupId, UserId, Name, DisplayName);
        }

        public async Task<int> AddClientAsync(long qqRef = 0)
        {
            int i = await UserService.AppendUserAsync(SelfId, GroupId, UserId, Name, qqRef);
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
                i = await FriendRepository.AppendAsync(SelfId, UserId, Name);
                if (i == -1)
                    return i;
            }

            return i;
        }
}
