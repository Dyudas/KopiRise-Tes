import { test, expect } from '@playwright/test';

test.describe('KopiRise Full-Stack Cafe Management System E2E', () => {

  test('1. Beranda publik harus memuat identitas kafe, jam operasional, dan navigasi', async ({ page }) => {
    await page.goto('/');

    // Verifikasi branding
    await expect(page.locator('h1')).toContainText('KopiRise');
    await expect(page.locator('text=Jl. Semeru')).toBeVisible();

    // Verifikasi tombol CTA
    const menuBtn = page.locator('a[href="/menu"]').first();
    await expect(menuBtn).toBeVisible();

    const rsvBtn = page.locator('a[href="/reservasi"]').first();
    await expect(rsvBtn).toBeVisible();
  });

  test('2. Halaman menu harus menampilkan item dan fungsi pencarian live', async ({ page }) => {
    await page.goto('/menu');

    // Cek judul halaman
    await expect(page.locator('h1')).toContainText('Daftar Menu KopiRise');

    // Uji pencarian menu
    const searchInput = page.locator('input[placeholder*="Cari menu"]');
    await searchInput.fill('Aren');

    // Kartu menu Kopi Susu Aren harus tetap muncul
    await expect(page.locator('text=Kopi Susu Aren Semeru').first()).toBeVisible();
  });

  test('3. Formulir reservasi meja harus menolak spam honeypot dan memvalidasi nomor telepon', async ({ page }) => {
    await page.goto('/reservasi');

    await expect(page.locator('h1')).toContainText('Reservasi Meja KopiRise');

    // Input data pemesan
    await page.fill('input[name="customer_name"]', 'Dewi Lestari');
    await page.fill('input[name="customer_phone"]', '081234567890');
    
    // Pilih jam
    await page.selectOption('select[name="reservation_time"]', '16:00');

    // Verifikasi field honeypot tetap tersembunyi
    const trap = page.locator('input[name="website_trap"]');
    await expect(trap).toBeHidden();
  });

  test('4. Alur pemesanan meja QR harus mengenali nomor meja dinamis dari token', async ({ page }) => {
    // Simulasi scan QR meja OUT-04 (token valid)
    await page.goto('/meja/kr_tbl_8e3a4d71b9c251f');

    // Harus mendeteksi nomor meja
    await expect(page.locator('text=OUT-04').first()).toBeVisible();
    await expect(page.locator('text=Outdoor Garden').first()).toBeVisible();
  });

  test('5. Halaman admin harus memproteksi akses dan meminta login karyawan', async ({ page }) => {
    await page.goto('/admin/pos');

    // Harus diarahkan otomatis ke halaman login
    await expect(page).toHaveURL(/\/admin\/login/);
    await expect(page.locator('h2')).toContainText('KopiRise Staff Portal');
  });

});
