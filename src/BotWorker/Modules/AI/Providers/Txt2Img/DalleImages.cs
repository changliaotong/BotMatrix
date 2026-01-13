namespace BotWorker.Modules.AI.Providers.Txt2Img
{
    public class DalleImages : MetaData<DalleImages>
    {
        public override string DataBase => "robot_images";
        public override string TableName => "dalle_images";
        public override string KeyField => "id";

        private static string _localImagePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "Images");
        private static readonly HttpClient httpClient = new();

        public static async Task<int> SaveImageAsync(long botUin, long groupId, string groupName, long qq, string name, string prompt, string prompt_gpt, string url)
        {
            try
            {
                byte[] data = await httpClient.GetByteArrayAsync(url);

                string fileName = GenerateFileName();
                string folderPath = GenerateFolderPath();
                string imagePath = Path.Combine(folderPath, fileName);

                await File.WriteAllBytesAsync(imagePath, data);

                return Insert([
                    new Cov("robot_qq", botUin),
                    new Cov("group_id", groupId),
                    new Cov("group_name", groupName),
                    new Cov("qq", qq),
                    new Cov("name", name),
                    new Cov("prompt", prompt),
                    new Cov("prompt_gpt", prompt_gpt),
                    new Cov("url", url),
                    new Cov("data", data),
                    new Cov("path", imagePath),
                ]);
            }
            catch (Exception ex)
            {
                Logger.Error($"An error occurred while saving image: {ex.Message}");
                return -1;
            }
        }

        public static async Task<string> GetImagePathAsync(string imageUrl)
        {
            string fileName = GetFileNameFromUrl(imageUrl);
            string localFilePath = Path.Combine(_localImagePath, fileName);

            if (!File.Exists(localFilePath))
            {
                byte[]? imageBytes = await GetImageFromDatabaseAsync(imageUrl);
                if (imageBytes != null)
                {
                    Directory.CreateDirectory(Path.GetDirectoryName(localFilePath) ?? _localImagePath);
                    await File.WriteAllBytesAsync(localFilePath, imageBytes);
                }
            }

            return localFilePath;
        }

        private static async Task<byte[]?> GetImageFromDatabaseAsync(string imageUrl)
        {
            var imageId = GetWhere("ImageId", $"ImageUrl={imageUrl}", "ImageId desc");
            return await GetBytes("ImageData", imageId);
        }

        private static string GenerateFolderPath()
        {
            DateTime now = DateTime.Now;
            string folderPath = Path.Combine(_localImagePath, now.ToString("yyyy"), now.ToString("MM"));
            Directory.CreateDirectory(folderPath);
            return folderPath;
        }

        private static string GenerateFileName()
        {
            // 获取当前日期
            DateTime now = DateTime.Now;
            // 生成文件名（例如：20240828_001.png）
            string fileName = $"{now:yyyyMMdd}_{Guid.NewGuid().ToString("N")[..3]}.png";
            return fileName;
        }

        private static string GetFileNameFromUrl(string imageUrl)
        {
            // Generate a file name based on the image URL
            return Path.GetFileName(imageUrl);
        }

    }
}
