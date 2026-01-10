using System;
using System.Collections.Generic;

namespace BotWorker.Domain.Models
{
    public class UserSession
    {
        public string SessionId { get; set; } = string.Empty;
        public long UserId { get; set; }
        public DateTime LastActive { get; set; } = DateTime.Now;
        public Dictionary<string, string> Data { get; set; } = new();
    }
}


