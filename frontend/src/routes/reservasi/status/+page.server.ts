import type { PageServerLoad } from './$types';
import { fetchApi } from '$lib/api/client';
import type { Reservation } from '$lib/types';

export const load: PageServerLoad = async ({ url, fetch }) => {
  const code = url.searchParams.get('code');

  if (!code) {
    return {
      code: '',
      reservation: null
    };
  }

  try {
    const reservation = await fetchApi<Reservation>(`/public/reservations/${encodeURIComponent(code)}`, {}, fetch);
    return {
      code,
      reservation
    };
  } catch (err: any) {
    return {
      code,
      reservation: null,
      error: err.message || 'Data reservasi tidak ditemukan.'
    };
  }
};
