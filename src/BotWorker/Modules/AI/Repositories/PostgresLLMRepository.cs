using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using Dapper;
using Npgsql;

using BotWorker.Infrastructure.Persistence.Repositories;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresLLMRepository : BasePostgresRepository<LLMProvider>, ILLMRepository
    {
        public PostgresLLMRepository(string? connectionString = null)
            : base("ai_providers", connectionString)
        {
        }

        public async Task<IEnumerable<LLMProvider>> GetActiveProvidersAsync()
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<LLMProvider>(
                $"SELECT * FROM {_tableName} WHERE is_active = true ORDER BY id ASC");
        }

        public async Task<IEnumerable<LLMProvider>> GetAllProvidersAsync()
        {
            return await GetAllAsync();
        }

        public async Task<LLMProvider?> GetProviderByIdAsync(long id)
        {
            return await GetByIdAsync(id);
        }

        public async Task<long> AddProviderAsync(LLMProvider provider)
        {
            const string sql = @"
                INSERT INTO ai_providers (name, type, endpoint, api_key, config, is_active, owner_id, is_shared) 
                VALUES (@Name, @Type, @Endpoint, @ApiKey, @Config::jsonb, @IsActive, @OwnerId, @IsShared) 
                RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, provider);
        }

        public async Task<bool> UpdateProviderAsync(LLMProvider provider)
        {
            const string sql = @"
                UPDATE ai_providers SET 
                    name = @Name, type = @Type, endpoint = @Endpoint, api_key = @ApiKey, 
                    config = @Config::jsonb, is_active = @IsActive, owner_id = @OwnerId, is_shared = @IsShared
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, provider) > 0;
        }

        public async Task<bool> DeleteProviderAsync(long id)
        {
            return await DeleteAsync(id);
        }

        // Model methods
        public async Task<IEnumerable<LLMModel>> GetActiveModelsAsync()
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<LLMModel>(
                "SELECT * FROM ai_models WHERE is_active = true AND is_paused = false ORDER BY id ASC");
        }

        public async Task<LLMModel?> GetModelByIdAsync(long id)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<LLMModel>(
                "SELECT * FROM ai_models WHERE id = @id", new { id });
        }

        public async Task<IEnumerable<LLMModel>> GetModelsByProviderIdAsync(long providerId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<LLMModel>(
                "SELECT * FROM ai_models WHERE provider_id = @providerId ORDER BY id ASC", new { providerId });
        }

        public async Task<long> AddModelAsync(LLMModel model)
        {
            const string sql = @"
                INSERT INTO ai_models (provider_id, name, api_model_id, type, context_window, max_output_tokens, input_price_per_1k_tokens, output_price_per_1k_tokens, base_url, api_key, config, is_active, is_paused) 
                VALUES (@ProviderId, @Name, @ApiModelId, @Type, @ContextWindow, @MaxOutputTokens, @InputPricePer1kTokens, @OutputPricePer1kTokens, @BaseUrl, @ApiKey, @Config::jsonb, @IsActive, @IsPaused) 
                RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, model);
        }

        public async Task<bool> UpdateModelAsync(LLMModel model)
        {
            const string sql = @"
                UPDATE ai_models SET 
                    provider_id = @ProviderId, name = @Name, api_model_id = @ApiModelId, type = @Type, 
                    context_window = @ContextWindow, max_output_tokens = @MaxOutputTokens, 
                    input_price_per_1k_tokens = @InputPricePer1kTokens, output_price_per_1k_tokens = @OutputPricePer1kTokens, 
                    base_url = @BaseUrl, api_key = @ApiKey, config = @Config::jsonb, is_active = @IsActive, is_paused = @IsPaused
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, model) > 0;
        }

        public async Task<bool> DeleteModelAsync(long id)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteAsync("DELETE FROM ai_models WHERE id = @id", new { id }) > 0;
        }

        public async Task<LLMProvider?> GetUserProviderAsync(long userId, string providerName)
        {
            const string sql = "SELECT * FROM ai_providers WHERE owner_id = @userId AND name = @providerName AND is_active = true LIMIT 1";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<LLMProvider>(sql, new { userId, providerName });
        }

        public async Task<IEnumerable<LLMProvider>> GetSharedProvidersAsync(string providerName)
        {
            const string sql = "SELECT * FROM ai_providers WHERE name = @providerName AND is_shared = true AND is_active = true";
            using var conn = CreateConnection();
            return await conn.QueryAsync<LLMProvider>(sql, new { providerName });
        }

        public async Task<IEnumerable<LLMProvider>> GetUserProvidersAsync(long userId)
        {
            const string sql = "SELECT * FROM ai_providers WHERE owner_id = @userId AND is_active = true";
            using var conn = CreateConnection();
            return await conn.QueryAsync<LLMProvider>(sql, new { userId });
        }

        public async Task<bool> SaveUserProviderAsync(LLMProvider provider)
        {
            if (provider.Id > 0)
            {
                return await UpdateProviderAsync(provider);
            }
            else
            {
                var id = await AddProviderAsync(provider);
                provider.Id = id;
                return id > 0;
            }
        }

        public async Task<LLMModel?> GetModelByNameAsync(string modelName)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<LLMModel>(
                "SELECT * FROM ai_models WHERE name = @modelName AND is_active = true AND is_paused = false", new { modelName });
        }

        public async Task<bool> UpdateUsageAsync(long providerId)
        {
            const string sql = @"
                UPDATE ai_providers 
                SET config = config || jsonb_build_object(
                    'use_count', (COALESCE((config->>'use_count')::int, 0) + 1),
                    'last_used_at', CURRENT_TIMESTAMP
                )
                WHERE id = @providerId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { providerId }) > 0;
        }

        public async Task<(long ModelId, string ProviderName, string ModelName)> GetBestAvailableModelAsync(long preferredModelId)
        {
            using var conn = CreateConnection();
            var sql = @"
                SELECT m.id as ModelId, p.name as ProviderName, m.name as ModelName 
                FROM ai_models m
                JOIN ai_providers p ON m.provider_id = p.id
                WHERE m.id = @preferredModelId AND m.is_active = true AND p.is_active = true";
            
            var result = await conn.QueryFirstOrDefaultAsync<(long ModelId, string ProviderName, string ModelName)>(sql, new { preferredModelId });
            
            if (result.ModelId == 0)
            {
                // Fallback to first available model
                sql = @"
                    SELECT m.id as ModelId, p.name as ProviderName, m.name as ModelName 
                    FROM ai_models m
                    JOIN ai_providers p ON m.provider_id = p.id
                    WHERE m.is_active = true AND p.is_active = true
                    LIMIT 1";
                result = await conn.QueryFirstOrDefaultAsync<(long ModelId, string ProviderName, string ModelName)>(sql);
            }
            
            return result;
        }
    }
}
