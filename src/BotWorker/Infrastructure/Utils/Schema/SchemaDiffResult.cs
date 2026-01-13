namespace BotWorker.Infrastructure.Utils.Schema;

public class SchemaDiffResult
{
    public string TableName { get; set; } = string.Empty;
    public List<string> MissingColumns { get; set; } = [];
    public List<string> ExtraColumns { get; set; } = [];
    public Dictionary<string, (string Expected, string Actual)> MismatchedTypes { get; set; } = [];
}
