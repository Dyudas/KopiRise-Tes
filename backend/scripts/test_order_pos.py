import sqlite3
import uuid
import datetime

DB_PATH = r"d:\Kopirise\data\kopirise.db"

def run_test():
    conn = sqlite3.connect(DB_PATH)
    conn.execute("PRAGMA foreign_keys = ON;")
    cursor = conn.cursor()

    print("=== TEST TAHAP 4: ORDER, QR, DAN POS DI DATABASE SQLITE ===")

    # 1. Ambil kasir dan akun kas
    cursor.execute("""
        SELECT u.id, u.name 
        FROM users u 
        JOIN user_roles ur ON u.id = ur.user_id 
        WHERE ur.role_id = 'cashier' 
        LIMIT 1;
    """)
    cashier = cursor.fetchone()
    if not cashier:
        raise Exception("Kasir tidak ditemukan di database")
    cashier_id, cashier_name = cashier
    print(f"[OK] Kasir: {cashier_name} ({cashier_id})")

    cursor.execute("SELECT id, name FROM financial_accounts WHERE account_type = 'cash' LIMIT 1;")
    account = cursor.fetchone()
    account_id, account_name = account
    print(f"[OK] Akun Keuangan Kas: {account_name} ({account_id})")

    # 2. Buka Shift Kasir
    shift_id = "shift-test-" + str(uuid.uuid4())[:8]
    shift_num = "SHF-" + datetime.datetime.now().strftime("%Y%m%d") + "-" + str(uuid.uuid4())[:4].upper()
    start_cash = 500000
    now_utc = datetime.datetime.now(datetime.timezone.utc).isoformat()

    cursor.execute("""
        INSERT INTO cash_shifts (id, cashier_user_id, shift_number, opened_at, initial_cash, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, 'open', ?, ?)
    """, (shift_id, cashier_id, shift_num, now_utc, start_cash, now_utc, now_utc))
    conn.commit()
    print(f"[OK] Buka Shift Kasir #{shift_num} ID: {shift_id} dengan Modal Awal Rp {start_cash:,}")

    # 3. Buat Order POS (Dine-in di Meja IN-01)
    cursor.execute("SELECT id, table_number FROM cafe_tables WHERE table_number = 'IN-01' LIMIT 1;")
    tbl = cursor.fetchone()
    tbl_id, tbl_no = tbl

    order_id = "ord-test-" + str(uuid.uuid4())[:8]
    order_num = "ORD-" + datetime.datetime.now().strftime("%Y%m%d") + "-" + uuid.uuid4().hex[:4].upper()
    idem_key = "idem-" + str(uuid.uuid4())

    # Ambil 2 menu item nyata
    cursor.execute("SELECT id, name, base_price FROM menu_items WHERE is_available = 1 LIMIT 2;")
    menu_items = cursor.fetchall()
    item1_id, item1_name, item1_price = menu_items[0]
    item2_id, item2_name, item2_price = menu_items[1]

    # Ambil modifier untuk item 1
    cursor.execute("SELECT id, name, additional_price FROM modifier_options LIMIT 1;")
    mod = cursor.fetchone()
    mod_id, mod_name, mod_price = mod

    qty1 = 2
    qty2 = 1
    subtotal = (item1_price + mod_price) * qty1 + (item2_price * qty2)
    tax_10 = (subtotal * 10) // 100
    total = subtotal + tax_10

    print(f"Detail Order: {item1_name} x{qty1} (@Rp {item1_price:,} + {mod_name} @Rp {mod_price:,}), {item2_name} x{qty2} (@Rp {item2_price:,})")
    print(f"Subtotal: Rp {subtotal:,} | Pajak PB1 (10%): Rp {tax_10:,} | Total Akhir: Rp {total:,}")

    cursor.execute("""
        INSERT INTO orders (
            id, order_number, idempotency_key, order_type, table_id, customer_name, status,
            payment_status, subtotal, discount_amount, tax_amount, total_amount,
            paid_amount, created_by_user_id, cash_shift_id, created_at, updated_at
        ) VALUES (?, ?, ?, 'dine_in', ?, 'Budi Santoso', 'pending', 'unpaid', ?, 0, ?, ?, 0, ?, ?, ?, ?)
    """, (order_id, order_num, idem_key, tbl_id, subtotal, tax_10, total, cashier_id, shift_id, now_utc, now_utc))

    # Snapshot Order Items
    order_item1_id = "oi-test-" + str(uuid.uuid4())[:8]
    order_item2_id = "oi-test-" + str(uuid.uuid4())[:8]

    cursor.execute("""
        INSERT INTO order_items (
            id, order_id, menu_item_id, item_name_snapshot, unit_base_price_snapshot,
            quantity, subtotal, notes, status, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, 'Less sugar', 'pending', ?)
    """, (order_item1_id, order_id, item1_id, item1_name, item1_price, qty1, (item1_price + mod_price) * qty1, now_utc))

    cursor.execute("""
        INSERT INTO order_items (
            id, order_id, menu_item_id, item_name_snapshot, unit_base_price_snapshot,
            quantity, subtotal, notes, status, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, '', 'pending', ?)
    """, (order_item2_id, order_id, item2_id, item2_name, item2_price, qty2, item2_price * qty2, now_utc))

    # Snapshot Modifier Item
    cursor.execute("""
        INSERT INTO order_item_modifiers (
            id, order_item_id, modifier_option_id, group_name_snapshot, option_name_snapshot, additional_price_snapshot
        ) VALUES (?, ?, ?, 'Extra Topping', ?, ?)
    """, (str(uuid.uuid4()), order_item1_id, mod_id, mod_name, mod_price))

    # Update Table status menjadi occupied
    cursor.execute("UPDATE cafe_tables SET status = 'occupied' WHERE id = ?", (tbl_id,))
    conn.commit()
    print(f"[OK] Order {order_num} berhasil dibuat atomik dengan snapshot harga!")

    # 4. KDS (Kitchen Display System) - Barista update item ke preparing -> ready
    cursor.execute("UPDATE order_items SET status = 'preparing' WHERE order_id = ?", (order_id,))
    cursor.execute("UPDATE orders SET status = 'preparing' WHERE id = ?", (order_id,))
    conn.commit()
    print("[OK] KDS Barista: Order status diperbarui menjadi 'preparing'")

    # 5. Pembayaran Tunai di POS
    pay_id = "pay-test-" + str(uuid.uuid4())[:8]
    pay_num = "PAY-" + datetime.datetime.now().strftime("%Y%m%d") + "-" + uuid.uuid4().hex[:4].upper()
    cash_tendered = total + 10000 # Bayar lebih Rp 10.000
    change_due = 10000

    cursor.execute("""
        INSERT INTO payments (
            id, order_id, cash_shift_id, payment_number, idempotency_key, payment_method,
            amount, cash_tendered, change_due, processed_by_user_id, financial_account_id,
            status, created_at
        ) VALUES (?, ?, ?, ?, ?, 'cash', ?, ?, ?, ?, ?, 'completed', ?)
    """, (pay_id, order_id, shift_id, pay_num, str(uuid.uuid4()), total, cash_tendered, change_due, cashier_id, account_id, now_utc))

    cursor.execute("""
        UPDATE orders SET payment_status = 'paid', paid_amount = ?, status = 'completed', updated_at = ?
        WHERE id = ?
    """, (total, now_utc, order_id))

    # Update total penjualan tunai di shift kasir
    cursor.execute("""
        UPDATE cash_shifts SET cash_sales = cash_sales + ? WHERE id = ?
    """, (total, shift_id))
    conn.commit()
    print(f"[OK] Pembayaran {pay_num} sukses. Diterima: Rp {cash_tendered:,}, Kembalian: Rp {change_due:,}")

    # 6. Tutup Shift Kasir
    # Kas seharusnya = kas_awal (500.000) + cash_sales (total)
    expected_cash = start_cash + total
    actual_physical_cash = expected_cash # Uang fisik pas sesuai hitungan sistem
    variance = actual_physical_cash - expected_cash

    cursor.execute("""
        UPDATE cash_shifts
        SET closed_at = ?, actual_cash = ?, expected_cash = ?, discrepancy = ?, status = 'closed', updated_at = ?
        WHERE id = ?
    """, (now_utc, actual_physical_cash, expected_cash, variance, now_utc, shift_id))
    conn.commit()
    print(f"[OK] Tutup Shift Berhasil! Kas Seharusnya: Rp {expected_cash:,}, Kas Fisik: Rp {actual_physical_cash:,}, Selisih: Rp {variance:,}")

    # 7. Verifikasi Foreign Key Integrity
    cursor.execute("PRAGMA foreign_key_check;")
    fk_errors = cursor.fetchall()
    if fk_errors:
        raise Exception(f"Foreign key violations: {fk_errors}")
    print("[OK] Foreign key check: 0 VIOLATIONS. Seluruh relasi integritas terverifikasi.")

    # 8. Verifikasi Idempotency Key Constraint
    try:
        cursor.execute("""
            INSERT INTO orders (id, order_number, idempotency_key, order_type, table_id, customer_name, created_at, updated_at)
            VALUES ('dummy', 'ORD-DUPLICATE', ?, 'dine_in', ?, 'Test', ?, ?)
        """, (idem_key, tbl_id, now_utc, now_utc))
        raise Exception("Idempotency constraint GAGAL bekerja!")
    except sqlite3.IntegrityError:
        print("[OK] Idempotency Key constraint BERHASIL mencegah duplicate order!")

    print("\n>>> TAHAP 4 VERIFIKASI SELESAI & 100% LULUS PENGUJIAN! <<<")

if __name__ == "__main__":
    run_test()
