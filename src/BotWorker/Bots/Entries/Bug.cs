using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Entries;
public class Bug : MetaData<Bug>
{
    public override string TableName => "Bug";
    public override string KeyField => "Id";

    // 增加一条Bug信息
    public static int Insert(object bugInfo, string? bugGroup = null)
    {
        return Insert([
            new Cov("BugGroup", bugGroup.AsString()),
                new Cov("BugInfo", bugInfo.AsString()),
            ]);
    }
}

