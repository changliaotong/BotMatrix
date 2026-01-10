using Newtonsoft.Json;
using System.Text;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.Database
{
    //远程执行sql
    public static class Remote
    {    


    }

    public class SqlRequest
    {
        public string Sql { get; set; } = string.Empty;

        public List<DbParameterDTO> Parameters { get; set; } = [];

        public bool IsDebug { get; set; } = true;
    }

    public class ExecTransRequest
    {
        public List<(string, List<DbParameterDTO>)>? Sqls { get; set; }

        public bool IsDebug { get; set; } = true;
    }
}

