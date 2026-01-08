using System.Text;
using System.Text.Json;

namespace BotWorker.Agents.Providers.Txt2Img
{
    public class Doubao :MetaData<Doubao>
    {
        public override string TableName => throw new NotImplementedException();

        public override string KeyField => throw new NotImplementedException();


        public static async Task<string> GenerateImageDoubaoAsync(string text)
        {
            var apiUrl = "https://visual.volcengineapi.com?Action=CVProcess&Version=2022-08-31";

            // 创建请求数据
            var requestData = new
            {
                req_key = "high_aes_general_v21_L",
                prompt = text,
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
                    logo_text_content = "这里是明水印内容"
                }
            };

            // 转换请求数据为 JSON
            var jsonRequest = JsonSerializer.Serialize(requestData);

            using var httpClient = new HttpClient();
            try
            {
                // 设置请求头                
                httpClient.DefaultRequestHeaders.Add("Authorization", $"Bearer {Configs.Doubao.Secret}");

                // 发送 POST 请求
                var content = new StringContent(jsonRequest, Encoding.UTF8, "application/json");
                var response = await httpClient.PostAsync(apiUrl, content);                

                // 读取响应数据
                var responseContent = await response.Content.ReadAsStringAsync();

                if (response.IsSuccessStatusCode)
                {
                    Console.WriteLine("请求成功，响应内容如下：");
                    Console.WriteLine(responseContent);

                    // 解析返回 JSON 数据
                    var responseData = JsonSerializer.Deserialize<ResponseData>(responseContent);
                    if (responseData?.data?.image_urls != null && responseData.data.image_urls.Count > 0)
                    {
                        Console.WriteLine("生成的图片 URL：" + responseData.data.image_urls[0]);
                        return responseData.data.image_urls[0];
                    }
                }
                else
                {
                    Console.WriteLine("请求失败，状态码：" + response.StatusCode);
                    Console.WriteLine("响应内容：" + responseContent);
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine("发生错误：" + ex.Message);
            }
            return "";
        }

        // 定义返回数据结构
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
