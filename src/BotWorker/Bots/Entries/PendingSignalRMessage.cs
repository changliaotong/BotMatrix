namespace sz84.Bots.Entries
{
    public class PendingSignalRMessage
    {
        public string MethodName { get; set; } = string.Empty;
        public object[] Arguments { get; set; } = Array.Empty<object>();
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    }
}
