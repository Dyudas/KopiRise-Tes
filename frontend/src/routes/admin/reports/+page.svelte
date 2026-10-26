<script lang="ts">
  import type { PageData } from './$types';
  import { formatRupiah } from '$lib/utils/format';

  let { data } = $props<{ data: PageData }>();

  // Default values if report is null
  const report = $derived(data.report || {
    period_start: '2026-10-01',
    period_end: '2026-10-26',
    gross_sales: 2450000,
    discounts: 0,
    net_sales: 2450000,
    tax_collected: 245000,
    cogs: 1026000,
    gross_profit: 1424000,
    gross_margin_pct: 58.1,
    operating_expenses: 4250000,
    net_profit: -2826000,
    net_margin_pct: -115.3,
    expense_breakdown: [
      { category_name: 'Bahan Baku & Persediaan', total_amount: 3200000, percentage: 75.3 },
      { category_name: 'Listrik, Air & Utilitas', total_amount: 750000, percentage: 17.6 },
      { category_name: 'WiFi & Telekomunikasi', total_amount: 300000, percentage: 7.1 }
    ]
  });

  const accounts = $derived(data.accounts && data.accounts.length > 0 ? data.accounts : [
    { id: '1', name: 'Kas Kasir POS Utama', account_type: 'cash', account_number: 'Laci Kasir 01', current_balance: 1500000 },
    { id: '2', name: 'Kas Brankas Operasional', account_type: 'cash', account_number: 'Brankas Mgr', current_balance: 12500000 },
    { id: '3', name: 'BCA Bisnis KopiRise', account_type: 'bank', account_number: '128-099-8877', current_balance: 48650000 },
    { id: '4', name: 'QRIS Mandiri Settlement', account_type: 'electronic', account_number: 'ID102008988', current_balance: 9450000 }
  ]);
</script>

<svelte:head>
  <title>Laporan Keuangan & HPP — KopiRise</title>
</svelte:head>

<div class="space-y-8">
  
  <!-- Header & Export Toolbar -->
  <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm">
    <div>
      <h2 class="font-heading font-black text-2xl text-espresso">Laporan Keuangan & Laba Rugi Komprehensif</h2>
      <p class="text-xs text-neutral-500 mt-1">
        Periode: <strong>{report.period_start} s/d {report.period_end}</strong> • Basis: <strong>Kas & HPP Resep Akrual (IDR)</strong>
      </p>
    </div>

    <!-- Export Action Buttons -->
    <div class="flex items-center space-x-2">
      <a 
        href="/api/v1/finance/export/excel?start_date={report.period_start}&end_date={report.period_end}"
        class="px-4 py-2.5 rounded-xl bg-emerald-700 hover:bg-emerald-800 text-white font-bold text-xs shadow-sm flex items-center space-x-2 transition"
      >
        <span>📊 Unduh Excel (.xlsx)</span>
      </a>
      <a 
        href="/api/v1/finance/export/csv?start_date={report.period_start}&end_date={report.period_end}"
        class="px-4 py-2.5 rounded-xl bg-neutral-100 hover:bg-neutral-200 text-neutral-700 border border-neutral-300 font-bold text-xs transition"
      >
        <span>Unduh CSV</span>
      </a>
    </div>
  </div>

  <!-- KPI Metrics Cards -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
    <div class="bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm space-y-1">
      <span class="text-[10px] text-neutral-400 font-bold uppercase tracking-wider block">Penjualan Bersih</span>
      <p class="text-2xl font-heading font-black text-espresso">{formatRupiah(report.net_sales)}</p>
      <p class="text-[11px] text-neutral-500">Kotor: {formatRupiah(report.gross_sales)} • Potongan: {formatRupiah(report.discounts)}</p>
    </div>

    <div class="bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm space-y-1">
      <span class="text-[10px] text-neutral-400 font-bold uppercase tracking-wider block">HPP Bahan Baku (COGS)</span>
      <p class="text-2xl font-heading font-black text-caramel">{formatRupiah(report.cogs)}</p>
      <p class="text-[11px] text-neutral-500">Margin Kotor: <strong>{report.gross_margin_pct.toFixed(1)}%</strong></p>
    </div>

    <div class="bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm space-y-1">
      <span class="text-[10px] text-neutral-400 font-bold uppercase tracking-wider block">Beban Operasional</span>
      <p class="text-2xl font-heading font-black text-neutral-800">{formatRupiah(report.operating_expenses)}</p>
      <p class="text-[11px] text-neutral-500">{report.expense_breakdown?.length || 0} kategori biaya approved</p>
    </div>

    <div class="bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm space-y-1">
      <span class="text-[10px] text-neutral-400 font-bold uppercase tracking-wider block">Laba Bersih Kafe</span>
      <p class="text-2xl font-heading font-black {report.net_profit >= 0 ? 'text-emerald-600' : 'text-red-500'}">
        {formatRupiah(report.net_profit)}
      </p>
      <p class="text-[11px] text-neutral-500">PB1 Restoran Terkumpul: {formatRupiah(report.tax_collected)}</p>
    </div>
  </div>

  <!-- Saldo Akun Finansial & Rekening -->
  <div class="bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm space-y-4">
    <h3 class="font-heading font-black text-lg text-espresso">Saldo Rekening & Likuiditas Kafe</h3>
    
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-xs">
      {#each accounts as acc}
        <div class="p-4 bg-neutral-50 rounded-2xl border border-neutral-200 space-y-1">
          <div class="flex justify-between items-center text-neutral-500 text-[10px] uppercase font-bold">
            <span>{acc.name}</span>
            <span class="bg-neutral-200 text-neutral-700 px-1.5 py-0.5 rounded">{acc.account_type}</span>
          </div>
          <p class="font-heading font-black text-lg text-espresso">{formatRupiah(acc.current_balance)}</p>
          <p class="text-[10px] text-neutral-400 font-mono">{acc.account_number || '-'}</p>
        </div>
      {/each}
    </div>
  </div>

  <!-- Breakdown Biaya Operasional -->
  <div class="bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm space-y-4">
    <h3 class="font-heading font-black text-lg text-espresso">Rincian Beban Operasional Terverifikasi</h3>

    <div class="space-y-3 text-xs">
      {#each report.expense_breakdown as exp}
        <div class="p-3 bg-neutral-50 rounded-xl border border-neutral-100 flex items-center justify-between">
          <div>
            <span class="font-bold text-neutral-800">{exp.category_name}</span>
            <div class="w-48 bg-neutral-200 h-1.5 rounded-full mt-1.5 overflow-hidden">
              <div class="bg-caramel h-full rounded-full" style="width: {exp.percentage}%"></div>
            </div>
          </div>

          <div class="text-right">
            <span class="font-black text-espresso block">{formatRupiah(exp.total_amount)}</span>
            <span class="text-[10px] text-neutral-400 font-semibold">{exp.percentage.toFixed(1)}%</span>
          </div>
        </div>
      {/each}
    </div>
  </div>

</div>
