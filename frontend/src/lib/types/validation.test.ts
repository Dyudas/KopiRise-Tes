import { describe, it, expect } from 'vitest';
import { z } from 'zod';

const reservationSchema = z.object({
  customer_name: z.string().min(2, 'Nama pemesan minimal 2 karakter'),
  customer_phone: z.string().min(9, 'Nomor WhatsApp tidak valid (minimal 9 digit)'),
  customer_email: z.string().email('Format email tidak valid').optional().or(z.literal('')),
  reservation_date: z.string().min(10, 'Tanggal kunjungan wajib diisi'),
  reservation_time: z.string().min(5, 'Jam kedatangan wajib dipilih'),
  guest_count: z.coerce.number().min(1, 'Jumlah tamu minimal 1 orang').max(50, 'Maksimal 50 orang per pemesanan'),
  website_trap: z.string().optional()
});

describe('Zod Validation: Reservation Form', () => {
  it('harus menerima input reservasi yang valid', () => {
    const validData = {
      customer_name: 'Budi Santoso',
      customer_phone: '081234567890',
      customer_email: 'budi@example.com',
      reservation_date: '2026-10-27',
      reservation_time: '18:30',
      guest_count: 4,
      website_trap: ''
    };
    const result = reservationSchema.safeParse(validData);
    expect(result.success).toBe(true);
  });

  it('harus menolak nama yang terlalu pendek', () => {
    const invalidData = {
      customer_name: 'A',
      customer_phone: '081234567890',
      reservation_date: '2026-10-27',
      reservation_time: '18:30',
      guest_count: 2
    };
    const result = reservationSchema.safeParse(invalidData);
    expect(result.success).toBe(false);
  });

  it('harus menolak jumlah tamu 0 atau negatif', () => {
    const invalidData = {
      customer_name: 'Budi Santoso',
      customer_phone: '081234567890',
      reservation_date: '2026-10-27',
      reservation_time: '18:30',
      guest_count: 0
    };
    const result = reservationSchema.safeParse(invalidData);
    expect(result.success).toBe(false);
  });
});
