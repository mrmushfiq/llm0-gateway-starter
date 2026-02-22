-- LLM Gateway Starter - Database Schema
-- PostgreSQL 15+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- API KEYS (Gateway Access)
-- ============================================================================

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash VARCHAR(255) NOT NULL,  -- SHA-256 hash
    key_prefix VARCHAR(20) NOT NULL,  -- gw_abc... (for display)
    name VARCHAR(255) NOT NULL,
    
    -- Rate limiting
    rate_limit_per_minute INT DEFAULT 100,
    
    -- Caching
    cache_enabled BOOLEAN DEFAULT true,
    cache_ttl_seconds INT DEFAULT 3600,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMPTZ,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_active ON api_keys(is_active);

-- ============================================================================
-- MODEL PRICING (Cost Tracking)
-- ============================================================================

CREATE TABLE model_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,  -- 'openai', 'anthropic', 'google'
    model VARCHAR(255) NOT NULL,
    input_per_1k_tokens DECIMAL(10,6) NOT NULL,
    output_per_1k_tokens DECIMAL(10,6) NOT NULL,
    context_window INT,
    supports_streaming BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(provider, model)
);

CREATE INDEX idx_model_pricing_provider ON model_pricing(provider);

-- Insert default pricing
INSERT INTO model_pricing (provider, model, input_per_1k_tokens, output_per_1k_tokens, context_window, supports_streaming) VALUES
-- OpenAI
('openai', 'gpt-4', 0.03, 0.06, 8192, true),
('openai', 'gpt-4-turbo', 0.01, 0.03, 128000, true),
('openai', 'gpt-4o', 0.005, 0.015, 128000, true),
('openai', 'gpt-4o-mini', 0.00015, 0.0006, 128000, true),
('openai', 'gpt-3.5-turbo', 0.0005, 0.0015, 16385, true),

-- Anthropic
('anthropic', 'claude-opus-4-5-20251101', 0.015, 0.075, 200000, true),
('anthropic', 'claude-sonnet-4-5-20250929', 0.003, 0.015, 200000, true),
('anthropic', 'claude-haiku-4-5-20251001', 0.001, 0.005, 200000, true),
('anthropic', 'claude-3-5-haiku-20241022', 0.001, 0.005, 200000, true),

-- Google Gemini
('google', 'gemini-2.5-flash', 0.0001, 0.0004, 1048576, true),
('google', 'gemini-2.5-pro', 0.00125, 0.005, 2097152, true),
('google', 'gemini-2.0-flash', 0.0001, 0.0004, 1048576, true),
('google', 'gemini-2.0-flash-exp', 0.0001, 0.0004, 1048576, true);

-- ============================================================================
-- REQUEST LOGS (Analytics)
-- ============================================================================

CREATE TABLE gateway_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    
    -- Request details
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    
    -- Cost & performance
    cost_usd DECIMAL(10,6) DEFAULT 0,
    latency_ms INT NOT NULL,
    
    -- Token usage
    prompt_tokens INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    total_tokens INT DEFAULT 0,
    
    -- Cache & failover
    cache_hit BOOLEAN DEFAULT false,
    failover_used BOOLEAN DEFAULT false,
    original_provider VARCHAR(50),
    
    -- Status
    status_code INT NOT NULL,
    error_message TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_gateway_logs_api_key ON gateway_logs(api_key_id);
CREATE INDEX idx_gateway_logs_created ON gateway_logs(created_at DESC);
CREATE INDEX idx_gateway_logs_model ON gateway_logs(model);
CREATE INDEX idx_gateway_logs_provider ON gateway_logs(provider);

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Get daily spend for an API key
CREATE OR REPLACE FUNCTION get_daily_spend(p_api_key_id UUID)
RETURNS DECIMAL AS $$
    SELECT COALESCE(SUM(cost_usd), 0)
    FROM gateway_logs
    WHERE api_key_id = p_api_key_id
      AND created_at >= date_trunc('day', NOW());
$$ LANGUAGE SQL;

-- Get monthly spend for an API key
CREATE OR REPLACE FUNCTION get_monthly_spend(p_api_key_id UUID)
RETURNS DECIMAL AS $$
    SELECT COALESCE(SUM(cost_usd), 0)
    FROM gateway_logs
    WHERE api_key_id = p_api_key_id
      AND created_at >= date_trunc('month', NOW());
$$ LANGUAGE SQL;

-- ============================================================================
-- SAMPLE API KEY (FOR TESTING)
-- ============================================================================

-- Create a test API key: gw_test_abc123
-- Hash: SHA-256 of "gw_test_abc123"
INSERT INTO api_keys (key_hash, key_prefix, name, rate_limit_per_minute, cache_enabled, is_active)
VALUES (
    encode(digest('gw_test_abc123', 'sha256'), 'hex'),
    'gw_test_abc',
    'Test API Key',
    100,
    true,
    true
);

-- Print the test key
DO $$
BEGIN
    RAISE NOTICE 'Test API key created: gw_test_abc123';
    RAISE NOTICE 'Use this key for testing: curl -H "Authorization: Bearer gw_test_abc123" ...';
END $$;
