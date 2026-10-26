import type { PageServerLoad, Actions } from './$types';
import { fetchApi } from '$lib/api/client';
import { fail } from '@sveltejs/kit';
import type { Category, MenuItem } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
  const sessionToken = cookies.get('kopirise_session');
  const headers = { 'Cookie': `kopirise_session=${sessionToken}` };

  try {
    const [categories, items] = await Promise.all([
      fetchApi<Category[]>('/public/menu/categories', {}, fetch).catch(() => []),
      fetchApi<MenuItem[]>('/public/menu', {}, fetch).catch(() => [])
    ]);

    return { categories, items };
  } catch {
    return { categories: [], items: [] };
  }
};

export const actions: Actions = {
  toggleAvailability: async ({ request, fetch, cookies }) => {
    const formData = await request.formData();
    const itemId = formData.get('item_id')?.toString() || '';
    const sessionToken = cookies.get('kopirise_session');

    try {
      await fetchApi(`/menu/items/${itemId}/availability`, {
        method: 'PATCH',
        headers: { 'Cookie': `kopirise_session=${sessionToken}` }
      }, fetch);

      return { success: true };
    } catch (err: any) {
      return fail(400, { error: err.message || 'Gagal mengubah ketersediaan menu' });
    }
  }
};
