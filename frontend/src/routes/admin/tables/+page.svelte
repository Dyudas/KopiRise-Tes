<script lang="ts">
  import type { PageData } from './$types';
  import type { Table } from '$lib/types';

  let { data } = $props<{ data: PageData }>();

  // Fallback demo tables if backend tables are empty
  const defaultTables = [
    { id: 't-1', table_number: 'IN-01', area_name: 'Indoor AC Non-Smoking', capacity: 2, status: 'available', qr_token: 'kr_tbl_9a7d3f82e1c440a' },
    { id: 't-2', table_number: 'IN-02', area_name: 'Indoor AC Non-Smoking', capacity: 2, status: 'available', qr_token: 'kr_tbl_4b1c8e73d9a541b' },
    { id: 't-3', table_number: 'IN-03', area_name: 'Indoor AC Non-Smoking', capacity: 4, status: 'occupied', qr_token: 'kr_tbl_7e2a9d61c8f442c' },
    { id: 't-4', table_number: 'IN-04', area_name: 'Indoor AC Non-Smoking', capacity: 4, status: 'available', qr_token: 'kr_tbl_3c8f1e94b7a243d' },
    { id: 't-5', table_number: 'OUT-01', area_name: 'Outdoor Garden Terrace', capacity: 2, status: 'available', qr_token: 'kr_tbl_5d8a3c21f9e448c' },
    { id: 't-6', table_number: 'OUT-04', area_name: 'Outdoor Garden Terrace', capacity: 4, status: 'occupied', qr_token: 'kr_tbl_8e3a4d71b9c251f' },
    { id: 't-7', table_number: 'MEZ-01', area_name: 'Mezzanine Bar (Lt 2)', capacity: 2, status: 'available', qr_token: 'kr_tbl_7b2e9c84a1d358a' },
    { id: 't-8', table_number: 'VIP-01', area_name: 'VIP Room Semeru', capacity: 10, status: 'reserved', qr_token: 'kr_tbl_5e9a8b14c7d362e' }
  ];

  let tables = $state(data.tables && data.tables.length > 0 ? data.tables : defaultTables);
  let activeModalTable = $state<any | null>(null);

  function printTentCard() {
    window.print();
  }
</script>

<svelte:head>
  <title>Denah Meja & Kartu QR — KopiRise</title>
</svelte:head>

<div class="space-y-6">
  
  <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm">
    <div>
      <h2 class="font-heading font-black text-2xl text-espresso">Manajemen Denah Meja & QR Token</h2>
      <p class="text-xs text-neutral-500 mt-1">
        Monitor okupansi 24 meja kafe dan cetak kartu QR Tent Card siap pajang.
      </p>
    </div>

    <!-- Quick Stats -->
    <div class="flex items-center space-x-3 text-xs">
      <span class="px-3 py-1.5 rounded-xl bg-emerald-100 text-emerald-800 font-bold">
        Tersedia: {tables.filter(t => t.status === 'available').length}
      </span>
      <span class="px-3 py-1.5 rounded-xl bg-amber-100 text-amber-800 font-bold">
        Terisi: {tables.filter(t => t.status === 'occupied').length}
      </span>
      <span class="px-3 py-1.5 rounded-xl bg-purple-100 text-purple-800 font-bold">
        Reservasi: {tables.filter(t => t.status === 'reserved').length}
      </span>
    </div>
  </div>

  <!-- Table Cards Grid -->
  <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-6">
    {#each tables as table}
      <div class="bg-white rounded-3xl border border-neutral-200 p-5 shadow-sm space-y-4 flex flex-col justify-between hover:border-caramel transition">
        <div>
          <div class="flex justify-between items-center">
            <span class="font-heading font-black text-2xl text-espresso">{table.table_number}</span>
            <span class="text-[10px] font-extrabold uppercase px-2.5 py-1 rounded-full {
              table.status === 'occupied' ? 'bg-amber-100 text-amber-800 border border-amber-300' :
              table.status === 'reserved' ? 'bg-purple-100 text-purple-800 border border-purple-300' :
              'bg-emerald-100 text-emerald-800 border border-emerald-300'
            }">
              {table.status}
            </span>
          </div>

          <p class="text-xs text-neutral-500 font-medium mt-1">{table.area_name || 'Area Kafe'}</p>
          <p class="text-xs text-neutral-700 font-bold mt-1">Kapasitas: {table.capacity} Tamu</p>
          <p class="text-[9px] font-mono text-neutral-400 mt-2 truncate">Token: {table.qr_token || 'kr_tbl_active'}</p>
        </div>

        <button 
          onclick={() => activeModalTable = table}
          class="w-full py-2.5 rounded-xl bg-espresso hover:bg-caramel text-creamWarm font-bold text-xs shadow-sm transition flex items-center justify-center space-x-1.5"
        >
          <span>📱 Lihat & Cetak QR Tent Card</span>
        </button>
      </div>
    {/each}
  </div>

</div>

<!-- TENT CARD PRINT MODAL -->
{#if activeModalTable}
  <div class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="bg-white rounded-3xl max-w-sm w-full shadow-2xl p-6 space-y-6">
      
      <!-- Tent Card Print Template -->
      <div id="tent-card-print-area" class="border-4 border-espresso rounded-3xl p-6 text-center space-y-4 bg-creamWarm">
        <div class="space-y-1">
          <span class="text-[10px] font-black tracking-widest uppercase text-caramel">KopiRise Cafe & Resto</span>
          <h2 class="font-heading font-black text-3xl text-espresso leading-none">
            MEJA {activeModalTable.table_number}
          </h2>
          <p class="text-[10px] text-neutral-500">{activeModalTable.area_name || 'Area Kafe'}</p>
        </div>

        <!-- Simulated Clean QR Code Graphic -->
        <div class="w-44 h-44 mx-auto bg-white p-3 rounded-2xl shadow-inner border border-neutral-300 flex flex-col items-center justify-center space-y-1">
          <div class="w-32 h-32 bg-espresso p-1.5 rounded-xl flex items-center justify-center text-creamWarm text-[10px] font-mono text-center">
            [ KOPIRISE QR ]<br/>
            {activeModalTable.table_number}<br/>
            SCAN TO ORDER
          </div>
        </div>

        <div class="space-y-1 text-xs">
          <p class="font-heading font-extrabold text-espresso">Scan untuk Pesan dari Meja</p>
          <p class="text-[10px] text-neutral-500 leading-tight">
            Arahkan kamera smartphone Anda ke kode QR di atas untuk melihat menu & memesan tanpa perlu ke kasir.
          </p>
        </div>

        <div class="pt-2 border-t border-dashed border-neutral-300 text-[9px] text-neutral-500 font-mono">
          WiFi: KopiRise_Guest / semerujember
        </div>
      </div>

      <div class="flex gap-2">
        <button 
          onclick={printTentCard}
          class="flex-1 py-3 bg-caramel hover:bg-caramelHover text-white rounded-xl font-bold text-xs shadow"
        >
          🖨️ Cetak Kartu Laminasi
        </button>
        <button 
          onclick={() => activeModalTable = null}
          class="px-4 py-3 bg-neutral-100 hover:bg-neutral-200 text-neutral-700 rounded-xl font-bold text-xs"
        >
          Tutup
        </button>
      </div>

    </div>
  </div>
{/if}
