namespace BotWorker.Domain.Entities;
public partial class GroupOffical : MetaData<GroupOffical>
{
    public override string TableName => "Group";
    public override string KeyField => "GroupOpenid";

    public const long MIN_GROUP_ID = 990000000000;
    public const long MAX_GROUP_ID = 1000000000000;

    public static (long groupId, bool isNew) GetGroupId(string groupOpenid, string groupName, long userId, long botUin = 0, string botName = "")
    {
        var groupId = GetTargetGroup(groupOpenid);
        if (groupId != 0)
            return (groupId, false);

        groupId = GetMaxGroupId();
        int i = GroupInfo.Append(groupId, groupName, botUin, botName, userId, userId, groupOpenid);
        return i == -1 ? (0, false) : (groupId, true);
    }

    private static long GetMaxGroupId()
    {
        var groupId = GetWhere<long>("max(Id)", $"Id > {MIN_GROUP_ID} and Id < {MAX_GROUP_ID}");
        return groupId <= MIN_GROUP_ID ? MIN_GROUP_ID + 1 : groupId + 1;
    }

    public static long GetTargetGroup(string groupOpenid)
    {
        return GetLong("isnull(TargetGroup, Id)", groupOpenid);
    }

    public static string GetGroupOpenid(long groupId, long botQQ)
    {
        var groupOpenid = GetValueAandB<string>("GroupOpenid", "TargetGroup", groupId, "BotUin", botQQ);
        if (!groupOpenid.IsNull())
            return groupOpenid;

        return GetValueAandB<string>("GroupOpenid", "Id", groupId, "BotUin", botQQ);
    }
}