using System;

namespace BotWorker.Domain.Entities
{
    public class GroupSignIn
    {
        public long WeiboId { get; set; }
        public long RobotQq { get; set; }
        public long WeiboQq { get; set; }
        public string WeiboInfo { get; set; } = string.Empty;
        public int WeiboType { get; set; }
        public long GroupId { get; set; }
        public DateTime InsertDate { get; set; }
    }
}

