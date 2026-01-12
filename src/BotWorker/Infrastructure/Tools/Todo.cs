using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Infrastructure.Tools
{
    public class Todo : MetaData<Todo>
    {
        public override string TableName => "Todo";
        public override string KeyField => "Id";

        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public long Id { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public int TodoNo { get; set; }
        public string TodoTitle { get; set; } = "";
        public string Description { get; set; } = ""; // 详细描述
        public string Priority { get; set; } = "Medium"; // Low, Medium, High
        public int Progress { get; set; } = 0; // 进度 0-100
        public string Status { get; set; } = "Pending"; // Pending, InProgress, Completed
        public string Category { get; set; } = "Todo"; // Todo, Dev, Test
        public DateTime? DueDate { get; set; } // 截止日期
        public DateTime InsertDate { get; set; } = DateTime.Now;

        public static new async Task EnsureTableCreatedAsync()
        {
            await MetaData<Todo>.EnsureTableCreatedAsync();

            // 确保 Todo 表结构完整 (支持旧表升级)
            await SQLConn.ExecAsync(@"
                IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID('Todo') AND name = 'Description')
                    ALTER TABLE Todo ADD Description NVARCHAR(MAX) DEFAULT '' NOT NULL;
                IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID('Todo') AND name = 'Priority')
                    ALTER TABLE Todo ADD Priority NVARCHAR(20) DEFAULT 'Medium' NOT NULL;
                IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID('Todo') AND name = 'Progress')
                    ALTER TABLE Todo ADD Progress INT DEFAULT 0 NOT NULL;
                IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID('Todo') AND name = 'Status')
                    ALTER TABLE Todo ADD Status NVARCHAR(50) DEFAULT 'Pending' NOT NULL;
                IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID('Todo') AND name = 'Category')
                    ALTER TABLE Todo ADD Category NVARCHAR(100) DEFAULT 'Todo' NOT NULL;
                IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID('Todo') AND name = 'DueDate')
                    ALTER TABLE Todo ADD DueDate DATETIME NULL;
            ");
        }

        public const string format = "todo + 内容 [Dev/Test] [P1/P2/P3] 新增\ntodo - 数字 删除\ntodo #数字 [进度/done/P1/P2/P3/desc 内容] 更新\ntodo + #关键字 查询";

        // todo
        public static string GetTodoRes(long groupId, string groupName, long qq, string name, string cmdName, string cmdPara)
        {
            Console.WriteLine($"[Todo] GetTodoRes: cmdName='{cmdName}', cmdPara='{cmdPara}'");
            cmdName = cmdName.ToLower();
            if (cmdName == "todo" || cmdName == "td")
            {
                if (cmdPara.StartsWith("+"))
                {
                    cmdName = "+";
                    cmdPara = cmdPara[1..].Trim();
                }
                else if (cmdPara.StartsWith("-"))
                {
                    cmdName = "-";
                    cmdPara = cmdPara[1..].Trim();
                }
                else if (cmdPara.StartsWith("#"))
                {
                    // 保持 cmdName 为 todo，后面会处理 # 开头的 cmdPara
                }
                else if (cmdPara.IsNull())
                {
                    // 列表查询
                }
                else
                {
                    // 默认当作搜索
                    return GetTodos(qq, cmdPara);
                }
            }
            Console.WriteLine($"[Todo] Processed: cmdName='{cmdName}', cmdPara='{cmdPara}'");

            if (cmdName == "+" || cmdName.IsNull() && !cmdPara.IsNull() && !cmdPara.StartsWith('#'))
            {
                // 支持 todo + 内容 Dev P1 或 todo + 内容 Test
                string category = "Todo";
                string priority = "Medium";

                if (cmdPara.Contains(" Dev", StringComparison.OrdinalIgnoreCase))
                {
                    category = "Dev";
                    cmdPara = cmdPara.Replace(" Dev", "", StringComparison.OrdinalIgnoreCase).Trim();
                }
                else if (cmdPara.Contains(" Test", StringComparison.OrdinalIgnoreCase))
                {
                    category = "Test";
                    cmdPara = cmdPara.Replace(" Test", "", StringComparison.OrdinalIgnoreCase).Trim();
                }

                if (cmdPara.Contains(" P1", StringComparison.OrdinalIgnoreCase)) { priority = "High"; cmdPara = cmdPara.Replace(" P1", "", StringComparison.OrdinalIgnoreCase).Trim(); }
                else if (cmdPara.Contains(" P2", StringComparison.OrdinalIgnoreCase)) { priority = "Medium"; cmdPara = cmdPara.Replace(" P2", "", StringComparison.OrdinalIgnoreCase).Trim(); }
                else if (cmdPara.Contains(" P3", StringComparison.OrdinalIgnoreCase)) { priority = "Low"; cmdPara = cmdPara.Replace(" P3", "", StringComparison.OrdinalIgnoreCase).Trim(); }

                return Append(groupId, qq, cmdPara, category, priority, out int todo_no) == -1
                    ? RetryMsg
                    : $"待办添加成功，#{todo_no} [{category}] [{priority}]";
            }
            else if (cmdName == "-" || cmdName == "删除")
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
            else if (cmdPara.StartsWith('#'))
            {
                var parts = cmdPara.Split(' ', StringSplitOptions.RemoveEmptyEntries);
                if (parts.Length >= 2 && parts[0].StartsWith('#'))
                {
                    // 更新进度或属性: todo #1 50 或 todo #1 done 或 todo #1 P1
                    string todoNoStr = parts[0][1..];
                    if (todoNoStr.IsNum())
                    {
                        int todo_no = todoNoStr.AsInt();
                        string updateVal = parts[1].ToUpper();

                        if (updateVal == "DONE")
                        {
                            return UpdateWhere(new { Progress = 100, Status = "Completed" }, $"UserId = {qq} and TodoNo = {todo_no}") == -1
                                ? RetryMsg
                                : $"待办#{todo_no}已完成！";
                        }
                        else if (updateVal == "DESC")
                        {
                            string desc = string.Join(" ", parts.Skip(2));
                            return UpdateWhere(new { Description = desc }, $"UserId = {qq} and TodoNo = {todo_no}") == -1
                                ? RetryMsg
                                : $"待办#{todo_no}描述已更新";
                        }
                        else if (updateVal == "P1") return UpdateWhere(new { Priority = "High" }, $"UserId = {qq} and TodoNo = {todo_no}") == -1 ? RetryMsg : $"待办#{todo_no}优先级已更新为 High";
                        else if (updateVal == "P2") return UpdateWhere(new { Priority = "Medium" }, $"UserId = {qq} and TodoNo = {todo_no}") == -1 ? RetryMsg : $"待办#{todo_no}优先级已更新为 Medium";
                        else if (updateVal == "P3") return UpdateWhere(new { Priority = "Low" }, $"UserId = {qq} and TodoNo = {todo_no}") == -1 ? RetryMsg : $"待办#{todo_no}优先级已更新为 Low";
                        else if (updateVal.IsNum())
                        {
                            int progress = updateVal.AsInt();
                            progress = Math.Clamp(progress, 0, 100);
                            string status = progress == 100 ? "Completed" : (progress > 0 ? "InProgress" : "Pending");
                            return UpdateWhere(new { Progress = progress, Status = status }, $"UserId = {qq} and TodoNo = {todo_no}") == -1
                                ? RetryMsg
                                : $"待办#{todo_no}进度已更新为 {progress}%";
                        }
                    }
                }
                else
                {
                    // 查询特定ID: todo #1
                    string todoNoStr = cmdPara[1..];
                    if (todoNoStr.IsNum())
                    {
                        string res = GetTodo(qq, todoNoStr.AsInt());
                        return res == "" ? $"没有#{todoNoStr}" : res;
                    }
                }
            }
            return GetTodos(qq, cmdPara);
        }

        // 新增todo
        public static int Append(long groupId, long qq, string cmdPara, string category, string priority, out int todoNo)
        {
            todoNo = QueryScalar<int>($"select max(TodoNo)+1 from {FullName} where UserId = {qq}");
            todoNo = todoNo == 0 ? 1 : todoNo;
            return Insert(new
            {
                GroupId = groupId,
                UserId = qq,
                TodoNo = todoNo,
                TodoTitle = cmdPara,
                Category = category,
                Priority = priority,
                Progress = 0,
                Status = "Pending"
            });
        }

        // todo 数量
        public static long Count(long qq)
        {
            return CountWhere($"UserId = {qq}");
        }

        // 得到 todoNo 的 todo
        public static string GetTodo(long qq, int todoNo)
        {
            string sql = $"select {SqlTop(1)} TodoNo, Category, Priority, Progress, Status, TodoTitle, Description from {FullName} where UserId = {qq} and TodoNo = {todoNo} {SqlLimit(1)}";
            return QueryRes(sql, "#{0} [{1}] [{2}] {3}% {4} {5}\n描述: {6}");
        }

        // todo 列表
        public static string GetTodos(long qq, string cmdPara, int topN = 5)
        {
            string sWhere = $"UserId = {qq}";
            if (!string.IsNullOrEmpty(cmdPara))
            {
                if (cmdPara.Equals("dev", StringComparison.OrdinalIgnoreCase)) sWhere += " and Category = 'Dev'";
                else if (cmdPara.Equals("test", StringComparison.OrdinalIgnoreCase)) sWhere += " and Category = 'Test'";
                else if (cmdPara.Equals("p1", StringComparison.OrdinalIgnoreCase)) sWhere += " and Priority = 'High'";
                else sWhere += $" and TodoTitle like '%{cmdPara}%'";
            }

            string columns = $"{SqlTop(topN)} TodoNo, Category, Priority, Progress, {SqlIsNull("substring(TodoTitle, 1, 20)", "''")}";
            string res = QueryWhere(columns, sWhere, "TodoNo desc", "#{0} [{1}] [{2}] [{3}%] {4}\n", $"{{c}}/{Count(qq)}");
            return res.IsNull() ? $"太好了，没有todo" : res;
        }
    }
}
