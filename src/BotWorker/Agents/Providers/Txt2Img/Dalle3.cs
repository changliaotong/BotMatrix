using Azure.AI.OpenAI;
using Azure;
using OpenAI.Images;
using BotWorker.Core;
using BotWorker.Agents.Entries;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Agents.Providers.Txt2Img
{
    public class Dalle3 : MetaData<Dalle3>
    {
        public override string TableName => "robot_dalle";
        public override string KeyField => "dalle_id";

        public static string GenerateImage(long qq, string prompt)
        {
            var res = "";

            //if (IsPublic)
            //{
            //    var url = $"< a href =\"http://robot.pengguanghui.com/ai?t={Token.GetToken(qq)}";
            //    return $"以下地址直接<a href=\"{url}\">进入后台</a>使用文生图:\n{url}";
            //}
            //else if (prompt.IsNull())
            //    return $"命令格式：生图 + 提示词\n生图模型：DallE-3";
            //else
            //{
            //    //用ai来生成提示词：
            //    string prompt_gpt = prompt; //await Ai.GetResAsync(prompt, Ai.DellE3AgentId);                                         
            //    _ = Tokens.MinusTokensRes(bm, "使用生图模型 DallE-3");
            //    res = await GenerateImageAsync(prompt_gpt, GeneratedImageSize.W1024xH1024, GeneratedImageQuality.Standard, GeneratedImageStyle.Natural);
            //    res = res.IsNull() ? RetryMsg : res;
            //    //IsAI = res.IsNull() ? 0 : 1;

            //    //保存到数据库
            //    if (res != RetryMsg)
            //        _ = DalleImages.SaveImageAsync(qq, prompt, prompt_gpt, res);
            //}
            return res;
        }

        public static async Task<string> GenerateImageAsync(string prompt, GeneratedImageSize imageSize, GeneratedImageQuality quality, GeneratedImageStyle style)
        {
            string endpoint = "https://australia-east-derlin.openai.azure.com/";
            string key = "190629909e64471f927ab52a1c3d6e76";

            //AzureOpenAIClient client = new(new Uri(endpoint), new AzureKeyCredential(key));
            AzureOpenAIClient client = new(new Uri(endpoint), new System.ClientModel.ApiKeyCredential(key));

            try
            {
                var imageClient = client.GetImageClient("Dalle3");

                var res = await imageClient.GenerateImageAsync(prompt, new ImageGenerationOptions
                {
                    Size = imageSize,
                    Quality = quality,
                    Style = style,
                    ResponseFormat = GeneratedImageFormat.Uri
                });

                // Image Generations responses provide URLs you can use to retrieve requested images

                // 返回图片 URI
                return res.Value.ImageUri.ToString();
            }
            catch (RequestFailedException ex) when (ex.Status == 429)
            {
                Debug($"Error:{ex.Message}", "Dall-E 3 被限制调用频率，可能欠费");
                return RetryMsg;
            }
            catch (Exception ex)
            {
                Debug($"Error:{ex.Message}", "Dall-E 3");
                return RetryMsg;
            }
        }

    }
}
