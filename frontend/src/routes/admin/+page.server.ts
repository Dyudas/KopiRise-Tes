import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ locals }) => {
  if (!locals.user) {
    throw redirect(303, '/admin/login');
  }

  if (locals.user.roles.includes('barista')) {
    throw redirect(303, '/admin/kds');
  } else if (locals.user.roles.includes('cashier')) {
    throw redirect(303, '/admin/pos');
  } else {
    throw redirect(303, '/admin/reports');
  }
};
