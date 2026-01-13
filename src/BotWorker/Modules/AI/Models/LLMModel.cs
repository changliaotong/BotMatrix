using System;
using System.ComponentModel.DataAnnotations.Schema;
using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.AI.Models
{
    public enum LLMModelType
    {
        Chat,
        Image,
        Embedding,
        Audio
    }

    [Table("ai_models")]
    public class LLMModel
    {
        public long Id { get; set; }
        public long ProviderId { get; set; }
        public string Name { get; set; } = string.Empty;
        public string Type { get; set; } = "chat"; // chat, image, embedding, audio
        public int ContextWindow { get; set; } = 4096;
        public int? MaxOutputTokens { get; set; }
        public decimal InputPricePer1kTokens { get; set; } = 0;
        public decimal OutputPricePer1kTokens { get; set; } = 0;
        public string Config { get; set; } = "{}"; // JSONB
        public bool IsActive { get; set; } = true;
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

        public static (long ModelId, string? ProviderName, string? ModelName) GetModelInfo(long modelId)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<ILLMRepository>();
            var model = repo.GetModelByIdAsync(modelId).GetAwaiter().GetResult();
            if (model == null) return (0, null, null);

            var provider = repo.GetProviderByIdAsync(model.ProviderId).GetAwaiter().GetResult();
            return (model.Id, provider?.Name, model.Name);
        }
    }
}
