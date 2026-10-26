import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import { z } from 'zod';
import { fetchApi } from '$lib/api/client';
import type { Area, Reservation } from '$lib/types';

const reservationSchema = z.object({
  customer_name: z.string().min(2, 'Nama pemesan minimal 2 karakter'),
  customer_phone: z.string().min(9, 'Nomor WhatsApp tidak valid (minimal 9 digit)'),
  customer_email: z.string().email('Format email tidak valid').optional().or(z.literal('')),
  reservation_date: z.string().min(10, 'Tanggal kunjungan wajib diisi'),
  reservation_time: z.string().min(5, 'Jam kedatangan wajib dipilih'),
  guest_count: z.coerce.number().min(1, 'Jumlah tamu minimal 1 orang').max(50, 'Maksimal 50 orang per pemesanan'),
  area_id: z.string().optional(),
  notes: z.string().optional(),
  website_trap: z.string().optional() // Honeypot
});

export const load: PageServerLoad = async ({ fetch }) => {
  try {
    const areas = await fetchApi<Area[]>('/tables/areas', {}, fetch).catch(() => [
      { id: 'area-indoor-ac', name: 'Indoor AC Non-Smoking', is_smoking: false },
      { id: 'area-outdoor-garden', name: 'Outdoor Garden & Terrace', is_smoking: true },
      { id: 'area-mezzanine', name: 'Mezzanine Bar (Lt 2)', is_smoking: false },
      { id: 'area-vip-meeting', name: 'VIP Room Semeru', is_smoking: false }
    ]);
    return { areas };
  } catch {
    return { areas: [] };
  }
};

export const actions: Actions = {
  default: async ({ request, fetch }) => {
    const formData = await request.formData();
    const data = Object.fromEntries(formData.entries());

    // Validasi Zod
    const parsed = reservationSchema.safeParse(data);
    if (!parsed.success) {
      const errors = parsed.error.flatten().fieldErrors;
      return fail(400, {
        errors,
        values: data
      });
    }

    const { website_trap, ...payload } = parsed.data;

    // Honeypot check di server
    if (website_trap && website_trap.trim() !== '') {
      return fail(400, {
        generalError: 'Spam terdeteksi',
        values: data
      });
    }

    try {
      const reservation = await fetchApi<Reservation>('/public/reservations', {
        method: 'POST',
        body: JSON.stringify(payload)
      }, fetch);

      return {
        success: true,
        reservation
      };
    } catch (err: any) {
      return fail(400, {
        generalError: err.message || 'Gagal membuat reservasi meja. Silakan coba kembali.',
        values: data
      });
    }
  }
};
