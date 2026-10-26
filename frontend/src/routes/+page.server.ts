import type { PageServerLoad } from './$types';
import { fetchApi } from '$lib/api/client';
import type { Category, MenuItem } from '$lib/types';

export const load: PageServerLoad = async ({ fetch }) => {
  try {
    const [categories, menuItems] = await Promise.all([
      fetchApi<Category[]>('/public/menu/categories', {}, fetch).catch(() => []),
      fetchApi<MenuItem[]>('/public/menu', {}, fetch).catch(() => [])
    ]);

    const featuredItems = menuItems.filter(item => item.is_featured || item.is_available).slice(0, 6);

    return {
      categories,
      featuredItems
    };
  } catch (error) {
    return {
      categories: [],
      featuredItems: []
    };
  }
};
