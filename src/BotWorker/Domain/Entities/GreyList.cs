namespace BotWorker.Domain.Entities
{
    public class GreyList
    {
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public long GreyId { get; set; }
        public string GreyInfo { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        // 灰名单指令：灰、加灰、删灰、取消灰、解除灰名单…
        public const string regexGrey = @"^(?<cmdName>(取消|解除|删除)?(灰名单|灰|加灰|删灰))(?<cmdPara>([ ]*(\[?@:?)?[1-9]+\d*(\]?))+)$";
    }
}
