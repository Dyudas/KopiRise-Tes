-- +goose Up
-- SQL in this section is executed when the migration is applied.

PRAGMA foreign_keys = ON;

-- 1. Autentikasi dan Konfigurasi
CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive', 'suspended')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    assigned_at TEXT NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address TEXT,
    user_agent TEXT,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    description TEXT,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS activity_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    details TEXT,
    ip_address TEXT,
    created_at TEXT NOT NULL
);

-- 2. Menu, Kategori, Foto, Varian, Modifier, dan Promo
CREATE TABLE IF NOT EXISTS menu_categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS menu_items (
    id TEXT PRIMARY KEY,
    category_id TEXT NOT NULL REFERENCES menu_categories(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    base_price INTEGER NOT NULL CHECK(base_price >= 0),
    is_available INTEGER NOT NULL DEFAULT 1 CHECK(is_available IN (0, 1)),
    is_best_seller INTEGER NOT NULL DEFAULT 0 CHECK(is_best_seller IN (0, 1)),
    is_featured INTEGER NOT NULL DEFAULT 0 CHECK(is_featured IN (0, 1)),
    image_url TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT
);

CREATE TABLE IF NOT EXISTS menu_photos (
    id TEXT PRIMARY KEY,
    menu_item_id TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    photo_url TEXT NOT NULL,
    thumbnail_url TEXT,
    is_primary INTEGER NOT NULL DEFAULT 0 CHECK(is_primary IN (0, 1)),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS menu_variants (
    id TEXT PRIMARY KEY,
    menu_item_id TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    price_adjustment INTEGER NOT NULL DEFAULT 0,
    is_available INTEGER NOT NULL DEFAULT 1 CHECK(is_available IN (0, 1)),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS modifier_groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_required INTEGER NOT NULL DEFAULT 0 CHECK(is_required IN (0, 1)),
    min_select INTEGER NOT NULL DEFAULT 0 CHECK(min_select >= 0),
    max_select INTEGER NOT NULL DEFAULT 1 CHECK(max_select >= 1),
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS modifier_options (
    id TEXT PRIMARY KEY,
    modifier_group_id TEXT NOT NULL REFERENCES modifier_groups(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    additional_price INTEGER NOT NULL DEFAULT 0 CHECK(additional_price >= 0),
    is_available INTEGER NOT NULL DEFAULT 1 CHECK(is_available IN (0, 1)),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS menu_item_modifier_groups (
    menu_item_id TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    modifier_group_id TEXT NOT NULL REFERENCES modifier_groups(id) ON DELETE RESTRICT,
    PRIMARY KEY(menu_item_id, modifier_group_id)
);

CREATE TABLE IF NOT EXISTS promos (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    discount_type TEXT NOT NULL CHECK(discount_type IN ('percentage', 'fixed_amount')),
    discount_value INTEGER NOT NULL CHECK(discount_value > 0),
    max_discount_amount INTEGER CHECK(max_discount_amount IS NULL OR max_discount_amount >= 0),
    min_order_amount INTEGER NOT NULL DEFAULT 0 CHECK(min_order_amount >= 0),
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    usage_limit INTEGER CHECK(usage_limit IS NULL OR usage_limit >= 0),
    used_count INTEGER NOT NULL DEFAULT 0 CHECK(used_count >= 0),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- 3. Pelanggan, Meja, QR Token, dan Reservasi
CREATE TABLE IF NOT EXISTS customers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT,
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cafe_areas (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cafe_tables (
    id TEXT PRIMARY KEY,
    area_id TEXT NOT NULL REFERENCES cafe_areas(id) ON DELETE RESTRICT,
    table_number TEXT NOT NULL UNIQUE,
    capacity INTEGER NOT NULL CHECK(capacity > 0),
    status TEXT NOT NULL DEFAULT 'available' CHECK(status IN ('available', 'occupied', 'reserved', 'maintenance')),
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS table_qr_tokens (
    id TEXT PRIMARY KEY,
    table_id TEXT NOT NULL REFERENCES cafe_tables(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    created_at TEXT NOT NULL,
    revoked_at TEXT
);

CREATE TABLE IF NOT EXISTS reservations (
    id TEXT PRIMARY KEY,
    reservation_number TEXT NOT NULL UNIQUE,
    customer_name TEXT NOT NULL,
    customer_phone TEXT NOT NULL,
    customer_email TEXT,
    guest_count INTEGER NOT NULL CHECK(guest_count > 0),
    reservation_date TEXT NOT NULL,
    reservation_time TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 120 CHECK(duration_minutes > 0),
    area_id TEXT REFERENCES cafe_areas(id) ON DELETE SET NULL,
    table_id TEXT REFERENCES cafe_tables(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'confirmed', 'seated', 'completed', 'cancelled', 'no_show')),
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS reservation_status_history (
    id TEXT PRIMARY KEY,
    reservation_id TEXT NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL,
    notes TEXT,
    changed_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL
);

-- 4. Keuangan, Kas, dan Shift
CREATE TABLE IF NOT EXISTS financial_accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    account_type TEXT NOT NULL CHECK(account_type IN ('cash', 'bank', 'electronic', 'other')),
    account_number TEXT,
    current_balance INTEGER NOT NULL DEFAULT 0,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cash_shifts (
    id TEXT PRIMARY KEY,
    cashier_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    shift_number TEXT NOT NULL UNIQUE,
    opened_at TEXT NOT NULL,
    closed_at TEXT,
    initial_cash INTEGER NOT NULL CHECK(initial_cash >= 0),
    cash_sales INTEGER NOT NULL DEFAULT 0 CHECK(cash_sales >= 0),
    cash_in INTEGER NOT NULL DEFAULT 0 CHECK(cash_in >= 0),
    cash_out INTEGER NOT NULL DEFAULT 0 CHECK(cash_out >= 0),
    cash_refunds INTEGER NOT NULL DEFAULT 0 CHECK(cash_refunds >= 0),
    expected_cash INTEGER NOT NULL DEFAULT 0 CHECK(expected_cash >= 0),
    actual_cash INTEGER CHECK(actual_cash IS NULL OR actual_cash >= 0),
    discrepancy INTEGER,
    closing_notes TEXT,
    status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open', 'closed')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cash_movements (
    id TEXT PRIMARY KEY,
    cash_shift_id TEXT NOT NULL REFERENCES cash_shifts(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK(type IN ('cash_in', 'cash_out')),
    amount INTEGER NOT NULL CHECK(amount > 0),
    reason TEXT NOT NULL,
    authorized_by_user_id TEXT REFERENCES users(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL
);

-- 5. Pemesanan dan Pembayaran
CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    order_number TEXT NOT NULL UNIQUE,
    idempotency_key TEXT UNIQUE,
    order_type TEXT NOT NULL CHECK(order_type IN ('dine_in', 'takeaway', 'reservation')),
    table_id TEXT REFERENCES cafe_tables(id) ON DELETE SET NULL,
    customer_name TEXT,
    customer_phone TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'confirmed', 'preparing', 'ready', 'served', 'completed', 'cancelled')),
    payment_status TEXT NOT NULL DEFAULT 'unpaid' CHECK(payment_status IN ('unpaid', 'partially_paid', 'paid', 'refunded', 'partially_refunded')),
    subtotal INTEGER NOT NULL DEFAULT 0 CHECK(subtotal >= 0),
    discount_amount INTEGER NOT NULL DEFAULT 0 CHECK(discount_amount >= 0),
    tax_amount INTEGER NOT NULL DEFAULT 0 CHECK(tax_amount >= 0),
    service_charge_amount INTEGER NOT NULL DEFAULT 0 CHECK(service_charge_amount >= 0),
    total_amount INTEGER NOT NULL DEFAULT 0 CHECK(total_amount >= 0),
    paid_amount INTEGER NOT NULL DEFAULT 0 CHECK(paid_amount >= 0),
    promo_id TEXT REFERENCES promos(id) ON DELETE SET NULL,
    created_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    cash_shift_id TEXT REFERENCES cash_shifts(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS order_items (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id TEXT REFERENCES menu_items(id) ON DELETE SET NULL,
    menu_variant_id TEXT REFERENCES menu_variants(id) ON DELETE SET NULL,
    item_name_snapshot TEXT NOT NULL,
    variant_name_snapshot TEXT,
    unit_base_price_snapshot INTEGER NOT NULL CHECK(unit_base_price_snapshot >= 0),
    quantity INTEGER NOT NULL CHECK(quantity > 0),
    subtotal INTEGER NOT NULL CHECK(subtotal >= 0),
    notes TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'preparing', 'ready', 'served', 'cancelled')),
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS order_item_modifiers (
    id TEXT PRIMARY KEY,
    order_item_id TEXT NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    modifier_option_id TEXT REFERENCES modifier_options(id) ON DELETE SET NULL,
    group_name_snapshot TEXT NOT NULL,
    option_name_snapshot TEXT NOT NULL,
    additional_price_snapshot INTEGER NOT NULL CHECK(additional_price_snapshot >= 0)
);

CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    cash_shift_id TEXT REFERENCES cash_shifts(id) ON DELETE SET NULL,
    payment_number TEXT NOT NULL UNIQUE,
    idempotency_key TEXT UNIQUE,
    payment_method TEXT NOT NULL CHECK(payment_method IN ('cash', 'qris', 'bank_transfer', 'other')),
    amount INTEGER NOT NULL CHECK(amount > 0),
    cash_tendered INTEGER CHECK(cash_tendered IS NULL OR cash_tendered >= 0),
    change_due INTEGER CHECK(change_due IS NULL OR change_due >= 0),
    reference_number TEXT,
    processed_by_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'completed' CHECK(status IN ('completed', 'voided', 'refunded')),
    notes TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS refunds (
    id TEXT PRIMARY KEY,
    refund_number TEXT NOT NULL UNIQUE,
    payment_id TEXT NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    cash_shift_id TEXT REFERENCES cash_shifts(id) ON DELETE SET NULL,
    amount INTEGER NOT NULL CHECK(amount > 0),
    reason TEXT NOT NULL,
    refund_method TEXT NOT NULL CHECK(refund_method IN ('cash', 'qris', 'bank_transfer', 'other')),
    financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
    approved_by_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL
);

-- 6. Karyawan dan Profil
CREATE TABLE IF NOT EXISTS employee_profiles (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    employee_code TEXT NOT NULL UNIQUE,
    id_card_number TEXT,
    date_of_birth TEXT,
    joined_date TEXT NOT NULL,
    hourly_rate INTEGER NOT NULL DEFAULT 0 CHECK(hourly_rate >= 0),
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS shift_assignments (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shift_date TEXT NOT NULL,
    shift_type TEXT NOT NULL CHECK(shift_type IN ('morning', 'afternoon', 'night', 'full_day')),
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    notes TEXT,
    created_at TEXT NOT NULL
);

-- 7. Pengeluaran dan Transfer Akun
CREATE TABLE IF NOT EXISTS expense_categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS expenses (
    id TEXT PRIMARY KEY,
    expense_number TEXT NOT NULL UNIQUE,
    expense_date TEXT NOT NULL,
    category_id TEXT NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
    amount INTEGER NOT NULL CHECK(amount > 0),
    description TEXT NOT NULL,
    vendor TEXT,
    reference_number TEXT,
    status TEXT NOT NULL DEFAULT 'approved' CHECK(status IN ('pending', 'approved', 'rejected', 'voided')),
    created_by_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    approved_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS expense_attachments (
    id TEXT PRIMARY KEY,
    expense_id TEXT NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size INTEGER NOT NULL CHECK(file_size > 0),
    mime_type TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS account_transfers (
    id TEXT PRIMARY KEY,
    transfer_number TEXT NOT NULL UNIQUE,
    from_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
    to_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
    amount INTEGER NOT NULL CHECK(amount > 0),
    transfer_date TEXT NOT NULL,
    description TEXT,
    created_by_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL
);

-- 8. Persediaan, Bahan Baku, dan HPP
CREATE TABLE IF NOT EXISTS ingredients (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    unit TEXT NOT NULL,
    current_stock REAL NOT NULL DEFAULT 0 CHECK(current_stock >= 0),
    minimum_stock REAL NOT NULL DEFAULT 0 CHECK(minimum_stock >= 0),
    cost_per_unit INTEGER NOT NULL DEFAULT 0 CHECK(cost_per_unit >= 0),
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS suppliers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT,
    email TEXT,
    address TEXT,
    contact_person TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ingredient_purchases (
    id TEXT PRIMARY KEY,
    purchase_number TEXT NOT NULL UNIQUE,
    supplier_id TEXT REFERENCES suppliers(id) ON DELETE SET NULL,
    financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
    purchase_date TEXT NOT NULL,
    total_amount INTEGER NOT NULL CHECK(total_amount >= 0),
    expense_id TEXT REFERENCES expenses(id) ON DELETE SET NULL,
    created_by_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    notes TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ingredient_purchase_items (
    id TEXT PRIMARY KEY,
    purchase_id TEXT NOT NULL REFERENCES ingredient_purchases(id) ON DELETE CASCADE,
    ingredient_id TEXT NOT NULL REFERENCES ingredients(id) ON DELETE RESTRICT,
    quantity REAL NOT NULL CHECK(quantity > 0),
    unit_cost INTEGER NOT NULL CHECK(unit_cost >= 0),
    subtotal INTEGER NOT NULL CHECK(subtotal >= 0)
);

CREATE TABLE IF NOT EXISTS recipes (
    id TEXT PRIMARY KEY,
    menu_item_id TEXT NOT NULL UNIQUE REFERENCES menu_items(id) ON DELETE CASCADE,
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS recipe_items (
    id TEXT PRIMARY KEY,
    recipe_id TEXT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id TEXT NOT NULL REFERENCES ingredients(id) ON DELETE RESTRICT,
    quantity REAL NOT NULL CHECK(quantity > 0),
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS inventory_movements (
    id TEXT PRIMARY KEY,
    ingredient_id TEXT NOT NULL REFERENCES ingredients(id) ON DELETE RESTRICT,
    movement_type TEXT NOT NULL CHECK(movement_type IN ('purchase', 'sale_consumption', 'waste_loss', 'adjustment')),
    quantity REAL NOT NULL,
    previous_stock REAL NOT NULL,
    new_stock REAL NOT NULL,
    reference_type TEXT,
    reference_id TEXT,
    reason TEXT,
    performed_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL
);

-- ==========================================
-- INDEXES UNTUK PERFORMA QUERY PRODUKSI
-- ==========================================
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);

CREATE INDEX IF NOT EXISTS idx_menu_items_category_id ON menu_items(category_id);
CREATE INDEX IF NOT EXISTS idx_menu_items_slug ON menu_items(slug);
CREATE INDEX IF NOT EXISTS idx_menu_items_is_available ON menu_items(is_available);
CREATE INDEX IF NOT EXISTS idx_menu_variants_menu_item_id ON menu_variants(menu_item_id);
CREATE INDEX IF NOT EXISTS idx_modifier_options_group_id ON modifier_options(modifier_group_id);

CREATE INDEX IF NOT EXISTS idx_cafe_tables_area_id ON cafe_tables(area_id);
CREATE INDEX IF NOT EXISTS idx_cafe_tables_status ON cafe_tables(status);
CREATE INDEX IF NOT EXISTS idx_table_qr_tokens_token ON table_qr_tokens(token);
CREATE INDEX IF NOT EXISTS idx_table_qr_tokens_table_id ON table_qr_tokens(table_id);

CREATE INDEX IF NOT EXISTS idx_reservations_date_time ON reservations(reservation_date, reservation_time);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_customer_phone ON reservations(customer_phone);

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_payment_status ON orders(payment_status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_table_id ON orders(table_id);
CREATE INDEX IF NOT EXISTS idx_orders_cash_shift_id ON orders(cash_shift_id);
CREATE INDEX IF NOT EXISTS idx_orders_idempotency ON orders(idempotency_key);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_status ON order_items(status);
CREATE INDEX IF NOT EXISTS idx_order_item_modifiers_item_id ON order_item_modifiers(order_item_id);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_cash_shift_id ON payments(cash_shift_id);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
CREATE INDEX IF NOT EXISTS idx_payments_payment_method ON payments(payment_method);
CREATE INDEX IF NOT EXISTS idx_payments_idempotency ON payments(idempotency_key);

CREATE INDEX IF NOT EXISTS idx_refunds_payment_id ON refunds(payment_id);
CREATE INDEX IF NOT EXISTS idx_refunds_order_id ON refunds(order_id);

CREATE INDEX IF NOT EXISTS idx_cash_shifts_cashier_id ON cash_shifts(cashier_user_id);
CREATE INDEX IF NOT EXISTS idx_cash_shifts_status ON cash_shifts(status);
CREATE INDEX IF NOT EXISTS idx_cash_movements_shift_id ON cash_movements(cash_shift_id);

CREATE INDEX IF NOT EXISTS idx_expenses_category_id ON expenses(category_id);
CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(expense_date);
CREATE INDEX IF NOT EXISTS idx_expenses_account_id ON expenses(financial_account_id);

CREATE INDEX IF NOT EXISTS idx_inventory_movements_ingredient_id ON inventory_movements(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_inventory_movements_created_at ON inventory_movements(created_at);
CREATE INDEX IF NOT EXISTS idx_recipe_items_recipe_id ON recipe_items(recipe_id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS recipe_items;
DROP TABLE IF EXISTS recipes;
DROP TABLE IF EXISTS ingredient_purchase_items;
DROP TABLE IF EXISTS ingredient_purchases;
DROP TABLE IF EXISTS suppliers;
DROP TABLE IF EXISTS ingredients;
DROP TABLE IF EXISTS account_transfers;
DROP TABLE IF EXISTS expense_attachments;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS expense_categories;
DROP TABLE IF EXISTS shift_assignments;
DROP TABLE IF EXISTS employee_profiles;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS order_item_modifiers;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS cash_movements;
DROP TABLE IF EXISTS cash_shifts;
DROP TABLE IF EXISTS financial_accounts;
DROP TABLE IF EXISTS reservation_status_history;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS table_qr_tokens;
DROP TABLE IF EXISTS cafe_tables;
DROP TABLE IF EXISTS cafe_areas;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS promos;
DROP TABLE IF EXISTS menu_item_modifier_groups;
DROP TABLE IF EXISTS modifier_options;
DROP TABLE IF EXISTS modifier_groups;
DROP TABLE IF EXISTS menu_variants;
DROP TABLE IF EXISTS menu_photos;
DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS menu_categories;
DROP TABLE IF EXISTS activity_logs;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
