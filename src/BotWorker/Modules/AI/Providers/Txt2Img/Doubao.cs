using System.Text;
using System.Text.Json;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Providers.Configs;

namespace BotWorker.Modules.AI.Providers.Txt2Img
{
    public class Doubao : MetaData<Doubao>, ITxt2ImgProvider
    {
        public override string TableName => "robot_doubao_img";
        public override string KeyField => "id";

        public string ProviderName => "Doubao Txt2Img";

        public async Task<string> GenerateImageAsync(string prompt, BotWorker.Modules.AI.Interfaces.ImageGenerationOptions options)
        {
            var apiUrl = DoubaoTxt2Img.Url;

            var requestData = new
            {
                req_key = "high_aes_general_v21_L",
                prompt = prompt,
                model_version = "general_v2.1_L",
                req_schedule_conf = "general_v20_9B_pe",
                seed = -1,
                scale = 3.5,
                ddim_steps = 25,
                width = 512,
                height = 512,
                use_pre_llm = true,
                use_sr = true,
                return_url = true,
                logo_info = new
                {
                    add_logo = false,
                    position = 0,
                    language = 0,
                    opacity = 0.3,
                    logo_text_content = ""
                }
            };

            var jsonRequest = JsonSerializer.Serialize(requestData);

            using var httpClient = new HttpClient();
            try
            {
                httpClient.DefaultRequestHeaders.Add("Authorization", $"Bearer {DoubaoTxt2Img.Secret}");

                var content = new StringContent(jsonRequest, Encoding.UTF8, "application/json");
                var response = await httpClient.PostAsync(apiUrl, content);                

                var responseContent = await response.Content.ReadAsStringAsync();

                if (response.IsSuccessStatusCode)
                {
                    var responseData = JsonSerializer.Deserialize<ResponseData>(responseContent);
                    if (responseData?.data?.image_urls != null && responseData.data.image_urls.Count > 0)
                    {
                        return responseData.data.image_urls[0];
                    }
                }
            }
            catch (Exception ex)
            {
                Debug($"An error occurred while generating image with Doubao: {ex.Message}");
            }
            return string.Empty;
        }

        public class ResponseData
        {
            public int code { get; set; }
            public Data? data { get; set; } 
            public string message { get; set; } = string.Empty;
            public string request_id { get; set; } = string.Empty;
            public int status { get; set; }
            public string time_elapsed { get; set; } = string.Empty;
        }

        public class Data
        {
            public List<string> image_urls { get; set; } = [];
        }
    }
}
