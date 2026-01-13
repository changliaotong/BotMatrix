namespace BotWorker.Domain.Entities
{
    public class WhiteList : MetaData<WhiteList>
    {
        //白名单系统
        public override string TableName => "WhiteList";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "WhiteId";


        // 加入白名单
        public static async Task<int> AppendWhiteListAsync(long botUin, long groupId, string groupName, long qq, string name, long qqWhite)
        {
            return await ExistsAsync(groupId, qqWhite) 
                ? 0 
                : await InsertAsync([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("GroupName", groupName),
                            new Cov("UserId", qq),
                            new Cov("UserName", name),
                            new Cov("WhiteId", qqWhite),
                        ]);
        }

        public static int AppendWhiteList(long botUin, long groupId, string groupName, long qq, string name, long qqWhite)
        {
            return AppendWhiteListAsync(botUin, groupId, groupName, qq, name, qqWhite).GetAwaiter().GetResult();
        }





    }
}
