namespace BotWorker.Domain.Entities
{
    public partial class BotCmd : MetaData<BotCmd>
    {
        public override string TableName => "Cmd";
        public override string KeyField => "Id";
        private static readonly Dictionary<string, string> _baseCommandMap = new(StringComparer.OrdinalIgnoreCase)
        {
            { "帮助", "帮助" }, { "help", "帮助" }, { "指令", "帮助" },
            { "签到", "签到" }, { "checkin", "签到" },
            { "计算", "计算" }, { "calc", "计算" },
            { "钓鱼", "钓鱼" }, { "fish", "钓鱼" },
            { "抛竿", "抛竿" }, { "收竿", "收竿" },
            { "购买", "购买" }, { "buy", "购买" },
            { "买分", "买分" },
            { "我的宠物", "我的宠物" }, { "pet", "我的宠物" },
            { "我的待办", "todo" }, { "todo", "todo" }, { "td", "todo" },
            { "添加待办", "todo" },
            { "拍砖", "拍砖" },
            { "早安", "早安" }, { "午安", "午安" }, { "晚安", "晚安" },
            { "闲聊", "闲聊" }, { "chat", "闲聊" }, { "ai", "闲聊" },
            { "成语", "成语" },
            { "点歌", "点歌" },
        };

        private static readonly HashSet<string> _extraCommandKeywords = new(StringComparer.OrdinalIgnoreCase);

        public static void RegisterExtraCommands(IEnumerable<string> commands)
        {
            foreach (var cmd in commands)
            {
                _extraCommandKeywords.Add(cmd);
            }
        }

        public static string GetRegexCmd()
        {
            var sql = $"SELECT {Quote("CmdText")} FROM {FullName} WHERE {Quote("IsClose")} = 0 ORDER BY {SqlLen(Quote("CmdText"))} DESC";
            var res = QueryRes(sql, "{0}|").Trim('|');
            
            var dbCommands = string.IsNullOrEmpty(res) ? Array.Empty<string>() : res.Split('|');
            var allCommands = dbCommands
                .Concat(_baseCommandMap.Keys)
                .Concat(_extraCommandKeywords)
                .Distinct()
                .OrderByDescending(cmd => cmd.Length)
                .ToArray();

            return @$"^[#＃﹟/／ ]*(?<cmdName>({string.Join('|', allCommands)}))\s*(?<cmdPara>.*)";            
        }

        public static string GetCmdName(string cmdText)
        {
            if (string.IsNullOrEmpty(cmdText)) return "";

            // 优先从基础命令映射中查找
            if (_baseCommandMap.TryGetValue(cmdText, out var baseName))
                return baseName;

            // 再从数据库中查找
            return QueryScalar<string>($"SELECT CmdName FROM {FullName} WHERE CmdText = {cmdText.Quotes()} OR CmdText LIKE '%|{cmdText}|%' OR CmdText LIKE '{cmdText}|%' OR CmdText LIKE '%|{cmdText}'") ?? "";
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

        public static async Task EnsureCommandExistsAsync(string name, string text)
        {
            try
            {
                if (!await ExistsWhereAsync($"CmdName = {name.Quotes()}"))
                {
                    Console.WriteLine($"[BotCmd] Command '{name}' not found, inserting...");
                    await InsertAsync(new List<Cov>
                    {
                        new Cov("CmdName", name),
                        new Cov("CmdText", text),
                        new Cov("IsClose", 0)
                    });
                    Console.WriteLine($"[BotCmd] Command '{name}' inserted.");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[BotCmd] Error ensuring command '{name}': {ex.Message}");
            }
        }

        public static int SetCmdCloseAll(string cmdName, int isClose)
        {
            return UpdateWhere($"IsClose = {isClose}", $"CmdName = {cmdName.Quotes()}");
        }
    }
}
