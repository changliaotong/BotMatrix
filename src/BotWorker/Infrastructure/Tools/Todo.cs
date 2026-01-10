using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Infrastructure.Tools
{
    public class Todo : MetaData<Todo>
    {
        public override string TableName => "Todo";
        public override string KeyField => "Id";
        public const string format = "todo + 内容 新增\ntodo - 数字 删除\ntodo + #关键字 查询";

        // todo
        public static string GetTodoRes(long groupId, string groupName, long qq, string name, string cmdName, string cmdPara)
        {
            if (cmdName == "+" || cmdName.IsNull() && !cmdPara.IsNull())
            {
                if (cmdPara.StartsWith('#'))
                {
                    cmdPara = cmdPara[1..];
                    if (cmdPara.IsNum())
                    {
                        string res = GetTodo(qq, cmdPara.AsInt());
                        if (res == "")
                            return $"没有#{cmdPara}";
                        else
                            return res;
                    }
                    else
                        return GetTodos(qq, cmdPara);
                }
                return Append(groupId, qq, cmdPara, out int todo_no) == -1
                    ? RetryMsg
                    : $"添加成功，#{todo_no}";
            }
            else if (cmdName == "-")
            {
                if (cmdPara.StartsWith("#")) cmdPara = cmdPara[1..];
                if (cmdPara.IsNum())
                {
                    int todo_no = cmdPara.AsInt();
                    string sWhere = $"UserId = {qq} and TodoNo = {todo_no}";
                    if (!ExistsWhere(sWhere))
                        return $"不存在#{todo_no}";

                    return DeleteWhere(sWhere) == -1
                        ? RetryMsg
                        : $"成功删除#{todo_no}";
                }
                else if (cmdPara == "all")
                {
                    return DeleteWhere($"UserId = {qq}") == -1
                        ? RetryMsg
                        : $"真棒！全部完成！";
                }
                else
                    return "删除参数只能是数字";
            }
            return GetTodos(qq, cmdPara);
        }

        // 新增todo
        public static int Append(long groupId, long qq, string cmdPara, out int todoNo)
        {
            todoNo = GetWhere("max(TodoNo)+1", $"UserId = {qq}").AsInt();
            todoNo = todoNo == 0 ? 1 : todoNo;
            return Insert([
                            new Cov("GroupId", groupId),
                            new Cov("UserId", qq),
                            new Cov("TodoNo", todoNo),
                            new Cov("TodoTitle", cmdPara),
                        ]);
        }

        // todo 数量
        public static long Count(long qq)
        {
            return CountWhere($"UserId = {qq}");
        }

        // 得到 todoNo 的 todo
        public static string GetTodo(long qq, int todoNo)
        {
            string sql = $"select top 1 TodoNo, TodoTitle from {FullName} where UserId = {qq} and TodoNo = {todoNo}";
            return QueryRes(sql, "#{0}:{1}");
        }

        // todo 列表
        public static string GetTodos(long qq, string cmdPara, int topN = 5)
        {
            string res = QueryWhere($"top {topN} TodoNo, substring(TodoTitle, 1, 20)", $"UserId = {qq} {(cmdPara == "" ? "" : $" and TodoTitle like '%cmdPara%'")}", "TodoNo desc", "{0} {1}\n", $"{{c}}/{Count(qq)}");
            return res.IsNull() ? $"太好了，没有todo" : res;
        }
    }
}
