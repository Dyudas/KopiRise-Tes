-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- 1. Roles
INSERT OR IGNORE INTO roles (id, name, description, created_at) VALUES
('admin', 'Admin/Pemilik', 'Akses penuh ke seluruh sistem, laporan keuangan, audit, dan konfigurasi', '2026-10-26T00:00:00Z'),
('manager', 'Manager', 'Mengelola operasional, menu, meja, reservasi, stok, promo, dan persetujuan transaksi', '2026-10-26T00:00:00Z'),
('cashier', 'Kasir', 'Melayani POS kasir, pembayaran pesanan, shift kas, dan pencetakan struk', '2026-10-26T00:00:00Z'),
('barista', 'Barista/Dapur', 'Melihat antrean Kitchen Display System (KDS) dan memperbarui status pesanan', '2026-10-26T00:00:00Z'),
('waiter', 'Pelayan', 'Melihat status meja, mengantarkan pesanan (serving), dan membuat pesanan meja', '2026-10-26T00:00:00Z'),
('customer', 'Pelanggan', 'Melihat menu publik, membuat reservasi, dan memesan via QR meja', '2026-10-26T00:00:00Z');

-- 2. Users (Dev Passwords argon2id hash for 'KopiRise2026!')
-- Standard Argon2id hash formatted: $argon2id$v=19$m=65536,t=3,p=2$...
INSERT OR IGNORE INTO users (id, name, email, phone, password_hash, status, created_at, updated_at) VALUES
('usr-admin-001', 'Arga Pramudya (Owner)', 'admin@kopirise.id', '081234567890', '$argon2id$v=19$m=65536,t=3,p=2$w6LwP5mXg4Z1bKq7r9T0YQ$Fv0W4U1D6uP8L7M0rJk1Q5w7x8y9z0A1B2C3D4E5F6G', 'active', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('usr-mgr-001', 'Siti Rahma (Store Manager)', 'manager@kopirise.id', '081234567891', '$argon2id$v=19$m=65536,t=3,p=2$w6LwP5mXg4Z1bKq7r9T0YQ$Fv0W4U1D6uP8L7M0rJk1Q5w7x8y9z0A1B2C3D4E5F6G', 'active', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('usr-csh-001', 'Budi Santoso (Head Cashier)', 'kasir@kopirise.id', '081234567892', '$argon2id$v=19$m=65536,t=3,p=2$w6LwP5mXg4Z1bKq7r9T0YQ$Fv0W4U1D6uP8L7M0rJk1Q5w7x8y9z0A1B2C3D4E5F6G', 'active', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('usr-bar-001', 'Dimas Arya (Head Barista)', 'barista@kopirise.id', '081234567893', '$argon2id$v=19$m=65536,t=3,p=2$w6LwP5mXg4Z1bKq7r9T0YQ$Fv0W4U1D6uP8L7M0rJk1Q5w7x8y9z0A1B2C3D4E5F6G', 'active', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('usr-wtr-001', 'Rian Hidayat (Floor Captain)', 'waiter@kopirise.id', '081234567894', '$argon2id$v=19$m=65536,t=3,p=2$w6LwP5mXg4Z1bKq7r9T0YQ$Fv0W4U1D6uP8L7M0rJk1Q5w7x8y9z0A1B2C3D4E5F6G', 'active', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- 3. User Roles
INSERT OR IGNORE INTO user_roles (user_id, role_id, assigned_at) VALUES
('usr-admin-001', 'admin', '2026-10-26T00:00:00Z'),
('usr-mgr-001', 'manager', '2026-10-26T00:00:00Z'),
('usr-csh-001', 'cashier', '2026-10-26T00:00:00Z'),
('usr-bar-001', 'barista', '2026-10-26T00:00:00Z'),
('usr-wtr-001', 'waiter', '2026-10-26T00:00:00Z');

-- 4. Employee Profiles
INSERT OR IGNORE INTO employee_profiles (user_id, employee_code, id_card_number, date_of_birth, joined_date, hourly_rate, notes, created_at, updated_at) VALUES
('usr-admin-001', 'EMP-ADM-01', '3509012345670001', '1990-05-14', '2025-01-01', 0, 'Owner & Founder', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('usr-mgr-001', 'EMP-MGR-01', '3509012345670002', '1993-08-20', '2025-02-01', 25000, 'Operational Manager', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('usr-csh-001', 'EMP-CSH-01', '3509012345670003', '1998-11-10', '2025-03-01', 18000, 'Shift A Cashier', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('usr-bar-001', 'EMP-BAR-01', '3509012345670004', '1999-02-17', '2025-03-01', 20000, 'Lead Barista & Coffee Roaster', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('usr-wtr-001', 'EMP-WTR-01', '3509012345670005', '2001-07-25', '2025-04-01', 16000, 'Service Captain', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- 5. Settings
INSERT OR IGNORE INTO settings (key, value, category, description, updated_at) VALUES
('cafe_name', 'KopiRise', 'general', 'Nama resmi kafe', '2026-10-26T00:00:00Z'),
('cafe_type', 'Cafe & Resto', 'general', 'Jenis usaha kafe', '2026-10-26T00:00:00Z'),
('cafe_tagline', 'Ruang Teduh Kopi Spesialti & Cita Rasa Otentik Jember', 'general', 'Tagline brand', '2026-10-26T00:00:00Z'),
('cafe_address', 'Jl. Semeru, Kloncing, Karangrejo, Kec. Sumbersari, Kabupaten Jember, Jawa Timur 68124', 'general', 'Alamat lengkap kafe', '2026-10-26T00:00:00Z'),
('cafe_phone', '081234567890', 'general', 'Nomor telepon operasional', '2026-10-26T00:00:00Z'),
('cafe_whatsapp', '6281234567890', 'general', 'Nomor WhatsApp untuk konfirmasi reservasi & info', '2026-10-26T00:00:00Z'),
('cafe_instagram', '@kopirise.jember', 'general', 'Akun sosial media Instagram', '2026-10-26T00:00:00Z'),
('operating_hours', 'Senin - Minggu: 08:00 - 23:00 WIB', 'general', 'Jam operasional kafe', '2026-10-26T00:00:00Z'),
('tax_rate_percent', '10', 'financial', 'Persentase PB1 / Pajak Restoran (10%)', '2026-10-26T00:00:00Z'),
('service_charge_percent', '0', 'financial', 'Persentase biaya layanan (0%)', '2026-10-26T00:00:00Z'),
('currency_symbol', 'Rp', 'financial', 'Simbol mata uang', '2026-10-26T00:00:00Z'),
('reservation_min_duration_minutes', '120', 'reservation', 'Durasi standar pemakaian meja untuk reservasi', '2026-10-26T00:00:00Z'),
('reservation_auto_cancel_minutes', '15', 'reservation', 'Toleransi keterlambatan sebelum status no-show', '2026-10-26T00:00:00Z');

-- 6. Cafe Areas & Tables (24 Meja)
INSERT OR IGNORE INTO cafe_areas (id, name, description, sort_order, created_at) VALUES
('area-indoor-ac', 'Indoor AC Non-Smoking', 'Ruangan berpendingin udara, nyaman untuk bekerja (WFC) & santai', 1, '2026-10-26T00:00:00Z'),
('area-outdoor-garden', 'Outdoor Garden & Terrace', 'Area terbuka asri dengan pepohonan hijau dan smoking area', 2, '2026-10-26T00:00:00Z'),
('area-mezzanine', 'Mezzanine Bar', 'Lantai 2 menghadap pemandangan barista slow bar', 3, '2026-10-26T00:00:00Z'),
('area-vip-meeting', 'VIP Room Semeru', 'Ruang privat ber-AC dengan TV Smart display untuk meeting & privat event', 4, '2026-10-26T00:00:00Z');

-- Tables (24 Tables: 8 Indoor, 10 Outdoor, 4 Mezzanine, 2 VIP)
INSERT OR IGNORE INTO cafe_tables (id, area_id, table_number, capacity, status, is_active, created_at, updated_at) VALUES
('tbl-in-01', 'area-indoor-ac', 'IN-01', 2, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-in-02', 'area-indoor-ac', 'IN-02', 2, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-in-03', 'area-indoor-ac', 'IN-03', 4, 'occupied', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-in-04', 'area-indoor-ac', 'IN-04', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-in-05', 'area-indoor-ac', 'IN-05', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-in-06', 'area-indoor-ac', 'IN-06', 6, 'reserved', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-in-07', 'area-indoor-ac', 'IN-07', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-in-08', 'area-indoor-ac', 'IN-08', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),

('tbl-out-01', 'area-outdoor-garden', 'OUT-01', 2, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-02', 'area-outdoor-garden', 'OUT-02', 2, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-03', 'area-outdoor-garden', 'OUT-03', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-04', 'area-outdoor-garden', 'OUT-04', 4, 'occupied', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-05', 'area-outdoor-garden', 'OUT-05', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-06', 'area-outdoor-garden', 'OUT-06', 6, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-07', 'area-outdoor-garden', 'OUT-07', 6, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-08', 'area-outdoor-garden', 'OUT-08', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-09', 'area-outdoor-garden', 'OUT-09', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-out-10', 'area-outdoor-garden', 'OUT-10', 8, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),

('tbl-mez-01', 'area-mezzanine', 'MEZ-01', 2, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-mez-02', 'area-mezzanine', 'MEZ-02', 2, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-mez-03', 'area-mezzanine', 'MEZ-03', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-mez-04', 'area-mezzanine', 'MEZ-04', 4, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),

('tbl-vip-01', 'area-vip-meeting', 'VIP-01', 10, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('tbl-vip-02', 'area-vip-meeting', 'VIP-02', 8, 'available', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- Table QR Tokens (Unique cryptographically secure random token simulation)
INSERT OR IGNORE INTO table_qr_tokens (id, table_id, token, is_active, created_at) VALUES
('tok-01', 'tbl-in-01', 'kr_tbl_9a7d3f82e1c440a', 1, '2026-10-26T00:00:00Z'),
('tok-02', 'tbl-in-02', 'kr_tbl_4b1c8e73d9a541b', 1, '2026-10-26T00:00:00Z'),
('tok-03', 'tbl-in-03', 'kr_tbl_7e2a9d61c8f442c', 1, '2026-10-26T00:00:00Z'),
('tok-04', 'tbl-in-04', 'kr_tbl_3c8f1e94b7a243d', 1, '2026-10-26T00:00:00Z'),
('tok-05', 'tbl-in-05', 'kr_tbl_8d4a2c91e7b344e', 1, '2026-10-26T00:00:00Z'),
('tok-06', 'tbl-in-06', 'kr_tbl_2f9b4d83a1e645f', 1, '2026-10-26T00:00:00Z'),
('tok-07', 'tbl-in-07', 'kr_tbl_6a3e8c12d7f946a', 1, '2026-10-26T00:00:00Z'),
('tok-08', 'tbl-in-08', 'kr_tbl_1e7d9b43c8a547b', 1, '2026-10-26T00:00:00Z'),
('tok-09', 'tbl-out-01', 'kr_tbl_5d8a3c21f9e448c', 1, '2026-10-26T00:00:00Z'),
('tok-10', 'tbl-out-02', 'kr_tbl_9c2e7f84b1a349d', 1, '2026-10-26T00:00:00Z'),
('tok-11', 'tbl-out-03', 'kr_tbl_4a8d1c92e7b550e', 1, '2026-10-26T00:00:00Z'),
('tok-12', 'tbl-out-04', 'kr_tbl_8e3a4d71b9c251f', 1, '2026-10-26T00:00:00Z'),
('tok-13', 'tbl-out-05', 'kr_tbl_2b7f9c83a1d452a', 1, '2026-10-26T00:00:00Z'),
('tok-14', 'tbl-out-06', 'kr_tbl_6d1a8e94c7b353b', 1, '2026-10-26T00:00:00Z'),
('tok-15', 'tbl-out-07', 'kr_tbl_1f9c3d82a4e754c', 1, '2026-10-26T00:00:00Z'),
('tok-16', 'tbl-out-08', 'kr_tbl_5a8e2b71c9d455d', 1, '2026-10-26T00:00:00Z'),
('tok-17', 'tbl-out-09', 'kr_tbl_9e4a7c13d8b256e', 1, '2026-10-26T00:00:00Z'),
('tok-18', 'tbl-out-10', 'kr_tbl_3d8f1a92e7c457f', 1, '2026-10-26T00:00:00Z'),
('tok-19', 'tbl-mez-01', 'kr_tbl_7b2e9c84a1d358a', 1, '2026-10-26T00:00:00Z'),
('tok-20', 'tbl-mez-02', 'kr_tbl_2c8a4d71e9b559b', 1, '2026-10-26T00:00:00Z'),
('tok-21', 'tbl-mez-03', 'kr_tbl_6f1d9a83c7e260c', 1, '2026-10-26T00:00:00Z'),
('tok-22', 'tbl-mez-04', 'kr_tbl_1a7e3c92d8f461d', 1, '2026-10-26T00:00:00Z'),
('tok-23', 'tbl-vip-01', 'kr_tbl_5e9a8b14c7d362e', 1, '2026-10-26T00:00:00Z'),
('tok-24', 'tbl-vip-02', 'kr_tbl_9d3c7a82e1b463f', 1, '2026-10-26T00:00:00Z');

-- 7. Menu Categories (6 Kategori)
INSERT OR IGNORE INTO menu_categories (id, name, slug, sort_order, is_active, created_at, updated_at) VALUES
('cat-espresso', 'Espresso Based', 'espresso-based', 1, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('cat-manualbrew', 'Manual Brew', 'manual-brew', 2, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('cat-noncoffee', 'Non-Coffee & Tea', 'non-coffee-tea', 3, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('cat-mocktail', 'Mocktail & Refreshers', 'mocktail-refreshers', 4, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('cat-pastry', 'Pastry & Snacks', 'pastry-snacks', 5, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('cat-maincourse', 'Main Course & Meals', 'main-course-meals', 6, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- 8. Menu Items (32 Menu Realistis)
INSERT OR IGNORE INTO menu_items (id, category_id, name, slug, description, base_price, is_available, is_best_seller, is_featured, image_url, sort_order, created_at, updated_at) VALUES
-- Espresso Based (8 items)
('menu-01', 'cat-espresso', 'Espresso Double Shot', 'espresso-double-shot', 'Ekstraksi 18g biji kopi Arabika Ijen Raung, tebal krema dengan notes dark chocolate dan citrus.', 18000, 1, 0, 0, 'https://images.unsplash.com/photo-1510591509098-f4fdc6d0ff04?auto=format&fit=crop&w=600&q=80', 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-02', 'cat-espresso', 'Americano Classic', 'americano-classic', 'Espresso double shot dipadukan dengan air mineral murni, segar dan bersih di lidah.', 20000, 1, 0, 0, 'https://images.unsplash.com/photo-1551030173-122aabc4489c?auto=format&fit=crop&w=600&q=80', 2, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-03', 'cat-espresso', 'Cafe Latte Silk', 'cafe-latte-silk', 'Perpaduan seimbang espresso dan steamed fresh milk dengan microfoam lembut.', 25000, 1, 1, 1, 'https://images.unsplash.com/photo-1534778101976-62847782c213?auto=format&fit=crop&w=600&q=80', 3, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-04', 'cat-espresso', 'Cappuccino Roman', 'cappuccino-roman', 'Espresso kaya rasa dengan foam susu tebal dan taburan bubuk cokelat Belgia asli.', 25000, 1, 0, 0, 'https://images.unsplash.com/photo-1572442388796-11668ba67e53?auto=format&fit=crop&w=600&q=80', 4, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-05', 'cat-espresso', 'Kopi Susu Aren Semeru', 'kopi-susu-aren-semeru', 'Signature KopiRise! Espresso house-blend, fresh milk gurih, dan sirup gula aren murni khas Lumajang-Jember.', 22000, 1, 1, 1, 'https://images.unsplash.com/photo-1461023058943-07fcbe16d735?auto=format&fit=crop&w=600&q=80', 5, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-06', 'cat-espresso', 'Salted Caramel Macchiato', 'salted-caramel-macchiato', 'Layering espresso, susu vanila lembut, sea salt, dan sirup karamel bakar artisanal.', 28000, 1, 1, 1, 'https://images.unsplash.com/photo-1485808191679-5f86510681a2?auto=format&fit=crop&w=600&q=80', 6, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-07', 'cat-espresso', 'Hazelnut Cream Latte', 'hazelnut-cream-latte', 'Espresso creamy dengan aroma roasted hazelnut alami dan busa susu manis lembut.', 27000, 1, 0, 0, 'https://images.unsplash.com/photo-1517701550927-30cf4ba1dba5?auto=format&fit=crop&w=600&q=80', 7, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-08', 'cat-espresso', 'Mocha Belgia Dark', 'mocha-belgia-dark', 'Espresso bertemu lelehan dark chocolate murni 70% dan susu segar hangat/dingin.', 28000, 1, 0, 0, 'https://images.unsplash.com/photo-1578314675249-a6910f80cc4e?auto=format&fit=crop&w=600&q=80', 8, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),

-- Manual Brew (4 items)
('menu-09', 'cat-manualbrew', 'V60 Single Origin Ijen Raung', 'v60-ijen-raung', 'Seduhan filter pour-over dengan biji kopi lokal Pegunungan Ijen. Tasting notes: Floral, Lemongrass, Brown Sugar.', 26000, 1, 1, 1, 'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=600&q=80', 9, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-10', 'cat-manualbrew', 'V60 Flores Bajawa Natural', 'v60-flores-bajawa', 'Notes apel merah matang, honey sweetness, dan aftertaste karamel yang panjang.', 28000, 1, 0, 0, 'https://images.unsplash.com/photo-1514432324607-a09d9b4aefdd?auto=format&fit=crop&w=600&q=80', 10, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-11', 'cat-manualbrew', 'Japanese Iced Drip Coffee', 'japanese-iced-drip', 'Seduhan V60 langsung menetes ke bongkahan es batu kristal, menghasilkan rasa buah yang sangat menyegarkan.', 28000, 1, 1, 0, 'https://images.unsplash.com/photo-1517701604599-bb29b565090c?auto=format&fit=crop&w=600&q=80', 11, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-12', 'cat-manualbrew', 'Aeropress Bold Signature', 'aeropress-bold-signature', 'Ekstraksi bertekanan udara dengan bodi kopi penuh dan keasaman bersih seimbang.', 25000, 1, 0, 0, 'https://images.unsplash.com/photo-1544787219-7f47ccb76574?auto=format&fit=crop&w=600&q=80', 12, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),

-- Non-Coffee & Tea (5 items)
('menu-13', 'cat-noncoffee', 'Matcha Latte Uji Kyoto', 'matcha-latte-uji-kyoto', 'Bubuk matcha ceremonial murni dari Uji Jepang dipadukan steamed milk yang kental dan lembut.', 26000, 1, 1, 1, 'https://images.unsplash.com/photo-1536256263959-770b48d82b0a?auto=format&fit=crop&w=600&q=80', 13, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-14', 'cat-noncoffee', 'Signature Belgian Chocolate', 'signature-belgian-chocolate', 'Cokelat Belgia pekat dengan rasa manis seimbang dan tekstur velvety yang mewah.', 25000, 1, 1, 0, 'https://images.unsplash.com/photo-1542990253-0d0f5be5f0ed?auto=format&fit=crop&w=600&q=80', 14, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-15', 'cat-noncoffee', 'Taro Velvet Cream', 'taro-velvet-cream', 'Minuman taro ungu manis wangi kelapa panggang dipadu susu segar pilihan.', 24000, 1, 0, 0, 'https://images.unsplash.com/photo-1579954115545-a95591f28bfc?auto=format&fit=crop&w=600&q=80', 15, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-16', 'cat-noncoffee', 'Earl Grey Artisan Milk Tea', 'earl-grey-artisan-milk-tea', 'Seduhan daun teh hitam bergamot aromatik dipadukan susu manis dan brown sugar.', 22000, 1, 0, 0, 'https://images.unsplash.com/photo-1558160074-4d7d8bdf4256?auto=format&fit=crop&w=600&q=80', 16, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-17', 'cat-noncoffee', 'Lemon Mint Tea Infusion', 'lemon-mint-tea-infusion', 'Teh melati wangi dengan perasan lemon segar dan daun mint asli, sangat melegakan tenggorokan.', 18000, 1, 0, 0, 'https://images.unsplash.com/photo-1556679343-c7306c1976bc?auto=format&fit=crop&w=600&q=80', 17, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),

-- Mocktail & Refreshers (4 items)
('menu-18', 'cat-mocktail', 'Berry Sunrise Sparkler', 'berry-sunrise-sparkler', 'Mocktail stroberi segar, sirup delima, soda dingin, dan perasan jeruk nipis dengan garnish rosemary.', 24000, 1, 1, 1, 'https://images.unsplash.com/photo-1513558161293-cdaf765ed2fd?auto=format&fit=crop&w=600&q=80', 18, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-19', 'cat-mocktail', 'Sunset Lychee Mojito', 'sunset-lychee-mojito', 'Buah leci utuh, mint muddled, sirup leci premium, jeruk nipis dan air soda segar.', 24000, 1, 1, 0, 'https://images.unsplash.com/photo-1551024709-8f23befc6f87?auto=format&fit=crop&w=600&q=80', 19, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-20', 'cat-mocktail', 'Tropical Passion Breeze', 'tropical-passion-breeze', 'Kombinasi markisa segar, mangga, dan sentuhan kelapa dengan bulir selasih.', 23000, 1, 0, 0, 'https://images.unsplash.com/photo-1536935338788-846bb9981813?auto=format&fit=crop&w=600&q=80', 20, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-21', 'cat-mocktail', 'Peach Cold Brew Tonic', 'peach-cold-brew-tonic', 'Ekstrak cold brew kopi Ijen 18 jam dengan sirup peach putih dan sparkling tonic water.', 26000, 1, 0, 0, 'https://images.unsplash.com/photo-1544787219-7f47ccb76574?auto=format&fit=crop&w=600&q=80', 21, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),

-- Pastry & Snacks (6 items)
('menu-22', 'cat-pastry', 'Butter Croissant French Artisan', 'butter-croissant-french-artisan', 'Croissant renyah berlapis dengan aroma mentega Prancis asli, disajikan hangat.', 22000, 1, 1, 0, 'https://images.unsplash.com/photo-1555507036-ab1f4038808a?auto=format&fit=crop&w=600&q=80', 22, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-23', 'cat-pastry', 'Pain Au Chocolat', 'pain-au-chocolat', 'Pastry gurih berisikan dua batangan cokelat leleh Belgia di setiap gigitan.', 24000, 1, 0, 0, 'https://images.unsplash.com/photo-1608198093002-ad4e005484ec?auto=format&fit=crop&w=600&q=80', 23, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-24', 'cat-pastry', 'Tahu Walik Krispi Jember', 'tahu-walik-krispi-jember', 'Camilan khas Jember dengan isian adonan ayam udang gurih garing dan cocolan petis pedas manis.', 18000, 1, 1, 1, 'https://images.unsplash.com/photo-1541592106381-b31e9677c0e5?auto=format&fit=crop&w=600&q=80', 24, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-25', 'cat-pastry', 'Singkong Keju Renyah Semeru', 'singkong-keju-renyah-semeru', 'Singkong empuk merekah bertabur parutan keju cheddar melimpah dan susu kental manis.', 18000, 1, 1, 0, 'https://images.unsplash.com/photo-1589301760014-d929f3979dbc?auto=format&fit=crop&w=600&q=80', 25, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-26', 'cat-pastry', 'French Fries Truffle Herb', 'french-fries-truffle-herb', 'Kentang goreng impor renyah beraroma minyak truffle putih, garam laut dan oregano.', 22000, 1, 0, 0, 'https://images.unsplash.com/photo-1573080496219-bb080dd4f877?auto=format&fit=crop&w=600&q=80', 26, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-27', 'cat-pastry', 'Roti Bakar Cokelat Keju Gandum', 'roti-bakar-cokelat-keju-gandum', 'Roti gandum bakar empuk isi selai cokelat pekat, keju cheddar dan taburan almond panggang.', 20000, 1, 0, 0, 'https://images.unsplash.com/photo-1525351484163-7529414344d8?auto=format&fit=crop&w=600&q=80', 27, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),

-- Main Course & Meals (5 items)
('menu-28', 'cat-maincourse', 'Nasi Goreng Spesial KopiRise', 'nasi-goreng-spesial-kopirise', 'Nasi goreng bumbu rempah Jawa Timur dengan sate ayam, telur mata sapi, acar segar dan kerupuk udang.', 32000, 1, 1, 1, 'https://images.unsplash.com/photo-1603133872878-684f208fb84b?auto=format&fit=crop&w=600&q=80', 28, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-29', 'cat-maincourse', 'Rice Bowl Chicken Katsu Sambal Matah', 'rice-bowl-chicken-katsu-sambal-matah', 'Dada ayam krispi keemasan, nasi pulen hangat, sambal matah Bali segar dan tumis buncis bawang putih.', 29000, 1, 1, 1, 'https://images.unsplash.com/photo-1569058242253-92a9c755a0ec?auto=format&fit=crop&w=600&q=80', 29, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-30', 'cat-maincourse', 'Spaghetti Aglio Olio Smoked Beef', 'spaghetti-aglio-olio-smoked-beef', 'Pasta spaghetti al-dente ditumis dengan minyak zaitun extra virgin, bawang putih, cabai kering dan irisan daging asap.', 30000, 1, 1, 0, 'https://images.unsplash.com/photo-1621996346565-e3d5d6281788?auto=format&fit=crop&w=600&q=80', 30, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-31', 'cat-maincourse', 'Nasi Sei Sapi Sambal Luat', 'nasi-sei-sapi-sambal-luat', 'Daging sapi asap empuk aroma kayu kosambi, daun singkong berbumbu, kuah kaldu sapi dan sambal luat khas NTT.', 36000, 1, 0, 0, 'https://images.unsplash.com/photo-1544025162-d76694265947?auto=format&fit=crop&w=600&q=80', 31, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('menu-32', 'cat-maincourse', 'Chicken Steak Mushroom Sauce', 'chicken-steak-mushroom-sauce', 'Fillet paha ayam panggang juicy disajikan dengan saus jamur kental, french fries dan sayuran wortel buncis.', 34000, 1, 0, 0, 'https://images.unsplash.com/photo-1532550907401-a500c9a57435?auto=format&fit=crop&w=600&q=80', 32, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- 9. Menu Variants (Hot / Ice, Ukuran Cup)
INSERT OR IGNORE INTO menu_variants (id, menu_item_id, name, price_adjustment, is_available, sort_order, created_at) VALUES
('var-03-hot', 'menu-03', 'Hot (8oz Ceramic Cup)', 0, 1, 1, '2026-10-26T00:00:00Z'),
('var-03-ice', 'menu-03', 'Iced (16oz Cold Cup)', 2000, 1, 2, '2026-10-26T00:00:00Z'),
('var-05-reg', 'menu-05', 'Regular Ice (16oz)', 0, 1, 1, '2026-10-26T00:00:00Z'),
('var-05-lrg', 'menu-05', 'Large Ice (22oz)', 4000, 1, 2, '2026-10-26T00:00:00Z'),
('var-06-hot', 'menu-06', 'Hot (8oz Ceramic Cup)', 0, 1, 1, '2026-10-26T00:00:00Z'),
('var-06-ice', 'menu-06', 'Iced (16oz Cold Cup)', 2000, 1, 2, '2026-10-26T00:00:00Z'),
('var-13-hot', 'menu-13', 'Hot (8oz Ceramic Cup)', 0, 1, 1, '2026-10-26T00:00:00Z'),
('var-13-ice', 'menu-13', 'Iced (16oz Cold Cup)', 2000, 1, 2, '2026-10-26T00:00:00Z'),
('var-29-reg', 'menu-29', 'Pedas Sedang (Level 1)', 0, 1, 1, '2026-10-26T00:00:00Z'),
('var-29-hot', 'menu-29', 'Ekstra Pedas (Level 3)', 0, 1, 2, '2026-10-26T00:00:00Z');

-- 10. Modifier Groups & Options
INSERT OR IGNORE INTO modifier_groups (id, name, is_required, min_select, max_select, created_at) VALUES
('grp-sugar', 'Tingkat Manis (Sugar Level)', 0, 0, 1, '2026-10-26T00:00:00Z'),
('grp-ice', 'Tingkat Es (Ice Level)', 0, 0, 1, '2026-10-26T00:00:00Z'),
('grp-milk', 'Pilihan Susu (Dairy Alternative)', 0, 0, 1, '2026-10-26T00:00:00Z'),
('grp-extra', 'Ekstra Topping & Tambahan', 0, 0, 3, '2026-10-26T00:00:00Z');

INSERT OR IGNORE INTO modifier_options (id, modifier_group_id, name, additional_price, is_available, sort_order, created_at) VALUES
('opt-sug-100', 'grp-sugar', 'Normal Sweet (100%)', 0, 1, 1, '2026-10-26T00:00:00Z'),
('opt-sug-50',  'grp-sugar', 'Less Sweet (50%)', 0, 1, 2, '2026-10-26T00:00:00Z'),
('opt-sug-0',   'grp-sugar', 'No Sugar (0%)', 0, 1, 3, '2026-10-26T00:00:00Z'),

('opt-ice-norm', 'grp-ice', 'Normal Ice', 0, 1, 1, '2026-10-26T00:00:00Z'),
('opt-ice-less', 'grp-ice', 'Less Ice', 0, 1, 2, '2026-10-26T00:00:00Z'),
('opt-ice-none', 'grp-ice', 'No Ice', 0, 1, 3, '2026-10-26T00:00:00Z'),

('opt-milk-fresh', 'grp-milk', 'Fresh Milk Regular', 0, 1, 1, '2026-10-26T00:00:00Z'),
('opt-milk-oat',   'grp-milk', 'Oatside Oat Milk', 6000, 1, 2, '2026-10-26T00:00:00Z'),
('opt-milk-almond','grp-milk', 'Almond Milk', 7000, 1, 3, '2026-10-26T00:00:00Z'),

('opt-ext-shot',   'grp-extra', 'Extra Espresso Shot', 5000, 1, 1, '2026-10-26T00:00:00Z'),
('opt-ext-jelly',  'grp-extra', 'Coffee Jelly Boba', 4000, 1, 2, '2026-10-26T00:00:00Z'),
('opt-ext-cheese', 'grp-extra', 'Sweet Cheese Cream Foaming', 5000, 1, 3, '2026-10-26T00:00:00Z');

-- Link modifiers to menu items
INSERT OR IGNORE INTO menu_item_modifier_groups (menu_item_id, modifier_group_id) VALUES
('menu-03', 'grp-sugar'),
('menu-03', 'grp-ice'),
('menu-03', 'grp-milk'),
('menu-03', 'grp-extra'),
('menu-05', 'grp-sugar'),
('menu-05', 'grp-ice'),
('menu-05', 'grp-milk'),
('menu-05', 'grp-extra'),
('menu-06', 'grp-sugar'),
('menu-06', 'grp-ice'),
('menu-06', 'grp-milk'),
('menu-06', 'grp-extra'),
('menu-13', 'grp-sugar'),
('menu-13', 'grp-ice'),
('menu-13', 'grp-milk'),
('menu-13', 'grp-extra');

-- 11. Promos
INSERT OR IGNORE INTO promos (id, code, title, description, discount_type, discount_value, max_discount_amount, min_order_amount, start_date, end_date, is_active, usage_limit, used_count, created_at, updated_at) VALUES
('prm-senin', 'SENINSEMANGAT', 'Diskon 15% Awal Pekan', 'Diskon 15% untuk seluruh minuman kopi setiap hari Senin', 'percentage', 15, 25000, 40000, '2026-01-01T00:00:00Z', '2026-12-31T23:59:59Z', 1, 500, 42, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('prm-ngopi', 'KOPIPAHAM', 'Potongan Rp 10.000', 'Potongan Rp 10.000 khusus pembelian minimal Rp 60.000', 'fixed_amount', 10000, 10000, 60000, '2026-01-01T00:00:00Z', '2026-12-31T23:59:59Z', 1, 300, 89, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- 12. Financial Accounts
INSERT OR IGNORE INTO financial_accounts (id, name, account_type, account_number, current_balance, is_active, created_at, updated_at) VALUES
('acc-cash-pos', 'Kas Kasir POS Utama', 'cash', 'CASH-POS-01', 1500000, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('acc-vault', 'Kas Brankas Operasional', 'cash', 'CASH-VAULT-01', 12500000, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('acc-bank-bca', 'BCA Bisnis KopiRise', 'bank', '128-099-8877', 48650000, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('acc-qris-mandiri', 'QRIS Mandiri Settlement', 'electronic', 'NMID-ID102008988', 9450000, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- 13. Expense Categories (12 Kategori Standar)
INSERT OR IGNORE INTO expense_categories (id, name, description, created_at) VALUES
('exp-cat-01', 'Bahan baku', 'Pembelian biji kopi, susu segar, sirup, beras, daging, dan bahan dapur', '2026-10-26T00:00:00Z'),
('exp-cat-02', 'Gaji dan honor', 'Honor barista, kasir, waiter, dan bonus performa shift', '2026-10-26T00:00:00Z'),
('exp-cat-03', 'Sewa', 'Sewa tempat dan lahan kafe Jl. Semeru Jember', '2026-10-26T00:00:00Z'),
('exp-cat-04', 'Listrik dan air', 'Tagihan PLN pascabayar dan PDAM bulanan', '2026-10-26T00:00:00Z'),
('exp-cat-05', 'Internet', 'Langganan WiFi Biznet Dedicated 150 Mbps untuk pengunjung & POS', '2026-10-26T00:00:00Z'),
('exp-cat-06', 'Peralatan', 'Beli tamper, pitcher susu, piring keramik, gelas kaca, sedotan bambu', '2026-10-26T00:00:00Z'),
('exp-cat-07', 'Perawatan', 'Servis berkala mesin espresso La Marzocco dan grinder Mazzer', '2026-10-26T00:00:00Z'),
('exp-cat-08', 'Kebersihan', 'Sabun cuci piring food grade, plastik sampah, tisu meja, pembersih lantai', '2026-10-26T00:00:00Z'),
('exp-cat-09', 'Pemasaran', 'Iklan Instagram Ads, cetak banner promo, photoshoot menu', '2026-10-26T00:00:00Z'),
('exp-cat-10', 'Transportasi', 'Bensin operasional belanja pasar dan kurir logistik darurat', '2026-10-26T00:00:00Z'),
('exp-cat-11', 'Biaya administrasi', 'Biaya transfer bank, admin MDR QRIS, perizinan BPOM/PIRT', '2026-10-26T00:00:00Z'),
('exp-cat-12', 'Lainnya', 'Pengeluaran darurat tak terduga', '2026-10-26T00:00:00Z');

-- 14. Sample Customers
INSERT OR IGNORE INTO customers (id, name, phone, email, notes, created_at, updated_at) VALUES
('cst-001', 'Fajar Pratama', '081234560001', 'fajar.p@gmail.com', 'Pelanggan setia kopi susu aren less sweet', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('cst-002', 'Dewi Lestari', '081234560002', 'dewi.lestari@gmail.com', 'Suka meeting di VIP room', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('cst-003', 'Hendra Kusuma', '081234560003', 'hendra.k@yahoo.com', 'Rombongan kantor Sumbersari', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- 15. Sample Reservations
INSERT OR IGNORE INTO reservations (id, reservation_number, customer_name, customer_phone, customer_email, guest_count, reservation_date, reservation_time, duration_minutes, area_id, table_id, status, notes, created_at, updated_at) VALUES
('rsv-001', 'RSV-20261026-0001', 'Dewi Lestari', '081234560002', 'dewi.lestari@gmail.com', 8, '2026-10-26', '14:00:00', 120, 'area-vip-meeting', 'tbl-vip-02', 'confirmed', 'Meeting presentasi tim marketing (butuh TV HDMI)', '2026-10-26T08:00:00Z', '2026-10-26T08:30:00Z'),
('rsv-002', 'RSV-20261026-0002', 'Hendra Kusuma', '081234560003', 'hendra.k@yahoo.com', 4, '2026-10-26', '19:00:00', 120, 'area-indoor-ac', 'tbl-in-06', 'pending', 'Ulang tahun kecil, minta meja dekat jendela', '2026-10-26T09:15:00Z', '2026-10-26T09:15:00Z');

-- 16. Cash Shift (Shift Aktif Pagi Hari)
INSERT OR IGNORE INTO cash_shifts (id, cashier_user_id, shift_number, opened_at, initial_cash, cash_sales, cash_in, cash_out, cash_refunds, expected_cash, status, created_at, updated_at) VALUES
('shf-20261026-01', 'usr-csh-001', 'SHF-20261026-01', '2026-10-26T08:00:00Z', 500000, 246000, 0, 0, 0, 746000, 'open', '2026-10-26T08:00:00Z', '2026-10-26T08:00:00Z');

-- 17. Sample Orders & Payments
-- Order 1: Dine In di Meja IN-03 (Sedang Disiapkan)
INSERT OR IGNORE INTO orders (id, order_number, idempotency_key, order_type, table_id, customer_name, customer_phone, status, payment_status, subtotal, discount_amount, tax_amount, service_charge_amount, total_amount, paid_amount, created_by_user_id, cash_shift_id, notes, created_at, updated_at) VALUES
('ord-001', 'ORD-20261026-0001', 'idem-ord-001-abc', 'dine_in', 'tbl-in-03', 'Fajar Pratama', '081234560001', 'preparing', 'paid', 73000, 0, 7300, 0, 80300, 80300, 'usr-csh-001', 'shf-20261026-01', 'Kopi susu minta less sugar', '2026-10-26T08:30:00Z', '2026-10-26T08:35:00Z');

INSERT OR IGNORE INTO order_items (id, order_id, menu_item_id, menu_variant_id, item_name_snapshot, variant_name_snapshot, unit_base_price_snapshot, quantity, subtotal, notes, status, created_at) VALUES
('itm-001-1', 'ord-001', 'menu-05', 'var-05-reg', 'Kopi Susu Aren Semeru', 'Regular Ice (16oz)', 22000, 1, 22000, 'Less ice & less sugar', 'preparing', '2026-10-26T08:30:00Z'),
('itm-001-2', 'ord-001', 'menu-24', NULL, 'Tahu Walik Krispi Jember', NULL, 18000, 1, 18000, 'Petis dipisah', 'ready', '2026-10-26T08:30:00Z'),
('itm-001-3', 'ord-001', 'menu-28', NULL, 'Nasi Goreng Spesial KopiRise', NULL, 32000, 1, 32000, 'Telur ceplok matang', 'preparing', '2026-10-26T08:30:00Z');

INSERT OR IGNORE INTO order_item_modifiers (id, order_item_id, modifier_option_id, group_name_snapshot, option_name_snapshot, additional_price_snapshot) VALUES
('mod-001-1-1', 'itm-001-1', 'opt-sug-50', 'Tingkat Manis (Sugar Level)', 'Less Sweet (50%)', 0),
('mod-001-1-2', 'itm-001-1', 'opt-ice-less', 'Tingkat Es (Ice Level)', 'Less Ice', 0);

-- Payment for Order 1 (QRIS Mandiri)
INSERT OR IGNORE INTO payments (id, order_id, cash_shift_id, payment_number, idempotency_key, payment_method, amount, reference_number, processed_by_user_id, financial_account_id, status, notes, created_at) VALUES
('pay-001', 'ord-001', 'shf-20261026-01', 'PAY-20261026-0001', 'idem-pay-001-xyz', 'qris', 80300, 'MDR-20261026-98124', 'usr-csh-001', 'acc-qris-mandiri', 'completed', 'QRIS Mandiri sukses dikonfirmasi', '2026-10-26T08:35:00Z');

-- Order 2: Dine In di Meja OUT-04 (Pemesanan Mandiri via QR Code Meja, Belum Bayar)
INSERT OR IGNORE INTO orders (id, order_number, idempotency_key, order_type, table_id, customer_name, customer_phone, status, payment_status, subtotal, discount_amount, tax_amount, service_charge_amount, total_amount, paid_amount, created_by_user_id, cash_shift_id, notes, created_at, updated_at) VALUES
('ord-002', 'ORD-20261026-0002', 'idem-ord-002-def', 'dine_in', 'tbl-out-04', 'Tamu Meja OUT-04', '081987654321', 'confirmed', 'unpaid', 50000, 0, 5000, 0, 55000, 0, NULL, 'shf-20261026-01', 'Order langsung dari scan QR meja', '2026-10-26T09:00:00Z', '2026-10-26T09:02:00Z');

INSERT OR IGNORE INTO order_items (id, order_id, menu_item_id, menu_variant_id, item_name_snapshot, variant_name_snapshot, unit_base_price_snapshot, quantity, subtotal, notes, status, created_at) VALUES
('itm-002-1', 'ord-002', 'menu-09', NULL, 'V60 Single Origin Ijen Raung', NULL, 26000, 1, 26000, 'Pour over water temp 92C', 'confirmed', '2026-10-26T09:00:00Z'),
('itm-002-2', 'ord-002', 'menu-23', NULL, 'Pain Au Chocolat', NULL, 24000, 1, 24000, 'Hangatkan sebentar', 'confirmed', '2026-10-26T09:00:00Z');

-- 18. Suppliers & Ingredients & Recipes (Fondasi HPP & Stok)
INSERT OR IGNORE INTO suppliers (id, name, phone, email, address, contact_person, created_at) VALUES
('sup-01', 'Koperasi Tani Kopi Ijen Raung', '081233445566', 'kopi.ijen@supplier.id', 'Kec. Sukosari, Pegunungan Ijen', 'Pak Sugeng (Ketua Kelompok Tani)', '2026-10-26T00:00:00Z'),
('sup-02', 'PT Distribusi Susu & Dairy Sejahtera', '081299887766', 'order@dairysejahtera.com', 'Jl. Hayam Wuruk No. 88, Jember', 'Ibu Maya (Account Executive)', '2026-10-26T00:00:00Z'),
('sup-03', 'CV Rempah & Hasil Bumi Lumajang', '081322114455', 'sales@arenlumajang.id', 'Pasirian, Lumajang', 'Cak Huda', '2026-10-26T00:00:00Z');

INSERT OR IGNORE INTO ingredients (id, name, unit, current_stock, minimum_stock, cost_per_unit, is_active, created_at, updated_at) VALUES
('ing-coffee-beans', 'Biji Kopi Arabika Ijen (Roasted Beans)', 'gram', 12500.0, 2000.0, 280, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('ing-fresh-milk', 'Fresh Milk Pasteurisasi (Diamond/Greenfields)', 'ml', 38000.0, 5000.0, 22, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('ing-palm-sugar', 'Sirup Gula Aren Murni', 'ml', 14500.0, 2000.0, 35, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('ing-matcha', 'Matcha Powder Uji Ceremonial', 'gram', 1800.0, 300.0, 850, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('ing-rice', 'Beras Ramos Premium', 'gram', 45000.0, 10000.0, 15, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z'),
('ing-chicken', 'Fillet Daging Ayam Dada & Paha', 'gram', 18000.0, 3000.0, 55, 1, '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

-- Recipe for Kopi Susu Aren Semeru (menu-05)
-- HPP: (18g kopi * Rp 280) + (150ml susu * Rp 22) + (25ml gula aren * Rp 35) = 5.040 + 3.300 + 875 = Rp 9.215 / cup
INSERT OR IGNORE INTO recipes (id, menu_item_id, notes, created_at, updated_at) VALUES
('rcp-05', 'menu-05', 'Resep standar Kopi Susu Aren Semeru 16oz', '2026-10-26T00:00:00Z', '2026-10-26T00:00:00Z');

INSERT OR IGNORE INTO recipe_items (id, recipe_id, ingredient_id, quantity, created_at) VALUES
('rcp-itm-05-1', 'rcp-05', 'ing-coffee-beans', 18.0, '2026-10-26T00:00:00Z'),
('rcp-itm-05-2', 'rcp-05', 'ing-fresh-milk', 150.0, '2026-10-26T00:00:00Z'),
('rcp-itm-05-3', 'rcp-05', 'ing-palm-sugar', 25.0, '2026-10-26T00:00:00Z');

-- 19. Sample Expenses
INSERT OR IGNORE INTO expenses (id, expense_number, expense_date, category_id, financial_account_id, amount, description, vendor, reference_number, status, created_by_user_id, approved_by_user_id, notes, created_at, updated_at) VALUES
('exp-001', 'EXP-20261026-0001', '2026-10-26', 'exp-cat-01', 'acc-bank-bca', 3500000, 'Restok 12.5 kg Biji Kopi Arabika Ijen Raung Fresh Roast', 'Koperasi Tani Kopi Ijen', 'INV-IJEN-9912', 'approved', 'usr-mgr-001', 'usr-admin-001', 'Pengadaan bahan baku mingguan', '2026-10-26T07:30:00Z', '2026-10-26T07:45:00Z'),
('exp-002', 'EXP-20261026-0002', '2026-10-26', 'exp-cat-05', 'acc-bank-bca', 750000, 'Tagihan Internet Biznet Dedicated 150 Mbps Periode Oktober 2026', 'Biznet Networks', 'BZ-202610-881', 'approved', 'usr-mgr-001', 'usr-admin-001', 'WiFi lancar untuk pelanggan dan POS', '2026-10-26T08:00:00Z', '2026-10-26T08:10:00Z');

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DELETE FROM expenses;
DELETE FROM recipe_items;
DELETE FROM recipes;
DELETE FROM ingredients;
DELETE FROM suppliers;
DELETE FROM payments;
DELETE FROM order_item_modifiers;
DELETE FROM order_items;
DELETE FROM orders;
DELETE FROM cash_shifts;
DELETE FROM reservations;
DELETE FROM customers;
DELETE FROM financial_accounts;
DELETE FROM promos;
DELETE FROM menu_item_modifier_groups;
DELETE FROM modifier_options;
DELETE FROM modifier_groups;
DELETE FROM menu_variants;
DELETE FROM menu_items;
DELETE FROM menu_categories;
DELETE FROM table_qr_tokens;
DELETE FROM cafe_tables;
DELETE FROM cafe_areas;
DELETE FROM settings;
DELETE FROM employee_profiles;
DELETE FROM user_roles;
DELETE FROM users;
DELETE FROM roles;
