using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class ExceptionExtensions
    {
        public static void TryCatch(this Action action, Action<Exception>? onError = null)
        {
            try { action(); }
            catch (Exception ex) { onError?.Invoke(ex); }
        }

        public static T? TryCatch<T>(this Func<T> func, Func<Exception, T>? onError = null)
        {
            try { return func(); }
            catch (Exception ex) { return onError != null ? onError(ex) : default; }
        }

        // 获取异常的完整信息（支持内部异常递归�?
        public static string FullMessage(this Exception ex)
        {
            if (ex.InnerException == null) return ex.Message;
            return $"{ex.Message} --> {ex.InnerException.FullMessage()}";
        }

        // 转为一行（便于日志记录�?
        public static string Flatten(this Exception ex)
        {
            return ex.ToString().Replace(Environment.NewLine, " ");
        }
    }
}


