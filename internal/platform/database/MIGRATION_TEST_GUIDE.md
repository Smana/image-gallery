# Atlas Migration Testing Guide

This guide walks you through testing the Atlas-only migration workflow with a simple schema change.

## Prerequisites

- Docker and docker-compose installed
- Atlas CLI installed (`make install-tools` if needed)

**Note**: Tests now use Atlas CLI too! No more embedded schemas - Atlas everywhere for consistency.

**Important**: If you've previously run the application or tests, your database may already have the schema but no Atlas migration history. This will cause "database is not clean" errors. See Step 4 for solutions.

**Key Distinction**:
- **Existing schema**: Use `--baseline` to mark current schema as applied
- **Empty database**: Use `make migrate` to apply all migrations from scratch

## Test Scenario: Add and Remove a Test Column

### Step 1: Prepare Environment

```bash
# Clean start - remove volumes and restart PostgreSQL
docker-compose down -v postgres
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
sleep 5

# Verify PostgreSQL is running
docker-compose ps postgres

# Remove any test migration files for clean start
rm -f internal/platform/database/migrations/002_add_test_column.sql

# Regenerate migration hash with only 001 file
atlas migrate hash --env local
```

### Step 2: Simulate Existing Schema (Baseline Test)

```bash
# Apply initial schema first (simulates existing deployment with migration 001)
make migrate

# Remove Atlas tracking to simulate existing schema without Atlas history
docker-compose exec postgres psql -U testuser -d image_gallery_test -c "DROP SCHEMA IF EXISTS atlas_schema_revisions CASCADE;"

# Remove test column to simulate having only the initial schema
docker-compose exec postgres psql -U testuser -d image_gallery_test -c "ALTER TABLE images DROP COLUMN IF EXISTS test_migration_column;"

# Now Atlas sees existing schema but no migration history
atlas migrate status --env local
# Should show: Migration Status: PENDING, Next Version: 001
```

### Step 3: Create Test Migration

```bash
# Create the test migration file
cat > internal/platform/database/migrations/002_add_test_column.sql << 'EOF'
-- Add test column to images table (TEST MIGRATION - will be removed)
ALTER TABLE images ADD COLUMN test_migration_column VARCHAR(50) DEFAULT 'test_value';

-- Add comment to make it clear this is temporary
COMMENT ON COLUMN images.test_migration_column IS 'Temporary test column for migration testing';
EOF

# Update hash to include the new migration
atlas migrate hash --env local

# Verify the file was created
cat internal/platform/database/migrations/002_add_test_column.sql
```

### Step 4: Set Baseline

```bash
# Try to apply migrations - will get "database is not clean" error
make migrate

# Use baseline to mark existing schema as migration 001 already applied
atlas migrate apply --env local --baseline 001

# Check status - should show both migrations applied (001 as baseline, 002 executed)
atlas migrate status --env local
```

### Step 5: Apply Test Migration

```bash
# Migration 002 should already be applied by the baseline command
# Check final status
atlas migrate status --env local
# Should show: Already at latest version
```

### Step 6: Verify Migration Applied

```bash
# Connect to database and verify the column exists
docker-compose exec postgres psql -U testuser -d image_gallery_test -c "\d+ images;"

# Check Atlas migration history
docker-compose exec postgres psql -U testuser -d image_gallery_test -c "SELECT * FROM atlas_schema_revisions.atlas_schema_revisions ORDER BY version;"

# Inspect schema with Atlas
atlas schema inspect --env local
```

### Step 7: Test Application Behavior

```bash
# Start the application (should connect without issues)
make dev

# In another terminal, check the logs
# Should see: "Migrations are handled by Atlas - skipping application-level migrations"

# Stop the dev server (Ctrl+C)
```

### Step 8: Test Schema Validation

```bash
# Validate Atlas configuration
atlas schema validate --env local

# Generate a diff (should show no pending changes since we just applied)
atlas migrate diff --env local

# Try to generate a new migration (should show no changes)
atlas migrate diff --env local add_nothing
```

### Step 9: Create Reverse Migration (Cleanup)

```bash
# Create cleanup migration
cat > internal/platform/database/migrations/003_remove_test_column.sql << 'EOF'
-- Remove test column from images table (cleanup test migration)
ALTER TABLE images DROP COLUMN IF EXISTS test_migration_column;
EOF

# Apply the cleanup migration
make migrate

# Verify column is removed
docker-compose exec postgres psql -U testuser -d image_gallery_test -c "\d+ images;"
```

### Step 10: Test Kubernetes ConfigMap Generation

```bash
# Generate what would be the Kubernetes ConfigMap (requires kubectl)
kubectl create configmap image-gallery-migrations-test \
  --from-file=atlas.hcl \
  --from-file=migrations=internal/platform/database/migrations \
  --dry-run=client -o yaml > test-configmap.yaml

# Review the generated ConfigMap
cat test-configmap.yaml

# Clean up the test file
rm test-configmap.yaml
```

### Step 11: Test Error Scenarios

```bash
# Try to create a problematic migration to see Atlas safety
cat > internal/platform/database/migrations/004_bad_migration.sql << 'EOF'
-- This migration has a syntax error
ALTER TABLE images ADD COLUMN bad_column INVALID_TYPE;
EOF

# Try to apply it (should fail gracefully)
make migrate

# Remove the bad migration
rm internal/platform/database/migrations/004_bad_migration.sql
```

### Step 12: Clean Up Test Migrations

```bash
# Remove test migration files
rm internal/platform/database/migrations/002_add_test_column.sql
rm internal/platform/database/migrations/003_remove_test_column.sql

# Verify we're back to just the initial migration
ls -la internal/platform/database/migrations/

# Check final migration status
atlas migrate status --env local
```

### Step 13: Test Integration Tests

```bash
# Set test database URL for Atlas
export TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5432/image_gallery_test?sslmode=disable"

# Run integration tests to verify they work with Atlas
go test ./internal/platform/database -run Integration -v

# Integration tests now use Atlas CLI to set up test schema - no more embedded schemas!
# Tests use the same migration files as local development and production
```

### Step 14: Reset Database (Optional)

If you want to start completely fresh:

```bash
# Stop and remove database
docker-compose down -v postgres

# Start fresh database
docker-compose up -d postgres
sleep 5

# Apply initial migration
make migrate

# Verify clean state
atlas migrate status --env local
```

## Expected Results

### ✅ What Should Work

- Atlas CLI commands execute without errors
- Migrations apply and are tracked properly
- Application starts without trying to run migrations
- Schema changes are visible in database
- Atlas maintains migration history

### ❌ What Should Fail Gracefully

- Bad SQL syntax in migrations
- Missing Atlas CLI (should show helpful error)
- Database connection issues (clear error messages)

## Key Observations

During this test, you should notice:

1. **Application Simplicity**: The app never runs migrations - just logs that Atlas handles them
2. **Atlas Power**: Atlas tracks migration state, validates schemas, and provides safety
3. **Local/K8s Consistency**: Same migration files work in both environments
4. **Safety First**: Atlas catches problems before they reach production

## Troubleshooting

### If Atlas Commands Fail

```bash
# Check if Atlas is installed
atlas version

# Install if missing
make install-tools
```

### If Database Connection Fails

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart if needed
docker-compose restart postgres
```

### If Migration Fails

```bash
# Check Atlas logs for detailed error
atlas migrate status --env local

# Verify database connectivity
docker-compose exec postgres psql -U testuser -d image_gallery_test -c "SELECT 1;"

# If you get "database is not clean" error:
# This means the database has existing schema but no Atlas migration history
# Option 1: Use baseline to mark current schema as initial migration (recommended)
atlas migrate apply --env local --baseline 001

# Option 2: Clean the database (only for fresh testing)
docker-compose down -v postgres && docker-compose up -d postgres && sleep 5

# Option 3: Force apply if you're sure (use with caution)
atlas migrate apply --env local --allow-dirty
```

## Next Steps

After completing this test successfully:

1. **For Production**: The same migration files go to your cloud-native-ref repo
2. **For Development**: Continue using `make migrate` for all schema changes
3. **For New Migrations**: Always create new .sql files in the migrations directory

The workflow is now proven to work end-to-end! 🚀