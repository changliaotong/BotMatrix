namespace sz84.Agents.Providers.Configs
{
    public class ModelConfig(string modelId, string modelName, string description, string size, string type, int maxTokens)
    {
        public string ModelId { get; set; } = modelId;
        public string ModelName { get; set; } = modelName;
        public string Description { get; set; } = description;
        public string Size { get; set; } = size;
        public string Type { get; set; } = type;
        public int MaxTokens { get; set; } = maxTokens;

        public static long TokensLimit => 128000;
        public static long TokensOutputLimit => 16384;
        public static long TokensTimes => 1;
        public static long TokensTimesOutput => 2;
    }

}
