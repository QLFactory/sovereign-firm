-- 002_invites.up.sql
-- Add invitations table for enterprise user management

CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'member' CHECK (role IN ('admin', 'member', 'viewer')),
    token VARCHAR(64) UNIQUE NOT NULL,
    invited_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

-- Enable RLS
ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;

-- Invitations RLS: users can only see invitations in their tenant
CREATE POLICY invitations_tenant_isolation ON invitations
    USING (tenant_id = COALESCE(NULLIF(current_setting('app.tenant_id', true), '')::UUID, tenant_id));

CREATE POLICY invitations_insert_policy ON invitations
    FOR INSERT WITH CHECK (tenant_id = COALESCE(NULLIF(current_setting('app.tenant_id', true), '')::UUID, tenant_id));

CREATE INDEX idx_invitations_tenant ON invitations(tenant_id);
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_email ON invitations(email);
