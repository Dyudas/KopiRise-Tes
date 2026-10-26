import type { ApiResponse } from '$lib/types';

const API_BASE_URL = typeof window === 'undefined'
  ? (process.env.KOPIRISE_API_URL || 'http://127.0.0.1:8080/api/v1')
  : '/api/v1';

export async function fetchApi<T>(
  path: string,
  options: RequestInit = {},
  customFetch: typeof fetch = fetch
): Promise<T> {
  const url = path.startsWith('http') ? path : `${API_BASE_URL}${path.startsWith('/') ? path : `/${path}`}`;

  const headers = new Headers(options.headers);
  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await customFetch(url, {
    ...options,
    headers
  });

  const json: ApiResponse<T> = await response.json();

  if (!response.ok || json.error) {
    const errorMsg = json.error?.message || `Request failed with status ${response.status}`;
    throw new Error(errorMsg);
  }

  return json.data as T;
}
