<script lang="ts">
  import { enhance } from '$app/forms';
  import type { PageData, ActionData } from './$types';

  let { data, form } = $props<{ data: PageData; form: ActionData }>();

  let isSubmitting = $state(false);

  // Default tanggal besok
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  const minDate = tomorrow.toISOString().split('T')[0];
</script>

<svelte:head>
  <title>Reservasi Meja — KopiRise Cafe & Resto Jember</title>
  <meta name="description" content="Formulir reservasi meja online KopiRise Jember. Pilih area Indoor AC, Outdoor Garden, atau VIP Room Semeru. Konfirmasi cepat via WhatsApp." />
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12 space-y-8">
  
  <div class="text-center space-y-2">
    <span class="text-xs font-bold text-caramel uppercase tracking-widest">Layanan Reservasi</span>
    <h1 class="font-heading font-black text-3xl sm:text-4xl text-espresso">Reservasi Meja KopiRise</h1>
    <p class="text-xs sm:text-sm text-neutral-500 max-w-lg mx-auto">
      Amankan meja Anda untuk pertemuan santai, rapat kerja, atau perayaan keluarga tanpa antre.
    </p>
  </div>

  {#if form?.success && form.reservation}
    <!-- SUCCESS CONFIRMATION BOX -->
    <div class="bg-white rounded-3xl border border-emerald-200 p-8 shadow-xl space-y-6 text-center animate-in fade-in duration-300">
      <div class="w-16 h-16 rounded-3xl bg-emerald-100 text-emerald-700 flex items-center justify-center font-black text-2xl mx-auto">
        ✓
      </div>

      <div class="space-y-2">
        <span class="text-xs font-bold text-emerald-700 uppercase tracking-wider">Reservasi Berhasil Diajukan!</span>
        <h2 class="font-heading font-black text-2xl text-espresso">
          Nomor: <span class="font-mono text-caramel">{form.reservation.reservation_number}</span>
        </h2>
        <p class="text-xs text-neutral-600 max-w-md mx-auto">
          Terima kasih, <strong>{form.reservation.customer_name}</strong>. Permintaan reservasi Anda untuk <strong>{form.reservation.guest_count} orang</strong> pada <strong>{form.reservation.reservation_date} pukul {form.reservation.reservation_time} WIB</strong> telah tersimpan dalam sistem kami.
        </p>
      </div>

      <div class="pt-4 flex flex-col sm:flex-row items-center justify-center gap-3">
        <a 
          href="https://wa.me/6281234567890?text={encodeURIComponent(`Halo KopiRise Jember, saya ingin konfirmasi reservasi #${form.reservation.reservation_number} atas nama ${form.reservation.customer_name}.`)}"
          target="_blank"
          rel="noopener noreferrer"
          class="w-full sm:w-auto px-6 py-3.5 rounded-2xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs shadow-md transition flex items-center justify-center space-x-2"
        >
          <span>Konfirmasi ke WhatsApp Staff</span>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
          </svg>
        </a>

        <a 
          href="/reservasi/status?code={form.reservation.reservation_number}"
          class="w-full sm:w-auto px-6 py-3.5 rounded-2xl bg-neutral-100 hover:bg-neutral-200 text-neutral-700 font-bold text-xs transition"
        >
          Pantau Status Reservasi
        </a>
      </div>
    </div>
  {:else}
    <!-- RESERVATION FORM -->
    <div class="bg-white rounded-3xl border border-neutral-200/90 shadow-sm p-6 sm:p-10 space-y-8">
      
      {#if form?.generalError}
        <div class="p-4 rounded-2xl bg-red-50 border border-red-200 text-red-700 text-xs font-semibold flex items-center space-x-2">
          <span>⚠️ {form.generalError}</span>
        </div>
      {/if}

      <form 
        method="POST" 
        use:enhance={() => {
          isSubmitting = true;
          return async ({ update }) => {
            await update();
            isSubmitting = false;
          };
        }} 
        class="space-y-6"
      >
        <!-- Honeypot Field (Anti-Bot Hidden) -->
        <input 
          type="text" 
          name="website_trap" 
          tabindex="-1" 
          autocomplete="off" 
          class="hidden" 
          aria-hidden="true" 
        />

        <!-- Biodata Pemesan -->
        <div class="space-y-4">
          <h3 class="font-heading font-extrabold text-sm text-espresso uppercase tracking-wider">
            1. Informasi Pemesan
          </h3>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <label for="customer_name" class="font-bold text-xs text-neutral-700 block">
                Nama Lengkap <span class="text-red-500">*</span>
              </label>
              <input 
                id="customer_name"
                name="customer_name" 
                type="text" 
                required 
                placeholder="Contoh: Budi Santoso"
                value={form?.values?.customer_name || ''}
                class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
              />
              {#if form?.errors?.customer_name}
                <p class="text-[10px] text-red-500 font-medium">{form.errors.customer_name[0]}</p>
              {/if}
            </div>

            <div class="space-y-1.5">
              <label for="customer_phone" class="font-bold text-xs text-neutral-700 block">
                Nomor WhatsApp <span class="text-red-500">*</span>
              </label>
              <input 
                id="customer_phone"
                name="customer_phone" 
                type="tel" 
                required 
                placeholder="0812-3456-7890"
                value={form?.values?.customer_phone || ''}
                class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
              />
              {#if form?.errors?.customer_phone}
                <p class="text-[10px] text-red-500 font-medium">{form.errors.customer_phone[0]}</p>
              {/if}
            </div>

            <div class="space-y-1.5 sm:col-span-2">
              <label for="customer_email" class="font-bold text-xs text-neutral-700 block">
                Alamat Email (Opsional untuk bukti pemesanan)
              </label>
              <input 
                id="customer_email"
                name="customer_email" 
                type="email" 
                placeholder="nama@domain.com"
                value={form?.values?.customer_email || ''}
                class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
              />
            </div>
          </div>
        </div>

        <!-- Jadwal Kunjungan & Tamu -->
        <div class="space-y-4 pt-4 border-t border-neutral-100">
          <h3 class="font-heading font-extrabold text-sm text-espresso uppercase tracking-wider">
            2. Jadwal & Jumlah Tamu
          </h3>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div class="space-y-1.5">
              <label for="reservation_date" class="font-bold text-xs text-neutral-700 block">
                Tanggal Kunjungan <span class="text-red-500">*</span>
              </label>
              <input 
                id="reservation_date"
                name="reservation_date" 
                type="date" 
                required 
                min={minDate}
                value={form?.values?.reservation_date || ''}
                class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none bg-neutral-50"
              />
              {#if form?.errors?.reservation_date}
                <p class="text-[10px] text-red-500 font-medium">{form.errors.reservation_date[0]}</p>
              {/if}
            </div>

            <div class="space-y-1.5">
              <label for="reservation_time" class="font-bold text-xs text-neutral-700 block">
                Waktu Kedatangan <span class="text-red-500">*</span>
              </label>
              <select 
                id="reservation_time"
                name="reservation_time" 
                required 
                class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none bg-neutral-50"
              >
                <option value="">Pilih Jam</option>
                <option value="09:00">09:00 WIB</option>
                <option value="10:00">10:00 WIB</option>
                <option value="11:30">11:30 WIB (Makan Siang)</option>
                <option value="13:00">13:00 WIB</option>
                <option value="14:30">14:30 WIB</option>
                <option value="16:00">16:00 WIB (Ngopi Sore)</option>
                <option value="18:30">18:30 WIB (Makan Malam)</option>
                <option value="19:30">19:30 WIB</option>
                <option value="20:30">20:30 WIB</option>
              </select>
              {#if form?.errors?.reservation_time}
                <p class="text-[10px] text-red-500 font-medium">{form.errors.reservation_time[0]}</p>
              {/if}
            </div>

            <div class="space-y-1.5">
              <label for="guest_count" class="font-bold text-xs text-neutral-700 block">
                Jumlah Tamu <span class="text-red-500">*</span>
              </label>
              <input 
                id="guest_count"
                name="guest_count" 
                type="number" 
                required 
                min="1" 
                max="50" 
                placeholder="Contoh: 4"
                value={form?.values?.guest_count || '2'}
                class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none bg-neutral-50"
              />
              {#if form?.errors?.guest_count}
                <p class="text-[10px] text-red-500 font-medium">{form.errors.guest_count[0]}</p>
              {/if}
            </div>
          </div>
        </div>

        <!-- Pilihan Area & Catatan -->
        <div class="space-y-4 pt-4 border-t border-neutral-100">
          <h3 class="font-heading font-extrabold text-sm text-espresso uppercase tracking-wider">
            3. Pilihan Area & Catatan Khusus
          </h3>

          <div class="space-y-1.5">
            <label for="area_id" class="font-bold text-xs text-neutral-700 block">
              Preferensi Area Kafe
            </label>
            <select 
              id="area_id"
              name="area_id"
              class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none bg-neutral-50"
            >
              <option value="">Bebas (Pilihan Terbaik Sistem)</option>
              {#each data.areas as area}
                <option value={area.id}>{area.name} {area.is_smoking ? '(Smoking Allowed)' : '(Non-Smoking)'}</option>
              {/each}
            </select>
          </div>

          <div class="space-y-1.5">
            <label for="notes" class="font-bold text-xs text-neutral-700 block">
              Catatan Khusus (Kebutuhan kursi bayi, dekorasi ulang tahun, proyektor/TV rapat)
            </label>
            <textarea 
              id="notes"
              name="notes" 
              rows="3" 
              placeholder="Tuliskan catatan khusus Anda jika ada..."
              class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
            >{form?.values?.notes || ''}</textarea>
          </div>
        </div>

        <div class="pt-4">
          <button 
            type="submit" 
            disabled={isSubmitting}
            class="w-full py-4 rounded-2xl bg-caramel hover:bg-caramelHover text-white font-heading font-extrabold text-sm shadow-xl transition transform active:scale-95 flex items-center justify-center space-x-2 disabled:opacity-50"
          >
            {#if isSubmitting}
              <span class="animate-spin text-base">⟳</span>
              <span>Memproses Reservasi...</span>
            {:else}
              <span>Ajukan Reservasi Meja</span>
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
              </svg>
            {/if}
          </button>
        </div>

      </form>

    </div>
  {/if}

</div>
