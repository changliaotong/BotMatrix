using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Entries
{
    public partial class BotCmd : MetaData<BotCmd>
    {
        public override string TableName => "Cmd";
        public override string KeyField => "Id";
        public static string GetRegexCmd()
        {            
            try 
            {
                var sql = $"SELECT CmdText FROM {FullName} WHERE IsClose = 0 ORDER BY LEN(CmdName) DESC";
                // 按长度降序排序
                var res = QueryRes(sql, "{0}|").Trim('|');
                if (string.IsNullOrEmpty(res)) return DefaultRegex;

                var sortedCommands = res.Split('|')                           // 拆分为数组
                    .OrderByDescending(cmd => cmd.Length) // 按长度降序排序
                    .ToArray();

                // 重新拼接为正则表达式
                return @$"^[#＃﹟/／ ]*(?<cmdName>({string.Join('|', sortedCommands)}))[+ ]*(?<cmdPara>.*)";
            }
            catch
            {
                return DefaultRegex;
            }
        }

        private static string DefaultRegex => @"^[#＃﹟/／ ]*(?<cmdName>(菜单|你好|帮助|关于|状态))[+ ]*(?<cmdPara>.*)";

        public static string GetCmdName(string cmdText)
        {
            return Query($"SELECT CmdName FROM {FullName} WHERE CmdText LIKE '%|{cmdText}|%' OR CmdText LIKE '{cmdText}|%' OR CmdText LIKE '%|{cmdText}'");
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
