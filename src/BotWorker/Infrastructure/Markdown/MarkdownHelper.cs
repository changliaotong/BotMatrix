using Markdig;

namespace BotWorker.Infrastructure.Markdown
{   
    public static class MarkdownHelper
    {
        private static readonly MarkdownPipeline _pipeline = new MarkdownPipelineBuilder()
            .UseAdvancedExtensions()
            .Build();

        public static string ToHtml(string markdown)
        {
            return Markdig.Markdown.ToHtml(markdown ?? "", _pipeline);
        }
    }
}
