-- Order pipeline (ops desk). status on orders is the current step slug.

CREATE TABLE IF NOT EXISTS order_steps (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(64) UNIQUE NOT NULL,
    label VARCHAR(120) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_system BOOLEAN NOT NULL DEFAULT false,
    is_terminal BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO order_steps (slug, label, sort_order, is_system, is_terminal) VALUES
    ('created', 'Created', 10, true, false),
    ('awaiting_payment', 'Waiting for payment', 20, true, false),
    ('paid', 'Paid', 30, true, false),
    ('preparation', 'Preparation', 40, false, false),
    ('delivered', 'Delivered', 50, false, false),
    ('cancelled', 'Cancelled', 90, true, true),
    ('refunded', 'Refunded', 91, true, true),
    ('failed', 'Failed', 92, true, true)
ON CONFLICT (slug) DO NOTHING;

UPDATE orders SET status = 'awaiting_payment' WHERE status IN ('pending', 'processing', '');
UPDATE orders SET status = 'paid' WHERE status IN ('confirmed');
UPDATE orders SET status = 'delivered' WHERE status IN ('shipped', 'completed');
