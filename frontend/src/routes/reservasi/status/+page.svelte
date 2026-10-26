<script lang="ts">
  import type { PageData } from './$types';
  import { formatDateIndo } from '$lib/utils/format';

  let { data } = $props<{ data: PageData }>();

  let inputCode = $state(data.code || '');
</script>

<svelte:head>
  <title>Cek Status Reservasi — KopiRise Jember</title>
  <meta name="description" content="Lacak status reservasi meja KopiRise Jember secara real-time menggunakan nomor kode reservasi Anda." />
</svelte:head>

<div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-12 space-y-8">
  
  <div class="text-center space-y-2">
    <span class="text-xs font-bold text-caramel uppercase tracking-widest">Pelacakan Reservasi</span>
    <h1 class="font-heading font-black text-3xl text-espresso">Cek Status Reservasi Meja</h1>
    <p class="text-xs sm:text-sm text-neutral-500">
      Masukkan nomor reservasi unik yang Anda terima (misal: <code>RSV-20261026-0001</code>)
    </p>
  </div>

  <!-- Search Box -->
  <form method="GET" class="bg-white p-4 rounded-3xl border border-neutral-200 shadow-sm flex gap-3">
    <input 
      type="text" 
      name="code"
      bind:value={inputCode}
      placeholder="Masukkan Nomor Reservasi..."
      required
      class="flex-1 px-4 py-2.5 rounded-2xl border border-neutral-200 text-xs font-mono font-bold focus:ring-2 focus:ring-caramel focus:outline-none uppercase"
    />
    <button 
      type="submit"
      class="px-6 py-2.5 rounded-2xl bg-espresso hover:bg-caramel text-creamWarm font-bold text-xs transition shadow"
    >
      Periksa
    </button>
  </form>

  <!-- Result Container -->
  {#if data.reservation}
    <div class="bg-white rounded-3xl border border-neutral-200 p-8 shadow-sm space-y-6">
      
      <div class="flex items-center justify-between pb-6 border-b border-neutral-100">
        <div>
          <span class="text-[10px] text-neutral-400 font-bold block uppercase tracking-wider">Nomor Pemesanan</span>
          <span class="font-mono font-black text-xl text-espresso">{data.reservation.reservation_number}</span>
        </div>

        <span class="px-3 py-1 rounded-full text-xs font-extrabold uppercase tracking-wide {
          data.reservation.status === 'confirmed' ? 'bg-emerald-100 text-emerald-800 border border-emerald-300' :
          data.reservation.status === 'seated' ? 'bg-blue-100 text-blue-800 border border-blue-300' :
          data.reservation.status === 'completed' ? 'bg-neutral-100 text-neutral-800 border border-neutral-300' :
          data.reservation.status === 'cancelled' ? 'bg-red-100 text-red-800 border border-red-300' :
          'bg-amber-100 text-amber-800 border border-amber-300 animate-pulse'
        }">
          {data.reservation.status}
        </span>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
        <div>
          <span class="text-neutral-400 font-medium block">Nama Pemesan</span>
          <p class="font-bold text-neutral-800 text-sm mt-0.5">{data.reservation.customer_name}</p>
        </div>

        <div>
          <span class="text-neutral-400 font-medium block">Nomor WhatsApp</span>
          <p class="font-bold text-neutral-800 text-sm mt-0.5">{data.reservation.customer_phone}</p>
        </div>

        <div>
          <span class="text-neutral-400 font-medium block">Tanggal & Waktu</span>
          <p class="font-bold text-neutral-800 text-sm mt-0.5">
            {formatDateIndo(data.reservation.reservation_date)}, {data.reservation.reservation_time} WIB
          </p>
        </div>

        <div>
          <span class="text-neutral-400 font-medium block">Kapasitas Tamu</span>
          <p class="font-bold text-neutral-800 text-sm mt-0.5">{data.reservation.guest_count} Orang</p>
        </div>

        <div>
          <span class="text-neutral-400 font-medium block">Alokasi Meja / Area</span>
          <p class="font-bold text-caramel text-sm mt-0.5">
            {data.reservation.table_number ? `Meja ${data.reservation.table_number}` : 'Akan Ditentukan Petugas'}
            {data.reservation.area_name ? `(${data.reservation.area_name})` : ''}
          </p>
        </div>

        <div>
          <span class="text-neutral-400 font-medium block">Durasi Penggunaan</span>
          <p class="font-bold text-neutral-800 text-sm mt-0.5">{data.reservation.duration_minutes} Menit</p>
        </div>
      </div>

      {#if data.reservation.notes}
        <div class="p-4 bg-neutral-50 rounded-2xl border border-neutral-100 text-xs">
          <span class="text-neutral-400 font-medium block mb-1">Catatan Khusus:</span>
          <p class="text-neutral-700 italic">{data.reservation.notes}</p>
        </div>
      {/if}

      <div class="pt-4 border-t border-neutral-100 flex justify-between items-center text-xs">
        <a 
          href="https://wa.me/6281234567890?text={encodeURIComponent(`Halo KopiRise, konfirmasi status reservasi #${data.reservation.reservation_number}.`)}"
          target="_blank"
          rel="noopener noreferrer"
          class="text-caramel hover:underline font-bold"
        >
          Hubungi Petugas Kafe via WhatsApp →
        </a>

        <a href="/reservasi" class="text-neutral-500 hover:text-neutral-800">
          Buat Reservasi Baru
        </a>
      </div>

    </div>
  {:else if data.code}
    <div class="p-8 text-center bg-white rounded-3xl border border-neutral-200 space-y-3">
      <div class="w-12 h-12 rounded-2xl bg-red-100 text-red-600 flex items-center justify-center mx-auto text-xl font-black">
        ✕
      </div>
      <h3 class="font-heading font-extrabold text-base text-espresso">Nomor Reservasi Tidak Ditemukan</h3>
      <p class="text-xs text-neutral-500 max-w-sm mx-auto">
        Tidak ada data reservasi dengan kode "{data.code}". Pastikan penulisan kode sudah tepat sesuai bukti konfirmasi.
      </p>
    </div>
  {/if}

</div>
