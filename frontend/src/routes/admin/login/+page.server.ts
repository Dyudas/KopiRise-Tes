import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';

const API_BASE_URL = process.env.KOPIRISE_API_URL || 'http://127.0.0.1:8080/api/v1';

export const load: PageServerLoad = async ({ locals }) => {
  if (locals.user) {
    throw redirect(303, '/admin/pos');
  }
  return {};
};

export const actions: Actions = {
  default: async ({ request, fetch, cookies, url }) => {
    const formData = await request.formData();
    const email = formData.get('email')?.toString().trim() || '';
    const password = formData.get('password')?.toString() || '';
    const redirectUrl = url.searchParams.get('redirect') || '';

    if (!email || !password) {
      return fail(400, {
        error: 'Email dan kata sandi wajib diisi.',
        email
      });
    }

    try {
      const res = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });

      const json = await res.json();

      if (!res.ok || json.error) {
        return fail(400, {
          error: json.error?.message || 'Email atau kata sandi tidak cocok.',
          email
        });
      }

      // Ambil token dari respons atau set cookie
      const token = json.data?.token || json.data?.session_id;
      if (token) {
        cookies.set('kopirise_session', token, {
          path: '/',
          httpOnly: true,
          sameSite: 'lax',
          secure: process.env.NODE_ENV === 'production',
          maxAge: 60 * 60 * 24 * 7 // 7 hari
        });
      }

      const user = json.data?.user;
      if (redirectUrl) {
        throw redirect(303, redirectUrl);
      }

      if (user?.roles?.includes('barista')) {
        throw redirect(303, '/admin/kds');
      } else if (user?.roles?.includes('cashier')) {
        throw redirect(303, '/admin/pos');
      } else {
        throw redirect(303, '/admin/reports');
      }
    } catch (err: any) {
      if (err?.status === 303) throw err;
      return fail(500, {
        error: err.message || 'Gagal menghubungi server autentikasi.',
        email
      });
    }
  }
};
