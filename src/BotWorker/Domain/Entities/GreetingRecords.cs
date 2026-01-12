
namespace BotWorker.Domain.Entities
{
    public class GreetingRecords : MetaData<GreetingRecords>
    {
        public override string TableName => "GreetingRecords";
        public override string KeyField => "Id";

        public static int Append(long botQQ, long groupId, string groupName, long qq, string name, int greetingType = 0)
        {
            return Insert([ 
                            new Cov("BotQQ", botQQ),
                            new Cov("GroupId", groupId),
                            new Cov("GroupName", groupName),
                            new Cov("QQ", qq),
                            new Cov("Name", name),
                            new Cov("GreetingType", greetingType),
                            new Cov("LogicalDate", QueryScalar<DateTime>($"SELECT CONVERT(date, DATEADD(HOUR, {(greetingType == 0 ? -3 : -5)}, GETDATE()))")),
                        ]);
        }

        public static bool Exists(long groupId, long qq, int greetingType = 0)
        {
            var sql = $"SELECT TOP 1 1 FROM {FullName} WHERE GroupId = {groupId} AND QQ = {qq} AND GreetingType = {greetingType} AND LogicalDate = Convert(date, DATEADD(HOUR, {(greetingType == 0 ? -3 : -5)}, GETDATE()))";
            return QueryScalar<int>(sql).AsBool();
        }

        //全服第x位起床用户
        public static int GetCount(int greetingType = 0)
        {
            var minus = greetingType == 0 ? -3 : -5;
            var sql = $"SELECT COUNT(Id)+1 FROM {FullName} WHERE GreetingType = {greetingType} AND LogicalDate = Convert(date, DATEADD(HOUR, {(greetingType == 0 ? -3 : -5)}, GETDATE()))";
            return QueryScalar<int>(sql);
        }

        //本群第x位起床用户
        public static int GetCount(long groupId, int greetingType = 0)
        {
            var sql = $"SELECT COUNT(Id)+1 FROM {FullName} WHERE GroupId = {groupId} AND GreetingType = {greetingType} AND LogicalDate = Convert(date, DATEADD(HOUR,{(greetingType == 0 ? -3 : -5)}, GETDATE()))";
            return QueryScalar<int>(sql);
        }
    }
}
