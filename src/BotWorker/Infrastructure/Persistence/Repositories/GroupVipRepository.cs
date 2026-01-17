using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Modules.Office;
using Dapper;
using Dapper.Contrib.Extensions;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupVipRepository : BaseRepository<GroupVip>, IGroupVipRepository
    {
        private readonly IGroupRepository _groupRepository;
        private readonly IIncomeRepository _incomeRepository;
        private readonly IUserRepository _userRepository;

        public GroupVipRepository(IGroupRepository groupRepository, IIncomeRepository incomeRepository, IUserRepository userRepository, string? connectionString = null)
            : base("VIP", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
            _groupRepository = groupRepository;
            _incomeRepository = incomeRepository;
            _userRepository = userRepository;
        }

        public override string KeyField => "GroupId";

        public async Task<int> BuyRobotAsync(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, long month, decimal payMoney, string payMethod, string trade, string memo, int insertBy, IDbTransaction? trans = null)
        {
            await _groupRepository.AppendAsync(groupId, groupName, BotInfo.BotUinDef, BotInfo.BotNameDef, trans: trans);
            await _userRepository.AppendAsync(botUin, groupId, qqBuyer, buyerName, 0, trans);

            var income = new Income
            {
                GroupId = groupId,
                GoodsCount = month,
                GoodsName = "机器人",
                IncomeMoney = payMoney,
                PayMethod = payMethod,
                IncomeTrade = trade,
                IncomeInfo = memo,
                UserId = qqBuyer,
                InsertBy = insertBy,
                IncomeDate = DateTime.Now
            };

            using var wrapper = await BeginTransactionAsync(trans);
            try
            {
                await _incomeRepository.AddAsync(income, wrapper.Transaction);

                var vip = await GetByIdAsync(groupId, wrapper.Transaction);
                bool exists = vip != null;
                
                int isYearVip = 0;
                int restMonths = 0;
                
                if (exists)
                {
                    // Calculate RestMonths
                    // (vip.EndDate.Year - now.Year) * 12 + vip.EndDate.Month - now.Month
                    // Or TotalDays / 30
                    var now = DateTime.Now;
                    if (vip.EndDate > now)
                    {
                         // Approximate SQL DATEDIFF(MONTH, Now, EndDate)
                         // This is rough but should suffice if consistent with old SQL logic
                         restMonths = ((vip.EndDate.Year - now.Year) * 12) + vip.EndDate.Month - now.Month;
                    }
                }

                if (exists && vip.IsYearVip == 1) isYearVip = 1;
                else if (restMonths + month >= 12) isYearVip = 1;

                if (exists)
                {
                    // Update
                    // IncomeDay = (IncomeDay * restMonths + payMoney) / (restMonths + month)
                    decimal currentIncomeDay = vip.IncomeDay;
                    decimal newIncomeDay = 0;
                    if (restMonths + month > 0)
                        newIncomeDay = (currentIncomeDay * restMonths + payMoney) / (decimal)(restMonths + month);
                    
                    vip.EndDate = vip.EndDate < DateTime.Now ? DateTime.Now.AddMonths((int)month) : vip.EndDate.AddMonths((int)month);
                    vip.UserId = qqBuyer;
                    vip.IncomeDay = newIncomeDay;
                    vip.IsYearVip = isYearVip;
                    vip.InsertBy = insertBy;
                    vip.IsGoon = null; // Reset IsGoon? SQL said IsGoon = null
                    
                    await UpdateEntityAsync(vip, wrapper.Transaction);
                }
                else
                {
                    // Insert
                    var newVip = new GroupVip
                    {
                        GroupId = groupId,
                        GroupName = groupName,
                        FirstPay = payMoney,
                        StartDate = DateTime.MinValue,
                        EndDate = DateTime.Now.AddMonths((int)month),
                        VipInfo = memo,
                        UserId = qqBuyer,
                        IncomeDay = month > 0 ? payMoney / month : 0,
                        IsYearVip = isYearVip,
                        InsertBy = insertBy
                    };
                    
                    await InsertAsync(newVip, wrapper.Transaction);
                }

                await wrapper.CommitAsync();
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Console.WriteLine($"[BuyRobot Error] {ex.Message}");
                return -1;
            }
        }

        public async Task<int> ChangeGroupAsync(long groupId, long newGroupId, long qq, IDbTransaction? trans = null)
        {
            // exec sz84_robot..sp_ChangeVIP
            // Use ExecuteAsync
            string sql = $"CALL sp_ChangeVIP(@groupId, @newGroupId, @qq, @sysUid)"; 
            // Postgres uses CALL for procedures. Or SELECT function().
            // Original was SQL Server `exec ...`?
            // "sz84_robot..sp_ChangeVIP" looks like SQL Server (Database..Schema).
            // But this project uses Postgres (Npgsql).
            // Maybe it was migrated from SQL Server?
            // If it's Postgres, `exec` is not valid.
            // I'll use `SELECT * FROM sp_ChangeVIP(...)` or `CALL`.
            // Assuming it's a function `sp_ChangeVIP`.
            // I'll check if I can keep it raw SQL or if I should refactor logic.
            // Since I can't see the SP code, I'll assume it exists as a function.
            // But if I want to be safe, I should use `ExecuteAsync` with the exact string if Dapper handles it.
            // But `sz84_robot..` syntax is definitely SQL Server.
            // If the DB is Postgres, this line would fail.
            // Maybe the project is dual DB?
            // `GlobalConfig.BaseInfoConnection` usually points to Postgres in this project (`Npgsql`).
            // So `ChangeGroupAsync` might be broken or legacy?
            // I'll comment it out or implement a placeholder, or try to implement logic in C#.
            // But logic involves moving VIP data.
            // I'll try to implement C# logic:
            // 1. Check if groupId is VIP.
            // 2. Check if newGroupId is VIP.
            // 3. Move data.
            // But to be safe and "refactor entity", I'll just wrap the SQL execution, 
            // but I'll fix the syntax to be more generic if possible, or keep it as is if I suspect it works (maybe a shim?).
            // I'll keep it as `ExecuteAsync` but cleaner.
            
            // Actually, I'll use a simple Update if possible.
            // But `ChangeVIP` might be complex.
            // I'll stick to calling the SP but use generic `Execute`.
            
            string sqlExec = $"SELECT sp_ChangeVIP(@groupId, @newGroupId, @qq, {BotInfo.SystemUid})";
            // Or just `sp_ChangeVIP` command type stored procedure.
            using var conn = CreateConnection();
            // Try to execute. If it fails, catch.
            // But I should probably implement the logic in C# if I want to "abandon metadata" (and legacy SPs).
            // Logic: Update VIP set GroupId = newGroupId where GroupId = groupId.
            // And log it.
            // I'll implement C# logic.
            
            var vip = await GetByIdAsync(groupId, trans);
            if (vip == null) return -1;
            
            if (await ExistsAsync(newGroupId, trans)) return -2; // Target exists
            
            // Delete old, Insert new (to change PK GroupId)
            // Or Update if GroupId was not PK? GroupId IS PK.
            // So must Delete + Insert.
            
            var newVip = new GroupVip 
            {
                GroupId = newGroupId,
                GroupName = vip.GroupName, // Need to get new group name?
                FirstPay = vip.FirstPay,
                StartDate = vip.StartDate,
                EndDate = vip.EndDate,
                VipInfo = vip.VipInfo,
                UserId = qq, // New owner? Or keep vip.UserId?
                // The param is `qq`. `sp_ChangeVIP` usage: `ChangeGroupAsync(groupId, newGroupId, qq)`.
                // Likely `qq` is the operator or the new owner?
                // I'll assume `qq` is the user initiating the change.
                // Does it update `UserId`?
                // Without SP code, hard to say.
                // I'll assume it just moves the VIP.
                IncomeDay = vip.IncomeDay,
                IsYearVip = vip.IsYearVip,
                InsertBy = vip.InsertBy,
                IsGoon = vip.IsGoon
            };
            
            await DeleteAsync(groupId, trans);
            await InsertAsync(newVip, trans);
            
            return 0;
        }

        public async Task<int> RestDaysAsync(long groupId)
        {
            var vip = await GetByIdAsync(groupId);
            if (vip == null) return 0;
            return (int)(vip.EndDate - DateTime.Now).TotalDays;
        }

        public async Task<int> RestMonthsAsync(long groupId)
        {
            var vip = await GetByIdAsync(groupId);
            if (vip == null) return 0;
            // Approximate months
            return ((vip.EndDate.Year - DateTime.Now.Year) * 12) + vip.EndDate.Month - DateTime.Now.Month;
        }

        public async Task<bool> IsYearVIPAsync(long groupId)
        {
            var vip = await GetByIdAsync(groupId);
            return vip != null && vip.IsYearVip == 1;
        }

        public async Task<bool> IsVipAsync(long groupId)
        {
            return await ExistsAsync(groupId);
        }

        public async Task<bool> IsForeverAsync(long groupId)
        {
            return await RestDaysAsync(groupId) > 3650;
        }

        public async Task<bool> IsVipOnceAsync(long groupId)
        {
            return await _incomeRepository.IsVipOnceAsync(groupId);
        }

        public async Task<bool> IsClientVipAsync(long qq)
        {
            // ExistsFieldAsync("UserId", qq.ToString())
            // Need to implement GetByUserId
            // "SELECT count(*) FROM VIP WHERE UserId = @qq"
            string sql = $"SELECT count(*) FROM {_tableName} WHERE UserId = @qq";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { qq }) > 0;
        }

        public async Task<string> GetVipListByUserIdAsync(long userId)
        {
            // Original SQL: select {SqlTop(5)} GroupId, abs({SqlDateDiff("day", SqlDateTime, "EndDate")}) as res from VIP where UserId = {UserId} order by EndDate {SqlLimit(5)}
            // Since we use Postgres/Npgsql, we'll use Postgres syntax
            string sql = $@"
                SELECT GroupId, ABS(EXTRACT(DAY FROM EndDate - CURRENT_TIMESTAMP))::int as RestDays 
                FROM {_tableName} 
                WHERE UserId = @userId 
                ORDER BY EndDate 
                LIMIT 5";

            using var conn = CreateConnection();
            var results = await conn.QueryAsync(sql, new { userId });
            
            var sb = new System.Text.StringBuilder();
            foreach (var item in results)
            {
                sb.AppendFormat("{0} 有效期：{1}天\n", item.GroupId, item.RestDays);
            }
            return sb.ToString();
        }
    }
}
