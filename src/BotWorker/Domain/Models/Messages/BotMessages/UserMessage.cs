namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage: MetaData<BotMessage>
{
        public string GetHeadCQ() => IsNapCat && IsGroup ? UserInfo.GetHeadCQ(UserId) : "";

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

        public int AddGroupMember(long groupCredit = 50, string confirmCode = "")
        {
            return GroupMember.Append(GroupId, UserId, Name, DisplayName, groupCredit, confirmCode);
        }

        public int AddClient()
        {
            return AddClient(GroupInfo.GetGroupOwner(GroupId));
        }

        public int AddClient(long qqRef)
        {
            int i = UserInfo.Append(SelfId, GroupId, UserId, Name, qqRef);
            if (i == -1)
                return i;

            if (Group.IsCredit)
            {
                i = AddGroupMember();
                if (i == -1)
                    return i;
            }

            if (SelfInfo.IsCredit)
            {
                i = Friend.Append(SelfId, UserId, Name);
                if (i == -1)
                    return i;
            }

            return i;
    }
}
