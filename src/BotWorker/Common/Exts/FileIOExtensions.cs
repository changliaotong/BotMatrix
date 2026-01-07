namespace BotWorker.Common.Exts
{
    static class FileIOExtensions
    {
        /// <summary>
        /// 读取文本文件，支持默认编码
        /// </summary>
        public static string ReadTextFile(this string filePath)
            => File.Exists(filePath) ? File.ReadAllText(filePath) : string.Empty;

        /// <summary>
        /// 安全写文本文件（包含异常捕获）
        /// </summary>
        public static bool TryWriteTextFile(this string filePath, string content)
        {
            try
            {
                File.WriteAllText(filePath, content);
                return true;
            }
            catch
            {
                return false;
            }
        }
    }
}
