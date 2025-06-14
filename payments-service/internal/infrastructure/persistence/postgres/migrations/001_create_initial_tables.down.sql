-- Drop indexes
DROP INDEX IF EXISTS idx_outbox_messages_created_at;
DROP INDEX IF EXISTS idx_outbox_messages_event_type;
DROP INDEX IF EXISTS idx_outbox_messages_status;

DROP INDEX IF EXISTS idx_inbox_messages_created_at;
DROP INDEX IF EXISTS idx_inbox_messages_event_type;
DROP INDEX IF EXISTS idx_inbox_messages_status;
DROP INDEX IF EXISTS idx_inbox_messages_event_id;

DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_user_id;
DROP INDEX IF EXISTS idx_payments_order_id;
DROP INDEX IF EXISTS idx_accounts_user_id;

-- Drop tables
DROP TABLE IF EXISTS outbox_messages;
DROP TABLE IF EXISTS inbox_messages;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS accounts; 