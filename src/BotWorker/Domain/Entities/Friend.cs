using System.Data;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities;
public class Friend : MetaData<Friend>
{
    public override string TableName => "Friend";
    public override string KeyField => "BotUin";
    public override string KeyField2 => "UserId";

    public static async Task<int> AppendAsync(long botUin, long userId, string nickName = "")
    {
        return await ExistsAsync(botUin, userId)
            ? 0
            : await InsertAsync(new
            {
                BotUin = botUin,
                UserId = userId,
                UserName = nickName
            });
    }

    public static int Append(long botUin, long userId, string nickName = "")
    {
        return Exists(botUin, userId)
            ? 0
            : Insert(new
            {
                BotUin = botUin,
                UserId = userId,
                UserName = nickName
            });
    }

    public static (string, IDataParameter[]) SqlAddCredit(long botUin, long userId, long creditPlus)
    {
        if (Exists(botUin, userId))
            return SqlPlus("Credit", creditPlus, botUin, userId);
        else
            return SqlInsert(new
            {
                BotUin = botUin,
                UserId = userId,
                Credit = creditPlus,
            });
    }

    public static async Task<long> GetCreditAsync(long botUin, long userId, IDbTransaction? trans = null)
    {
        return await GetLongAsync("Credit", botUin, userId, trans);
    }

    public static long GetCredit(long botUin, long userId)
    {
        return GetCreditAsync(botUin, userId).GetAwaiter().GetResult();
    }

    public static async Task<long> GetSaveCreditAsync(long botUin, long userId, IDbTransaction? trans = null)
    {
        return await GetLongAsync("SaveCredit", botUin, userId, trans);
    }

    public static long GetSaveCredit(long botUin, long userId)
    {
        return GetSaveCreditAsync(botUin, userId).GetAwaiter().GetResult();
    }


    public static (string, IDataParameter[]) SqlSaveCredit(long botUin, long userId, long creditSave)
    {
        return SqlSetValues($"Credit = Credit - ({creditSave}), SaveCredit = {SqlIsNull("SaveCredit", "0")} + ({creditSave})", botUin, userId);
    }
}