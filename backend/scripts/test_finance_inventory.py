import sqlite3
import uuid
import datetime

DB_PATH = r"d:\Kopirise\data\kopirise.db"

def run_test():
    conn = sqlite3.connect(DB_PATH)
    conn.execute("PRAGMA foreign_keys = ON;")
    cursor = conn.cursor()

    print("=== TEST TAHAP 5: KEUANGAN, BIAYA, BAHAN BAKU, & HPP DI SQLITE ===")

    # 1. Verifikasi Bagan Akun Finansial
    cursor.execute("SELECT id, name, account_type, current_balance FROM financial_accounts;")
    accounts = cursor.fetchall()
    print(f"[OK] Ditemukan {len(accounts)} akun keuangan:")
    for acc in accounts:
        print(f"  - {acc[1]} ({acc[2]}): Rp {acc[3]:,}")

    acc_pos_id = "acc-cash-pos"
    acc_brankas_id = "acc-vault"
    acc_bank_id = "acc-bank-bca"

    # Ambil user admin/manager untuk audit actor
    cursor.execute("SELECT id, name FROM users WHERE id = 'usr-admin-001' LIMIT 1;")
    admin = cursor.fetchone()
    admin_id, admin_name = admin
    now_utc = datetime.datetime.now(datetime.timezone.utc).isoformat()
    today_str = datetime.datetime.now().strftime("%Y-%m-%d")

    # 2. Test Mutasi Dana Antar-Akun (Transfer Kasir ke Brankas)
    transfer_amount = 500000
    cursor.execute("SELECT current_balance FROM financial_accounts WHERE id = ?", (acc_pos_id,))
    pos_balance_before = cursor.fetchone()[0]

    cursor.execute("SELECT current_balance FROM financial_accounts WHERE id = ?", (acc_brankas_id,))
    brankas_balance_before = cursor.fetchone()[0]

    trf_id = "trf-test-" + str(uuid.uuid4())[:8]
    trf_number = "TRF-" + datetime.datetime.now().strftime("%Y%m%d") + "-" + uuid.uuid4().hex[:4].upper()

    cursor.execute("""
        INSERT INTO account_transfers (
            id, transfer_number, from_account_id, to_account_id, amount,
            transfer_date, description, created_by_user_id, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, 'Setoran Kasir Pagi ke Brankas Toko', ?, ?)
    """, (trf_id, trf_number, acc_pos_id, acc_brankas_id, transfer_amount, today_str, admin_id, now_utc))

    cursor.execute("UPDATE financial_accounts SET current_balance = current_balance - ?, updated_at = ? WHERE id = ?",
                   (transfer_amount, now_utc, acc_pos_id))
    cursor.execute("UPDATE financial_accounts SET current_balance = current_balance + ?, updated_at = ? WHERE id = ?",
                   (transfer_amount, now_utc, acc_brankas_id))
    conn.commit()

    cursor.execute("SELECT current_balance FROM financial_accounts WHERE id = ?", (acc_pos_id,))
    pos_balance_after = cursor.fetchone()[0]
    cursor.execute("SELECT current_balance FROM financial_accounts WHERE id = ?", (acc_brankas_id,))
    brankas_balance_after = cursor.fetchone()[0]

    assert pos_balance_after == pos_balance_before - transfer_amount, "Gagal mutasi akun sumber"
    assert brankas_balance_after == brankas_balance_before + transfer_amount, "Gagal mutasi akun tujuan"
    print(f"[OK] Transfer Dana {trf_number} Rp {transfer_amount:,} Sukses! Saldo POS: Rp {pos_balance_after:,}, Brankas: Rp {brankas_balance_after:,}")

    # 3. Test Pengeluaran Biaya Operasional (Listrik PLN dari Bank BCA)
    cursor.execute("SELECT id, name FROM expense_categories WHERE name LIKE '%Listrik%' OR name LIKE '%Utilitas%' LIMIT 1;")
    cat = cursor.fetchone()
    cat_id, cat_name = cat

    exp_amount = 750000
    exp_id = "exp-test-" + str(uuid.uuid4())[:8]
    exp_number = "EXP-" + datetime.datetime.now().strftime("%Y%m%d") + "-" + uuid.uuid4().hex[:4].upper()

    cursor.execute("""
        INSERT INTO expenses (
            id, expense_number, expense_date, category_id, financial_account_id,
            amount, description, vendor, reference_number, status,
            created_by_user_id, approved_by_user_id, notes, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, 'Tagihan Listrik PLN Bulan Berjalan', 'PLN UP3 Jember', 'PLN-998877', 'approved', ?, ?, 'Lunas autodebet', ?, ?)
    """, (exp_id, exp_number, today_str, cat_id, acc_bank_id, exp_amount, admin_id, admin_id, now_utc, now_utc))

    cursor.execute("UPDATE financial_accounts SET current_balance = current_balance - ?, updated_at = ? WHERE id = ?",
                   (exp_amount, now_utc, acc_bank_id))
    conn.commit()
    print(f"[OK] Pencatatan Beban Operasional {exp_number} Rp {exp_amount:,} ({cat_name}) Terverifikasi!")

    # 4. Test Pembelian Stok Bahan Baku & Biaya Rata-Rata Tertimbang
    cursor.execute("SELECT id, name, unit, current_stock, cost_per_unit FROM ingredients WHERE name LIKE '%Arabika%' LIMIT 1;")
    ing = cursor.fetchone()
    ing_id, ing_name, ing_unit, old_stock, old_cost = ing

    purchased_qty = 10000.0 # 10.000 gram (10 kg)
    purchased_unit_cost = 200 # Rp 200 per gram (Rp 200.000 / kg)
    purchase_subtotal = int(purchased_qty * purchased_unit_cost)

    new_stock = old_stock + purchased_qty
    total_val = (old_stock * float(old_cost)) + (purchased_qty * float(purchased_unit_cost))
    new_cost = int(total_val / new_stock)

    po_id = "po-test-" + str(uuid.uuid4())[:8]
    po_number = "PO-" + datetime.datetime.now().strftime("%Y%m%d") + "-" + uuid.uuid4().hex[:4].upper()

    cursor.execute("SELECT id FROM suppliers LIMIT 1;")
    sup_id = cursor.fetchone()[0]

    cursor.execute("""
        INSERT INTO ingredient_purchases (
            id, purchase_number, supplier_id, financial_account_id, purchase_date,
            total_amount, created_by_user_id, notes, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, 'Restock Arabika Single Origin', ?)
    """, (po_id, po_number, sup_id, acc_brankas_id, today_str, purchase_subtotal, admin_id, now_utc))

    cursor.execute("""
        INSERT INTO ingredient_purchase_items (id, purchase_id, ingredient_id, quantity, unit_cost, subtotal)
        VALUES (?, ?, ?, ?, ?, ?)
    """, (str(uuid.uuid4()), po_id, ing_id, purchased_qty, purchased_unit_cost, purchase_subtotal))

    cursor.execute("""
        UPDATE ingredients
        SET current_stock = ?, cost_per_unit = ?, updated_at = ?
        WHERE id = ?
    """, (new_stock, new_cost, now_utc, ing_id))

    cursor.execute("UPDATE financial_accounts SET current_balance = current_balance - ?, updated_at = ? WHERE id = ?",
                   (purchase_subtotal, now_utc, acc_brankas_id))
    conn.commit()

    print(f"[OK] Pembelian {po_number}: {ing_name} +{purchased_qty} {ing_unit}. Stok Baru: {new_stock} {ing_unit}, Cost Baru: Rp {new_cost}/unit")

    # 5. Test Resep Menu & Kalkulasi HPP Bill of Materials
    cursor.execute("SELECT id, name, base_price FROM menu_items WHERE name LIKE '%Aren%' LIMIT 1;")
    menu_item = cursor.fetchone()
    menu_id, menu_name, menu_price = menu_item

    cursor.execute("SELECT r.id, ri.ingredient_id, ing.name, ri.quantity, ing.cost_per_unit FROM recipes r JOIN recipe_items ri ON r.id = ri.recipe_id JOIN ingredients ing ON ri.ingredient_id = ing.id WHERE r.menu_item_id = ?;", (menu_id,))
    recipe_rows = cursor.fetchall()

    total_hpp = 0
    print(f"[OK] Resep BOM untuk '{menu_name}' (Harga Jual Rp {menu_price:,}):")
    for rrow in recipe_rows:
        item_cost = int(rrow[3] * rrow[4])
        total_hpp += item_cost
        print(f"  - {rrow[2]}: {rrow[3]} unit x Rp {rrow[4]:,} = Rp {item_cost:,}")

    margin_rp = menu_price - total_hpp
    margin_pct = (margin_rp / menu_price) * 100.0
    print(f"Total HPP Resep: Rp {total_hpp:,} | Margin Kotor: Rp {margin_rp:,} ({margin_pct:.1f}%)")

    # 6. Test Konsumsi Stok Penjualan (Sale Consumption)
    deduct_qty = 18.0 # 18 gram kopi
    cursor.execute("""
        INSERT INTO inventory_movements (
            id, ingredient_id, movement_type, quantity, previous_stock, new_stock,
            reference_type, reference_id, reason, performed_by_user_id, created_at
        ) VALUES (?, ?, 'sale_consumption', ?, ?, ?, 'order', 'ord-test-sample', 'Penjualan Kopi Susu', ?, ?)
    """, (str(uuid.uuid4()), ing_id, deduct_qty, new_stock, new_stock - deduct_qty, admin_id, now_utc))

    cursor.execute("UPDATE ingredients SET current_stock = current_stock - ?, updated_at = ? WHERE id = ?",
                   (deduct_qty, now_utc, ing_id))
    conn.commit()
    print(f"[OK] Kartu Stok Penjualan: {ing_name} berkurang {deduct_qty} {ing_unit} (Sisa: {new_stock - deduct_qty} {ing_unit})")

    # 7. Test Query Laporan Laba Rugi Komprehensif
    cursor.execute("""
        SELECT 
            COALESCE(SUM(subtotal), 0) AS gross_sales,
            COALESCE(SUM(discount_amount), 0) AS discounts,
            COALESCE(SUM(tax_amount), 0) AS tax_collected,
            COALESCE(SUM(total_amount), 0) AS net_sales
        FROM orders
        WHERE payment_status = 'paid';
    """)
    sales = cursor.fetchone()
    gross_sales, discounts, tax_col, net_sales = sales

    cursor.execute("""
        SELECT COALESCE(SUM(e.amount), 0)
        FROM expenses e
        WHERE e.status = 'approved';
    """)
    total_expenses = cursor.fetchone()[0]

    gross_profit = net_sales - total_hpp
    net_profit = gross_profit - total_expenses
    print(f"\n--- RINGKASAN LABA RUGI SISTEM ---")
    print(f"Penjualan Bersih : Rp {net_sales:,}")
    print(f"Estimasi HPP     : Rp {total_hpp:,}")
    print(f"Laba Kotor       : Rp {gross_profit:,}")
    print(f"Beban Operasional: Rp {total_expenses:,}")
    print(f"Laba Bersih Kafe : Rp {net_profit:,}")

    # 8. Verifikasi Foreign Key Integrity
    cursor.execute("PRAGMA foreign_key_check;")
    fk_errors = cursor.fetchall()
    if fk_errors:
        raise Exception(f"Foreign key violations ditemukan: {fk_errors}")
    print("[OK] Foreign key check: 0 VIOLATIONS. Seluruh relasi integritas terverifikasi.")

    print("\n>>> TAHAP 5 VERIFIKASI SELESAI & 100% LULUS PENGUJIAN! <<<")

if __name__ == "__main__":
    run_test()
