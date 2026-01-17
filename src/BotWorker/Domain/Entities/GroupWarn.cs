using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("warn")]
    public class GroupWarn
    {
        [Key]
        public long Id { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public string WarnInfo { get; set; } = string.Empty;
        public long InsertBy { get; set; }
        public DateTime InsertDate { get; set; }

        public const string RegexCmdWarn = @"^[#＃﹟]?(撤回|扣分|警告|禁言|踢出|拉黑|加黑)词 *([＋－+-]*) *([\s\S]*)$";
        public const string regexParaKeyword = @"(?<keyword>[^ ]+[\s\S]*?[ $]*)";
        public const string regexQqImage = @"\[Image[\d\w {}-]*(.(jpg|png))*]";


    }
}
