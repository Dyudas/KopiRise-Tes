import type { PageServerLoad } from './$types';
import { fetchApi } from '$lib/api/client';
import type { Category, MenuItem } from '$lib/types';

export const load: PageServerLoad = async ({ fetch }) => {
  try {
    const [categories, items] = await Promise.all([
      fetchApi<Category[]>('/public/menu/categories', {}, fetch).catch(() => []),
      fetchApi<MenuItem[]>('/public/menu', {}, fetch).catch(() => [])
    ]);

    return {
      categories,
      items
    };
  } catch (error) {
    return {
      categories: [],
      items: []
    };
  }
};
