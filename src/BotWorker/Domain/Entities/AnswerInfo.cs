using System;
using System.ComponentModel.DataAnnotations.Schema;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;
using Newtonsoft.Json;

namespace BotWorker.Domain.Entities
{
    //问答系统
    [Dapper.Contrib.Extensions.Table("Answer")]
    public partial class AnswerInfo
    {
        [ExplicitKey]
        public long Id { get; set; }
        public long QuestionId { get; set; }
        public string Question { get; set; } = string.Empty;
        public string Answer { get; set; } = string.Empty;
        public string AnswerBak { get; set; } = string.Empty;
        
        [Write(false)]
        public DateTime InsertDate { get; set; }
        
        public long UserId { get; set; }        
        public long GroupId { get; set; }
        public long RobotId { get; set; }
        public long BotUin { get; set; }
        public int IsOnly { get; set; }
        public int Audit { get; set; }
        public long AuditBy { get; set; }
        public DateTime AuditDate { get; set; }
        public int Audit2 { get; set; }
        public long Audit2By { get; set; }
        public DateTime Audit2Date { get; set; }        
        public string Audit2Info { get; set; } = string.Empty;
        public DateTime UpdateDate { get; set; }
        
        [Write(false)]
        public int UsedTimes { get; set; }
        
        [Write(false)]
        public int GoonTimes { get; set; }
        
        [Write(false)]
        public int UsedTimesGroup { get; set; }
        
        [Write(false)]
        public int GoonTimesGroup { get; set; }
        
        [Write(false)]
        public int Credit { get; set; }
        
        public static long Rid => BotInfo.DefaultRobotId;     
    }
}
