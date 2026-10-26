import type { RequestHandler } from './$types';
import { redirect } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ cookies }) => {
  cookies.delete('kopirise_session', { path: '/' });
  throw redirect(303, '/admin/login');
};

export const POST: RequestHandler = async ({ cookies }) => {
  cookies.delete('kopirise_session', { path: '/' });
  throw redirect(303, '/admin/login');
};
