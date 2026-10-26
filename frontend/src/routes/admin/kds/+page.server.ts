import type { PageServerLoad, Actions } from './$types';
import { fetchApi } from '$lib/api/client';
import { fail } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
  const sessionToken = cookies.get('kopirise_session');
  const headers = { 'Cookie': `kopirise_session=${sessionToken}` };

  try {
    const orders = await fetchApi<any[]>('/orders/kitchen', { headers }, fetch).catch(() => []);
    return { orders };
  } catch {
    return { orders: [] };
  }
};

export const actions: Actions = {
  updateOrderStatus: async ({ request, fetch, cookies }) => {
    const formData = await request.formData();
    const orderId = formData.get('order_id')?.toString() || '';
    const status = formData.get('status')?.toString() || '';
    const sessionToken = cookies.get('kopirise_session');

    try {
      await fetchApi(`/orders/${orderId}/status`, {
        method: 'PATCH',
        headers: { 'Cookie': `kopirise_session=${sessionToken}` },
        body: JSON.stringify({ status })
      }, fetch);

      return { success: true };
    } catch (err: any) {
      return fail(400, { error: err.message || 'Gagal mengubah status pesanan' });
    }
  },

  updateItemStatus: async ({ request, fetch, cookies }) => {
    const formData = await request.formData();
    const itemId = formData.get('item_id')?.toString() || '';
    const status = formData.get('status')?.toString() || '';
    const sessionToken = cookies.get('kopirise_session');

    try {
      await fetchApi(`/orders/items/${itemId}/status`, {
        method: 'PATCH',
        headers: { 'Cookie': `kopirise_session=${sessionToken}` },
        body: JSON.stringify({ status })
      }, fetch);

      return { success: true };
    } catch (err: any) {
      return fail(400, { error: err.message || 'Gagal mengubah status item' });
    }
  }
};
