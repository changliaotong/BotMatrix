namespace BotWorker.Modules.AI.Providers.Configs
{
    public class OpenAIAzureConfig(string deploymentName, string endpoint, string apiKey)
    {
        public string DeploymentName { get; set; } = deploymentName;
        public string Endpoint { get; set; } = endpoint;
        public string ApiKey { get; set; } = apiKey;
    }

    public static class AzureOpenAI
    {
        public static string DeploymentName { get { return "gpt-4o-mini"; } }
        public static string Endpoint { get { return "https://east-us-derlin.openai.azure.com"; } }
        public static string ApiKey { get { return "sk-..."; } }
    }
}
