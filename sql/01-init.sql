-- AKEN Reporting Service - Development Database Setup
-- This script loads the production backup for development/testing

-- Note: The actual data is loaded from 20250816_backup.sql
-- This file serves as documentation for the database initialization process

-- To initialize the database with production data:
-- docker exec -i aken-postgres psql -U wizzit_pay -d wizzit_pay < 20250816_backup.sql

-- The backup includes:
-- - All production table structures
-- - Real transaction data (anonymized)
-- - Proper constraints and indexes
-- - All necessary extensions

SELECT 'Database should be initialized with 20250816_backup.sql' as initialization_note;