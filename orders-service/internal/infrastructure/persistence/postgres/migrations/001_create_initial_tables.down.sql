-- Drop triggers
DROP TRIGGER IF EXISTS update_outbox_messages_updated_at ON outbox_messages;
DROP TRIGGER IF EXISTS update_inbox_messages_updated_at ON inbox_messages;
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_outbox_created_at;
DROP INDEX IF EXISTS idx_outbox_event_type;
DROP INDEX IF EXISTS idx_outbox_status;

DROP INDEX IF EXISTS idx_inbox_created_at;
DROP INDEX IF EXISTS idx_inbox_event_type;
DROP INDEX IF EXISTS idx_inbox_status;
DROP INDEX IF EXISTS idx_inbox_event_id;

DROP INDEX IF EXISTS idx_orders_payment_id;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;

-- Drop tables
DROP TABLE IF EXISTS outbox_messages;
DROP TABLE IF EXISTS inbox_messages;
DROP TABLE IF EXISTS orders; 