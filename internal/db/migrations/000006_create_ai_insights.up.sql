CREATE TABLE IF NOT EXISTS ai_insights (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    insight_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

-- We add an index here to make searching faster!
-- Because our "Lazy Generation" will constantly ask: 
-- "Find the latest insight for this user of this type"
CREATE INDEX IF NOT EXISTS idx_ai_insights_user_type ON ai_insights(user_id, insight_type);