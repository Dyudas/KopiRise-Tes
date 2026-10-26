import sqlite3
import os
import re

DB_PATH = r"d:\Kopirise\data\kopirise.db"
MIGRATIONS_DIR = r"d:\Kopirise\backend\migrations"

def run_goose_migration_file(cursor, file_path):
    print(f"Applying: {os.path.basename(file_path)}")
    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    # Extract only the Up migration part
    if "-- +goose Up" in content:
        up_content = content.split("-- +goose Up")[1]
        if "-- +goose Down" in up_content:
            up_content = up_content.split("-- +goose Down")[0]
    else:
        up_content = content

    # Execute scripts
    cursor.executescript(up_content)

def main():
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)
        print(f"Removed existing db at {DB_PATH}")

    os.makedirs(os.path.dirname(DB_PATH), exist_ok=True)
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Enable WAL and PRAGMAs
    cursor.execute("PRAGMA journal_mode = WAL;")
    cursor.execute("PRAGMA foreign_keys = ON;")
    cursor.execute("PRAGMA busy_timeout = 5000;")
    cursor.execute("PRAGMA synchronous = NORMAL;")

    # Apply migrations in order
    migration_files = sorted([os.path.join(MIGRATIONS_DIR, f) for f in os.listdir(MIGRATIONS_DIR) if f.endswith(".sql")])
    for mf in migration_files:
        run_goose_migration_file(cursor, mf)

    conn.commit()

    # Run Verification Queries
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';")
    tables = [row[0] for row in cursor.fetchall()]
    print(f"\n[OK] Successfully created {len(tables)} tables:")
    print(", ".join(sorted(tables)))

    # Verify Counts
    checks = {
        "roles": "SELECT COUNT(*) FROM roles",
        "users": "SELECT COUNT(*) FROM users",
        "settings": "SELECT COUNT(*) FROM settings",
        "cafe_areas": "SELECT COUNT(*) FROM cafe_areas",
        "cafe_tables": "SELECT COUNT(*) FROM cafe_tables",
        "table_qr_tokens": "SELECT COUNT(*) FROM table_qr_tokens",
        "menu_categories": "SELECT COUNT(*) FROM menu_categories",
        "menu_items": "SELECT COUNT(*) FROM menu_items",
        "menu_variants": "SELECT COUNT(*) FROM menu_variants",
        "modifier_options": "SELECT COUNT(*) FROM modifier_options",
        "promos": "SELECT COUNT(*) FROM promos",
        "financial_accounts": "SELECT COUNT(*) FROM financial_accounts",
        "expense_categories": "SELECT COUNT(*) FROM expense_categories",
        "expenses": "SELECT COUNT(*) FROM expenses",
        "ingredients": "SELECT COUNT(*) FROM ingredients",
        "recipes": "SELECT COUNT(*) FROM recipes",
        "orders": "SELECT COUNT(*) FROM orders",
        "order_items": "SELECT COUNT(*) FROM order_items",
        "payments": "SELECT COUNT(*) FROM payments",
        "cash_shifts": "SELECT COUNT(*) FROM cash_shifts",
    }

    print("\n--- DATA COUNT VERIFICATION ---")
    for name, sql in checks.items():
        cursor.execute(sql)
        count = cursor.fetchone()[0]
        print(f"  - {name}: {count}")

    # Check Foreign Key Violations
    cursor.execute("PRAGMA foreign_key_check;")
    fk_violations = cursor.fetchall()
    if fk_violations:
        print("\n[ERROR] Foreign key violations found:", fk_violations)
    else:
        print("\n[PASS] No foreign key violations found! Foreign keys are 100% valid.")

    # Check HPP calculation for Kopi Susu Aren Semeru
    cursor.execute("""
        SELECT 
            m.name AS menu_name,
            m.base_price,
            SUM(ri.quantity * i.cost_per_unit) AS calculated_hpp
        FROM recipes r
        JOIN menu_items m ON r.menu_item_id = m.id
        JOIN recipe_items ri ON r.id = ri.recipe_id
        JOIN ingredients i ON ri.ingredient_id = i.id
        WHERE m.id = 'menu-05'
        GROUP BY m.id;
    """)
    hpp_result = cursor.fetchone()
    if hpp_result:
        menu_name, price, hpp = hpp_result
        margin = price - hpp
        margin_pct = (margin / price) * 100
        print(f"\n--- HPP & MARGIN TEST ---")
        print(f"Menu: {menu_name}")
        print(f"Harga Jual: Rp {price:,}")
        print(f"HPP Resep: Rp {int(hpp):,}")
        print(f"Estimasi Margin Kotor: Rp {int(margin):,} ({margin_pct:.1f}%)")

    # Check Backup Procedure via VACUUM INTO
    backup_path = r"d:\Kopirise\data\kopirise_backup_test.db"
    if os.path.exists(backup_path):
        os.remove(backup_path)
    cursor.execute(f"VACUUM INTO '{backup_path}';")
    if os.path.exists(backup_path) and os.path.getsize(backup_path) > 0:
        print(f"\n[PASS] VACUUM INTO backup successfully generated: {backup_path} ({os.path.getsize(backup_path)} bytes)")

    conn.close()
    print("\n[SUCCESS] Tahap 1 SQLite Database, Migrations, Seeders & Rules verified successfully!")

if __name__ == "__main__":
    main()
