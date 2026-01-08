using Microsoft.Data.SqlClient;

using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities;
public class Friend : MetaData<Friend>
{
    public override string TableName => "Friend";
    public override string KeyField => "BotUin";
    public override string KeyField2 => "UserId";

    public static int Append(long botUin, long userId, string nickName = "")
    {
        return Exists(botUin, userId)
            ? 0
            : Insert([
                new Cov("BotUin", botUin),
                    new Cov("UserId", userId),
                    new Cov("UserName", nickName)
            ]);
    }

    public static (string, SqlParameter[]) SqlAddCredit(long botUin, long userId, long creditPlus)
    {
        if (Exists(botUin, userId))
            return SqlPlus("Credit", creditPlus, botUin, userId);
        else
            return SqlInsert([
                new Cov("BotUin", botUin),
                    new Cov("UserId", userId),
                    new Cov("Credit", creditPlus),
                ]);
    }

    public static long GetCredit(long botUin, long userId)
    {
        return GetLong("Credit", botUin, userId);
    }

    public static long GetSaveCredit(long botUin, long userId)
    {
        return GetLong("SaveCredit", botUin, userId);
    }


    public static (string, SqlParameter[]) SqlSaveCredit(long botUin, long userId, long creditSave)
    {
        return SqlSetValues($"Credit = Credit - ({creditSave}), SaveCredit = isnull(SaveCredit, 0) + ({creditSave})", botUin, userId);
    }
}