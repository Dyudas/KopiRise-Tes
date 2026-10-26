<script lang="ts">
  import { enhance } from '$app/forms';
  import type { ActionData } from './$types';

  let { form } = $props<{ form: ActionData }>();

  let email = $state(form?.email || '');
  let password = $state('');
  let isLoggingIn = $state(false);

  function fillDemo(e: string, p: string) {
    email = e;
    password = p;
  }
</script>

<svelte:head>
  <title>Masuk Sistem Karyawan — KopiRise Jember</title>
</svelte:head>

<div class="min-h-screen bg-espresso flex flex-col justify-center py-12 sm:px-6 lg:px-8">
  <div class="sm:mx-auto sm:w-full sm:max-w-md text-center space-y-3">
    <div class="w-14 h-14 rounded-3xl bg-caramel mx-auto flex items-center justify-center text-white shadow-xl">
      <svg class="w-8 h-8 fill-current" viewBox="0 0 24 24">
        <path d="M2,21H20V19H2M20,8H18V5H20M20,3H4V13A4,4 0 0,0 8,17H14A4,4 0 0,0 18,13V10H20A2,2 0 0,0 22,8V5C22,3.89 21.1,3 20,3Z" />
      </svg>
    </div>
    <h2 class="font-heading font-black text-3xl text-creamWarm">KopiRise Staff Portal</h2>
    <p class="text-xs text-creamWarm/70">
      Sistem Terpadu POS Kasir, Barista KDS, Manajemen Meja, & Laporan Keuangan
    </p>
  </div>

  <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md px-4">
    <div class="bg-white py-8 px-6 sm:px-10 rounded-3xl shadow-2xl border border-neutral-100 space-y-6">
      
      {#if form?.error}
        <div class="p-3.5 rounded-2xl bg-red-50 border border-red-200 text-red-700 text-xs font-semibold flex items-center space-x-2">
          <span>⚠️ {form.error}</span>
        </div>
      {/if}

      <form 
        method="POST" 
        use:enhance={() => {
          isLoggingIn = true;
          return async ({ update }) => {
            await update();
            isLoggingIn = false;
          };
        }} 
        class="space-y-4"
      >
        <div class="space-y-1.5">
          <label for="staff-email" class="block text-xs font-bold text-neutral-700">Email Staf</label>
          <input 
            id="staff-email"
            name="email" 
            type="email" 
            required 
            bind:value={email}
            placeholder="nama@kopirise.id"
            class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
          />
        </div>

        <div class="space-y-1.5">
          <label for="staff-pass" class="block text-xs font-bold text-neutral-700">Kata Sandi</label>
          <input 
            id="staff-pass"
            name="password" 
            type="password" 
            required 
            bind:value={password}
            placeholder="••••••••"
            class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
          />
        </div>

        <button 
          type="submit" 
          disabled={isLoggingIn}
          class="w-full py-3.5 rounded-xl bg-caramel hover:bg-caramelHover text-white font-heading font-extrabold text-xs shadow-lg transition transform active:scale-95 disabled:opacity-50"
        >
          {#if isLoggingIn}
            <span>Memverifikasi Kredensial...</span>
          {:else}
            <span>Masuk ke Sistem</span>
          {/if}
        </button>
      </form>

      <!-- Quick Persona Login Fill for Development & Testing -->
      <div class="pt-4 border-t border-neutral-100 space-y-2">
        <span class="block text-[10px] text-neutral-400 font-bold uppercase tracking-wider text-center">
          Pilihan Akun Uji Coba Cepat (Seeder Dev)
        </span>
        <div class="grid grid-cols-2 gap-2 text-[11px]">
          <button 
            type="button" 
            onclick={() => fillDemo('admin@kopirise.id', 'AdminRise2026!')}
            class="p-2 rounded-xl bg-neutral-50 hover:bg-neutral-100 border border-neutral-200 text-left"
          >
            <strong class="block text-espresso font-bold">Admin / Owner</strong>
            <span class="text-[10px] text-neutral-400">Arga Pramudya</span>
          </button>

          <button 
            type="button" 
            onclick={() => fillDemo('manager@kopirise.id', 'ManagerRise2026!')}
            class="p-2 rounded-xl bg-neutral-50 hover:bg-neutral-100 border border-neutral-200 text-left"
          >
            <strong class="block text-espresso font-bold">Store Manager</strong>
            <span class="text-[10px] text-neutral-400">Siti Rahma</span>
          </button>

          <button 
            type="button" 
            onclick={() => fillDemo('kasir@kopirise.id', 'KasirRise2026!')}
            class="p-2 rounded-xl bg-neutral-50 hover:bg-neutral-100 border border-neutral-200 text-left"
          >
            <strong class="block text-espresso font-bold">Head Cashier</strong>
            <span class="text-[10px] text-neutral-400">Budi Santoso</span>
          </button>

          <button 
            type="button" 
            onclick={() => fillDemo('barista@kopirise.id', 'BaristaRise2026!')}
            class="p-2 rounded-xl bg-neutral-50 hover:bg-neutral-100 border border-neutral-200 text-left"
          >
            <strong class="block text-espresso font-bold">Barista KDS</strong>
            <span class="text-[10px] text-neutral-400">Dimas Arya</span>
          </button>
        </div>
      </div>

    </div>

    <div class="text-center mt-6">
      <a href="/" class="text-xs text-creamWarm/70 hover:text-white transition">
        ← Kembali ke Website Publik Kafe
      </a>
    </div>
  </div>
</div>
