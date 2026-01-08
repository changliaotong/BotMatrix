namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public class CommandResult
    {
        public bool Success { get; set; }
        public string? Message { get; set; } // ��ѡ��ʧ�ܻ���ʾ��Ϣ
        public Exception? Exception { get; set; } // ������쳣
        public static CommandResult Fail(string? message = null, Exception? ex = null)
        {
            return new CommandResult
            {
                Success = false,
                Message = message ?? ex?.Message ?? "����ʧ��",
                Exception = ex
            };
        }

        public static CommandResult Ok(string? message = null)
        {
            return new CommandResult
            {
                Success = true,
                Message = message ?? "�����ɹ�",
                Exception = null
            };
        }

        // ʵ��������ʽ�޸�
        public CommandResult SetSuccess(string? message = null)
        {
            Success = true;
            Message = message ?? "�����ɹ�";
            Exception = null;
            return this;
        }

        public CommandResult SetFail(string? message = null, Exception? ex = null)
        {
            Success = false;
            Message = message ?? ex?.Message ?? "����ʧ��";
            Exception = ex;
            return this;
        }
    }
    public static class CommandResultExtensions
    {
        /// <summary>
        /// �����ж��Ƿ�ʧ��
        /// </summary>
        public static bool IsFail(this CommandResult? result)
        {
            return result == null || !result.Success;
        }

        /// <summary>
        /// �����ж��Ƿ�ɹ�
        /// </summary>
        public static bool IsSuccess(this CommandResult? result)
        {
            return result != null && result.Success;
        }

        /// <summary>
        /// ʧ��ʱ������Ϣ���쳣������������
        /// </summary>
        public static CommandResult Fail(this CommandResult result, string? message = null, Exception? ex = null)
        {
            result.Success = false;
            result.Message = message ?? ex?.Message ?? "����ʧ��";
            result.Exception = ex;
            return result;
        }

        /// <summary>
        /// �ɹ�ʱ������Ϣ������������
        /// </summary>
        public static CommandResult Ok(this CommandResult result, string? message = null)
        {
            result.Success = true;
            result.Message = message ?? "�����ɹ�";
            result.Exception = null;
            return result;
        }

        /// <summary>
        /// ת��Ϊ��ʽ���ַ���
        /// </summary>
        public static string ToFormattedString(this CommandResult result)
        {
            return result.Success ? $"? �ɹ���{result.Message}" : $"? ʧ�ܣ�{result.Message ?? result.Exception?.Message}";
        }
    }

}


