using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Billing;
using Dapper;
using Npgsql;

using BotWorker.Infrastructure.Persistence.Repositories;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresWalletRepository : BasePostgresRepository<Wallet>, IWalletRepository
    {
        public PostgresWalletRepository(string? connectionString = null)
            : base("ai_wallets", connectionString)
        {
        }

        public async Task<Wallet?> GetByOwnerIdAsync(long ownerId)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<Wallet>(
                $"SELECT * FROM {_tableName} WHERE owner_id = @ownerId", new { ownerId });
        }

        public async Task<long> AddAsync(Wallet entity)
        {
            const string sql = @"
                INSERT INTO ai_wallets (owner_id, balance, currency, frozen_balance, total_spent, config)
                VALUES (@OwnerId, @Balance, @Currency, @FrozenBalance, @TotalSpent, @Config::jsonb)
                RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(Wallet entity)
        {
            const string sql = @"
                UPDATE ai_wallets SET 
                    balance = @Balance, frozen_balance = @FrozenBalance, 
                    total_spent = @TotalSpent, config = @Config::jsonb
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }

        public async Task<bool> UpdateBalanceAsync(long ownerId, decimal amount, bool isFreeze = false)
        {
            using var conn = CreateConnection();
            if (isFreeze)
            {
                return await conn.ExecuteAsync(
                    $"UPDATE {_tableName} SET balance = balance - @amount, frozen_balance = frozen_balance + @amount WHERE owner_id = @ownerId", 
                    new { ownerId, amount }) > 0;
            }
            else
            {
                return await conn.ExecuteAsync(
                    $"UPDATE {_tableName} SET balance = balance + @amount WHERE owner_id = @ownerId", 
                    new { ownerId, amount }) > 0;
            }
        }
    }

    public class PostgresLeaseResourceRepository : BasePostgresRepository<LeaseResource>, ILeaseResourceRepository
    {
        public PostgresLeaseResourceRepository(string? connectionString = null)
            : base("ai_lease_resources", connectionString)
        {
        }

        public async Task<IEnumerable<LeaseResource>> GetAvailableResourcesAsync(string? resourceType = null)
        {
            using var conn = CreateConnection();
            var sql = $"SELECT * FROM {_tableName} WHERE is_active = true AND status = 'available'";
            if (!string.IsNullOrEmpty(resourceType))
            {
                sql += " AND type = @resourceType";
            }
            return await conn.QueryAsync<LeaseResource>(sql, new { resourceType });
        }

        public async Task<IEnumerable<LeaseResource>> GetByProviderIdAsync(long providerId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<LeaseResource>($"SELECT * FROM {_tableName} WHERE provider_id = @providerId", new { providerId });
        }

        public async Task<long> AddAsync(LeaseResource entity)
        {
            const string sql = @"
                INSERT INTO ai_lease_resources (name, type, description, provider_id, price_per_hour, unit_name, max_capacity, current_usage, status, config, is_active)
                VALUES (@Name, @Type, @Description, @ProviderId, @PricePerHour, @UnitName, @MaxCapacity, @CurrentUsage, @Status, @Config::jsonb, @IsActive)
                RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(LeaseResource entity)
        {
            const string sql = @"
                UPDATE ai_lease_resources SET 
                    name = @Name, type = @Type, description = @Description,
                    price_per_hour = @PricePerHour, is_active = @IsActive, 
                    max_capacity = @MaxCapacity, current_usage = @CurrentUsage,
                    status = @Status, config = @Config::jsonb
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }
    }

    public class PostgresLeaseContractRepository : BasePostgresRepository<LeaseContract>, ILeaseContractRepository
    {
        public PostgresLeaseContractRepository(string? connectionString = null)
            : base("ai_lease_contracts", connectionString)
        {
        }

        public async Task<IEnumerable<LeaseContract>> GetActiveContractsAsync()
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<LeaseContract>($"SELECT * FROM {_tableName} WHERE status = 'active' AND end_time > CURRENT_TIMESTAMP");
        }

        public async Task<IEnumerable<LeaseContract>> GetByTenantIdAsync(long tenantId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<LeaseContract>($"SELECT * FROM {_tableName} WHERE tenant_id = @tenantId", new { tenantId });
        }

        public async Task<long> AddAsync(LeaseContract entity)
        {
            const string sql = @"
                INSERT INTO ai_lease_contracts (tenant_id, resource_id, start_time, end_time, status, auto_renew, total_paid, config)
                VALUES (@TenantId, @ResourceId, @StartTime, @EndTime, @Status, @AutoRenew, @TotalPaid, @Config::jsonb)
                RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(LeaseContract entity)
        {
            const string sql = @"
                UPDATE ai_lease_contracts SET 
                    status = @Status, end_time = @EndTime, auto_renew = @AutoRenew,
                    total_paid = @TotalPaid, config = @Config::jsonb
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }
    }

    public class PostgresBillingTransactionRepository : BasePostgresRepository<BillingTransaction>, IBillingTransactionRepository
    {
        public PostgresBillingTransactionRepository(string? connectionString = null)
            : base("ai_billing_transactions", connectionString)
        {
        }

        public async Task<IEnumerable<BillingTransaction>> GetByWalletIdAsync(long walletId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<BillingTransaction>($"SELECT * FROM {_tableName} WHERE wallet_id = @walletId ORDER BY id DESC", new { walletId });
        }

        public async Task<long> AddAsync(BillingTransaction entity)
        {
            const string sql = @"
                INSERT INTO ai_billing_transactions (wallet_id, type, amount, related_id, related_type, remark)
                VALUES (@WalletId, @Type, @Amount, @RelatedId, @RelatedType, @Remark)
                RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(BillingTransaction entity)
        {
            // 账单交易记录通常是不可变的
            return true;
        }
    }
}
