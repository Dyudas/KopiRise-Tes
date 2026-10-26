import { describe, it, expect } from 'vitest';
import { formatRupiah, formatDateIndo, formatTimeWIB } from './format';

describe('Format Rupiah Utility', () => {
  it('harus memformat angka integer positif ke Rupiah tanpa desimal', () => {
    expect(formatRupiah(25000)).toBe('Rp 25.000');
    expect(formatRupiah(1500000)).toBe('Rp 1.500.000');
    expect(formatRupiah(0)).toBe('Rp 0');
  });

  it('harus membulatkan nilai desimal jika ada ke integer terdekat', () => {
    expect(formatRupiah(22450.75)).toBe('Rp 22.451');
  });

  it('harus mengembalikan Rp 0 jika input NaN', () => {
    expect(formatRupiah(NaN)).toBe('Rp 0');
  });
});

describe('Format Waktu & Tanggal Indonesia', () => {
  it('harus memformat string jam dengan suffix WIB', () => {
    expect(formatTimeWIB('14:30')).toBe('14:30 WIB');
    expect(formatTimeWIB('')).toBe('');
  });

  it('harus memformat tanggal YYYY-MM-DD ke penamaan hari dan bulan Indonesia', () => {
    const formatted = formatDateIndo('2026-10-26');
    expect(formatted).toContain('2026');
    expect(formatted).toContain('Oktober');
  });
});
