namespace BotWorker.Modules.AI.Providers.Txt2Img
{
    public class DalleImages : MetaData<DalleImages>
    {
        public override string DataBase => "robot_images";
        public override string TableName => "dalle_images";
        public override string KeyField => "id";

        private readonly string _localImagePath = @"E:\Images";
        private static readonly HttpClient httpClient = new();

        public static async Task<int> SaveImageAsync(long botUin, long groupId, string groupName, long qq, string name, string prompt, string prompt_gpt, string url)
        {
            try
            {
                // 下载图片
                byte[] data = await httpClient.GetByteArrayAsync(url);

                // 生成保存路径和文件名
                string fileName = GenerateFileName();
                string folderPath = GenerateFolderPath();
                string imagePath = Path.Combine(folderPath, fileName);

                // 保存图片到文件系统
                await File.WriteAllBytesAsync(imagePath, data);

                // 保存图片信息到数据库
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
                Debug($"An error occurred while saving image: {ex.Message}");
                Console.WriteLine(ex.ToString());
                return -1;
            }

        }

        private static async Task<byte[]?> GetImageFromDatabaseAsync(string imageUrl)
        {
            var imageId = GetWhere("ImageId", $"ImageUrl={imageUrl}", "ImageId desc");
            return await GetBytes("ImageData", imageId);
        }


        public async Task<string> GetImagePathAsync(string imageUrl)
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

        private static string GenerateFolderPath()
        {
            // 获取当前日期
            DateTime now = DateTime.Now;
            // 生成按年月组织的文件夹路径
            string folderPath = Path.Combine("E:\\", "Images", now.ToString("yyyy"), now.ToString("MM"));
            // 确保文件夹存在
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
