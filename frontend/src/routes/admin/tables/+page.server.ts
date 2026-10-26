import type { PageServerLoad } from './$types';
import { fetchApi } from '$lib/api/client';
import type { Table, Area } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
  const sessionToken = cookies.get('kopirise_session');
  const headers = { 'Cookie': `kopirise_session=${sessionToken}` };

  try {
    const [tables, areas] = await Promise.all([
      fetchApi<Table[]>('/tables', { headers }, fetch).catch(() => []),
      fetchApi<Area[]>('/tables/areas', { headers }, fetch).catch(() => [])
    ]);

    return { tables, areas };
  } catch {
    return { tables: [], areas: [] };
  }
};
