import type { PageServerLoad, Actions } from './$types';
import { fetchApi } from '$lib/api/client';
import { fail } from '@sveltejs/kit';
import type { Category, MenuItem, Table } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
  const sessionToken = cookies.get('kopirise_session');
  const headers = { 'Cookie': `kopirise_session=${sessionToken}` };

  try {
    const [categories, items, tables, shift] = await Promise.all([
      fetchApi<Category[]>('/public/menu/categories', {}, fetch).catch(() => []),
      fetchApi<MenuItem[]>('/public/menu', {}, fetch).catch(() => []),
      fetchApi<Table[]>('/tables', { headers }, fetch).catch(() => []),
      fetchApi<any>('/cash-shifts/current', { headers }, fetch).catch(() => null)
    ]);

    return {
      categories,
      items,
      tables,
      currentShift: shift
    };
  } catch (error) {
    return {
      categories: [],
      items: [],
      tables: [],
      currentShift: null
    };
  }
};

export const actions: Actions = {
  openShift: async ({ request, fetch, cookies }) => {
    const formData = await request.formData();
    const initialCash = parseInt(formData.get('initial_cash')?.toString() || '0', 10);
    const sessionToken = cookies.get('kopirise_session');

    try {
      const shift = await fetchApi('/cash-shifts/open', {
        method: 'POST',
        headers: { 'Cookie': `kopirise_session=${sessionToken}` },
        body: JSON.stringify({ initial_cash: initialCash })
      }, fetch);

      return { success: true, shift };
    } catch (err: any) {
      return fail(400, { error: err.message || 'Gagal membuka shift kasir' });
    }
  },

  closeShift: async ({ request, fetch, cookies }) => {
    const formData = await request.formData();
    const actualCash = parseInt(formData.get('actual_cash')?.toString() || '0', 10);
    const notes = formData.get('closing_notes')?.toString() || '';
    const sessionToken = cookies.get('kopirise_session');

    try {
      const shift = await fetchApi('/cash-shifts/close', {
        method: 'POST',
        headers: { 'Cookie': `kopirise_session=${sessionToken}` },
        body: JSON.stringify({ actual_cash: actualCash, closing_notes: notes })
      }, fetch);

      return { success: true, closedShift: shift };
    } catch (err: any) {
      return fail(400, { error: err.message || 'Gagal menutup shift kasir' });
    }
  }
};
