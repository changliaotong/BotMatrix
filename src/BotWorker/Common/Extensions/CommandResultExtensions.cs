namespace BotWorker.Common.Extensions
{
    public class OperationResult
    {
        public bool Success { get; set; }
        public string? Message { get; set; } 
        public Exception? Exception { get; set; } 
        public static OperationResult Fail(string? message = null, Exception? ex = null)
        {
            return new OperationResult
            {
                Success = false,
                Message = message ?? ex?.Message ?? "操作失败",
                Exception = ex
            };
        }

        public static OperationResult Ok(string? message = null)
        {
            return new OperationResult
            {
                Success = true,
                Message = message ?? "操作成功",
                Exception = null
            };
        }

        public OperationResult SetSuccess(string? message = null)
        {
            Success = true;
            Message = message ?? "操作成功";
            Exception = null;
            return this;
        }

        public OperationResult SetFail(string? message = null, Exception? ex = null)
        {
            Success = false;
            Message = message ?? ex?.Message ?? "操作失败";
            Exception = ex;
            return this;
        }
    }
    public static class OperationResultExtensions
    {
        public static bool IsFail(this OperationResult? result)
        {
            return result == null || !result.Success;
        }

        public static bool IsOk(this OperationResult? result)
        {
            return result != null && result.Success;
        }
    }
}
