namespace BotWorker.Domain.Entities
{
    public partial class BotCmd : MetaData<BotCmd>
    {
        public override string TableName => "Cmd";
        public override string KeyField => "Id";
        public static string GetRegexCmd()
        {
            var sql = $"SELECT CmdText FROM {FullName} WHERE IsClose = 0 ORDER BY LEN(CmdName) DESC";
            // 按长度降序排序
            var sortedCommands = QueryRes(sql, "{0}|").Trim('|')
                .Split('|')                           // 拆分为数组
                .OrderByDescending(cmd => cmd.Length) // 按长度降序排序
                .ToArray();

            // 重新拼接为正则表达式
            return @$"^[#＃﹟/／ ]*(?<cmdName>({string.Join('|', sortedCommands)}))[+ ]*(?<cmdPara>.*)";            
        }

        public static string GetCmdName(string cmdText)
        {
            return QueryScalar<string>($"SELECT CmdName FROM {FullName} WHERE CmdText LIKE '%|{cmdText}|%' OR CmdText LIKE '{cmdText}|%' OR CmdText LIKE '%|{cmdText}'") ?? "";
        }

        public static string GetClosedCmd()
        {
            string res = QueryRes($"SELECT CmdName FROM {FullName} WHERE IsClose = 1", "{0} ");
            return res == "" ? "没有功能被关闭" : res;
        }

        public static bool IsClosedCmd(long groupId, string message)
        {
            var regex = GroupInfo.GetClosedRegex(groupId).Trim();
            return !regex.IsNull() && message.RemoveQqAds().IsMatch(regex);
        }


        public static bool IsCmdCloseAll(string cmdName)
        {
            return GetWhere("IsClose", $"CmdName = {cmdName.Quotes()}").AsBool();
        }

        public static int SetCmdCloseAll(string cmdName, int isClose)
        {
            return UpdateWhere($"IsClose = {isClose}", $"CmdName = {cmdName.Quotes()}");
        }
    }
}
