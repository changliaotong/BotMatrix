using Azure.AI.OpenAI;
using Azure;
using OpenAI.Images;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Providers.Configs;
using System.ClientModel;

namespace BotWorker.Modules.AI.Providers.Txt2Img
{
    public class Dalle3 : MetaData<Dalle3>, ITxt2ImgProvider
    {
        public override string TableName => "robot_dalle";
        public override string KeyField => "dalle_id";

        public string ProviderName => "Dall-E 3";

        public async Task<string> GenerateImageAsync(string prompt, BotWorker.Modules.AI.Interfaces.ImageGenerationOptions options)
        {
            var imageSize = options.Size switch
            {
                "1024x1024" => GeneratedImageSize.W1024xH1024,
                "1024x1792" => GeneratedImageSize.W1024xH1792,
                "1792x1024" => GeneratedImageSize.W1792xH1024,
                _ => GeneratedImageSize.W1024xH1024
            };

            var quality = options.Quality?.ToLower() == "hd" ? GeneratedImageQuality.High : GeneratedImageQuality.Standard;
            var style = options.Style?.ToLower() == "vivid" ? GeneratedImageStyle.Vivid : GeneratedImageStyle.Natural;

            return await GenerateImageInternalAsync(prompt, imageSize, quality, style);
        }

        private static async Task<string> GenerateImageInternalAsync(string prompt, GeneratedImageSize imageSize, GeneratedImageQuality quality, GeneratedImageStyle style)
        {
            string endpoint = AzureDalle.Endpoint;
            string key = AzureDalle.ApiKey;

            AzureOpenAIClient client = new(new Uri(endpoint), new ApiKeyCredential(key));

            try
            {
                var imageClient = client.GetImageClient(AzureDalle.DeploymentName);

                var res = await imageClient.GenerateImageAsync(prompt, new OpenAI.Images.ImageGenerationOptions
                {
                    Size = imageSize,
                    Quality = quality,
                    Style = style,
                    ResponseFormat = GeneratedImageFormat.Uri
                });

                return res.Value.ImageUri.ToString();
            }
            catch (Exception ex)
            {
                Debug($"Error: {ex.Message}", "Dall-E 3");
                return string.Empty;
            }
        }
    }
}
