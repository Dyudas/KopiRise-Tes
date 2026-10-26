/**
 * Memformat angka ke representasi mata uang Rupiah Indonesia (Rp).
 * Disimpan sebagai INTEGER tanpa koma sen sesuai aturan finansial KopiRise.
 */
export function formatRupiah(amount: number): string {
  if (isNaN(amount)) return 'Rp 0';
  return 'Rp ' + Math.round(amount).toLocaleString('id-ID');
}

/**
 * Memformat string tanggal (YYYY-MM-DD atau ISO UTC) ke format Indonesia (Asia/Jakarta WIB).
 */
export function formatDateIndo(dateStr: string): string {
  if (!dateStr) return '-';
  try {
    const d = new Date(dateStr);
    return d.toLocaleDateString('id-ID', {
      weekday: 'long',
      day: 'numeric',
      month: 'long',
      year: 'numeric'
    });
  } catch {
    return dateStr;
  }
}

/**
 * Memformat waktu kedatangan / jam (HH:mm).
 */
export function formatTimeWIB(timeStr: string): string {
  if (!timeStr) return '';
  return `${timeStr} WIB`;
}
