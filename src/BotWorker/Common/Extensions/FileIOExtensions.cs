namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    static class FileIOExtensions
    {
        /// <summary>
        /// ��ȡ�ı��ļ���֧��Ĭ�ϱ���
        /// </summary>
        public static string ReadTextFile(this string filePath)
            => File.Exists(filePath) ? File.ReadAllText(filePath) : string.Empty;

        /// <summary>
        /// ��ȫд�ı��ļ��������쳣����
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


