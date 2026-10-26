import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { fetchApi } from '$lib/api/client';
import type { Table, Category, MenuItem } from '$lib/types';

export const load: PageServerLoad = async ({ params, fetch }) => {
  const { token } = params;

  try {
    // 1. Resolve QR token meja
    const table = await fetchApi<Table>(`/public/qr/${encodeURIComponent(token)}`, {}, fetch);

    // 2. Fetch menu categories and items for direct ordering
    const [categories, items] = await Promise.all([
      fetchApi<Category[]>('/public/menu/categories', {}, fetch).catch(() => []),
      fetchApi<MenuItem[]>('/public/menu', {}, fetch).catch(() => [])
    ]);

    return {
      token,
      table,
      categories,
      items
    };
  } catch (err: any) {
    throw error(404, {
      message: 'Kode QR Meja tidak valid atau telah diperbarui oleh kasir. Silakan tanyakan kepada staf kafe.'
    });
  }
};
