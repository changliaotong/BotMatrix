namespace BotWorker.Common.Exts
{
    public class CommandResult
    {
        public bool Success { get; set; }
        public string? Message { get; set; } // 可选的失败或提示信息
        public Exception? Exception { get; set; } // 捕获的异常
        public static CommandResult Fail(string? message = null, Exception? ex = null)
        {
            return new CommandResult
            {
                Success = false,
                Message = message ?? ex?.Message ?? "操作失败",
                Exception = ex
            };
        }

        public static CommandResult Ok(string? message = null)
        {
            return new CommandResult
            {
                Success = true,
                Message = message ?? "操作成功",
                Exception = null
            };
        }

        // 实例方法链式修改
        public CommandResult SetSuccess(string? message = null)
        {
            Success = true;
            Message = message ?? "操作成功";
            Exception = null;
            return this;
        }

        public CommandResult SetFail(string? message = null, Exception? ex = null)
        {
            Success = false;
            Message = message ?? ex?.Message ?? "操作失败";
            Exception = ex;
            return this;
        }
    }
    public static class CommandResultExtensions
    {
        /// <summary>
        /// 快速判断是否失败
        /// </summary>
        public static bool IsFail(this CommandResult? result)
        {
            return result == null || !result.Success;
        }

        /// <summary>
        /// 快速判断是否成功
        /// </summary>
        public static bool IsSuccess(this CommandResult? result)
        {
            return result != null && result.Success;
        }

        /// <summary>
        /// 失败时设置消息和异常，并返回自身
        /// </summary>
        public static CommandResult Fail(this CommandResult result, string? message = null, Exception? ex = null)
        {
            result.Success = false;
            result.Message = message ?? ex?.Message ?? "操作失败";
            result.Exception = ex;
            return result;
        }

        /// <summary>
        /// 成功时设置消息，并返回自身
        /// </summary>
        public static CommandResult Ok(this CommandResult result, string? message = null)
        {
            result.Success = true;
            result.Message = message ?? "操作成功";
            result.Exception = null;
            return result;
        }

        /// <summary>
        /// 转换为格式化字符串
        /// </summary>
        public static string ToFormattedString(this CommandResult result)
        {
            return result.Success ? $"✅ 成功：{result.Message}" : $"❌ 失败：{result.Message ?? result.Exception?.Message}";
        }
    }

}
