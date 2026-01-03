-- 确保豆包 Provider 存在
INSERT INTO ai_provider_gorms (id, name, type, enabled, base_url, api_key, created_at, updated_at)
VALUES (7, 'Doubao', 'openai', true, 'https://ark.cn-beijing.volces.com/api/v3', 'YOUR_API_KEY_HERE', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET 
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    enabled = EXCLUDED.enabled,
    base_url = EXCLUDED.base_url;

-- 确保豆包 Embedding 模型存在
INSERT INTO ai_model_gorms (provider_id, model_id, model_name, created_at, updated_at)
VALUES (7, 'doubao-embedding-vision-251215', 'Doubao Embedding Vision', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (model_id) DO NOTHING;
