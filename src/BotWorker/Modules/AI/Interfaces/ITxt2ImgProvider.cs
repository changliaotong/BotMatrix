namespace BotWorker.Modules.AI.Interfaces
{
    public interface ITxt2ImgProvider
    {
        string ProviderName { get; }
        Task<string> GenerateImageAsync(string prompt, ImageGenerationOptions options);
    }

    public class ImageGenerationOptions
    {
        public string? Size { get; set; }
        public string? Quality { get; set; }
        public string? Style { get; set; }
        public CancellationToken CancellationToken { get; set; } = default;
    }
}
