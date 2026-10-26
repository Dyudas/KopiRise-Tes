import type { PageServerLoad } from './$types';
import { fetchApi } from '$lib/api/client';
import type { ProfitLossReport, FinancialAccount, Ingredient } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
  const sessionToken = cookies.get('kopirise_session');
  const headers = { 'Cookie': `kopirise_session=${sessionToken}` };

  try {
    const [report, accounts, lowStock] = await Promise.all([
      fetchApi<ProfitLossReport>('/finance/report/profit-loss', { headers }, fetch).catch(() => null),
      fetchApi<FinancialAccount[]>('/finance/accounts', { headers }, fetch).catch(() => []),
      fetchApi<Ingredient[]>('/inventory/ingredients/low-stock', { headers }, fetch).catch(() => [])
    ]);

    return {
      report,
      accounts,
      lowStock
    };
  } catch (error) {
    return {
      report: null,
      accounts: [],
      lowStock: []
    };
  }
};
