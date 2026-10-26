import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

const API_BASE_URL = process.env.KOPIRISE_API_URL || 'http://127.0.0.1:8080/api/v1';

export const handle: Handle = async ({ event, resolve }) => {
  const sessionToken = event.cookies.get('kopirise_session');

  // 1. Verifikasi Session Token jika ada
  if (sessionToken) {
    try {
      const res = await event.fetch(`${API_BASE_URL}/auth/me`, {
        headers: {
          'Cookie': `kopirise_session=${sessionToken}`
        }
      });

      if (res.ok) {
        const json = await res.json();
        if (json.data) {
          event.locals.user = json.data;
        }
      } else {
        event.cookies.delete('kopirise_session', { path: '/' });
      }
    } catch {
      // Backend mungkin sedang cold-start atau offline
    }
  }

  // 2. Proteksi Rute Admin (/admin/*)
  const path = event.url.pathname;
  if (path.startsWith('/admin')) {
    const isLoginPage = path === '/admin/login';

    if (!event.locals.user && !isLoginPage) {
      throw redirect(303, `/admin/login?redirect=${encodeURIComponent(path)}`);
    }

    if (event.locals.user && isLoginPage) {
      // Pengguna sudah login, arahkan ke halaman sesuai role
      if (event.locals.user.roles.includes('barista')) {
        throw redirect(303, '/admin/kds');
      } else if (event.locals.user.roles.includes('cashier')) {
        throw redirect(303, '/admin/pos');
      } else {
        throw redirect(303, '/admin/reports');
      }
    }

    // Role-based restrictions
    if (event.locals.user && !isLoginPage) {
      const roles = event.locals.user.roles;
      const isAdminOrManager = roles.includes('admin') || roles.includes('manager');

      if (path.startsWith('/admin/reports') && !isAdminOrManager) {
        throw redirect(303, '/admin/pos');
      }
      if (path.startsWith('/admin/menu') && !isAdminOrManager) {
        throw redirect(303, '/admin/pos');
      }
    }
  }

  return resolve(event);
};
