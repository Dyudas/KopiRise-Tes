<script lang="ts">
  import { page } from '$app/stores';
  import type { Snippet } from 'svelte';
  import type { LayoutData } from './$types';

  let { data, children } = $props<{
    data: LayoutData;
    children: Snippet;
  }>();

  const user = $derived(data.user);
  const isAdminOrManager = $derived(
    user?.roles?.includes('admin') || user?.roles?.includes('manager')
  );
</script>

{#if $page.url.pathname === '/admin/login'}
  <!-- Login page without admin shell -->
  {@render children()}
{:else}
  <div class="min-h-screen bg-[#F4F3EF] flex">
    
    <!-- LEFT SIDEBAR -->
    <aside class="w-64 bg-espresso text-creamWarm shrink-0 hidden md:flex flex-col justify-between shadow-2xl border-r border-espressoDark">
      <div>
        <!-- Brand Header -->
        <div class="h-20 px-6 flex items-center space-x-3 border-b border-creamWarm/10">
          <div class="w-10 h-10 rounded-2xl bg-caramel flex items-center justify-center text-white shadow-md">
            <svg class="w-6 h-6 fill-current" viewBox="0 0 24 24">
              <path d="M2,21H20V19H2M20,8H18V5H20M20,3H4V13A4,4 0 0,0 8,17H14A4,4 0 0,0 18,13V10H20A2,2 0 0,0 22,8V5C22,3.89 21.1,3 20,3Z" />
            </svg>
          </div>
          <div>
            <h1 class="font-heading font-black text-xl text-creamWarm tracking-tight leading-none">KopiRise</h1>
            <span class="text-[10px] text-creamWarm/60 tracking-wider uppercase font-semibold">Staff & POS Engine</span>
          </div>
        </div>

        <!-- Navigation Links -->
        <nav class="p-4 space-y-1.5 text-xs font-semibold">
          <span class="block px-3 py-2 text-[10px] uppercase font-bold text-creamWarm/40 tracking-wider">
            Operasional
          </span>

          <a 
            href="/admin/pos" 
            class="flex items-center space-x-3 px-3 py-2.5 rounded-xl transition {
              $page.url.pathname === '/admin/pos' 
                ? 'bg-caramel text-white font-bold shadow' 
                : 'text-creamWarm/80 hover:bg-creamWarm/10 hover:text-white'
            }"
          >
            <span>🖥️</span>
            <span>Aplikasi Kasir / POS</span>
          </a>

          <a 
            href="/admin/kds" 
            class="flex items-center space-x-3 px-3 py-2.5 rounded-xl transition {
              $page.url.pathname === '/admin/kds' 
                ? 'bg-caramel text-white font-bold shadow' 
                : 'text-creamWarm/80 hover:bg-creamWarm/10 hover:text-white'
            }"
          >
            <span>🔥</span>
            <span>Kitchen KDS Barista</span>
          </a>

          <a 
            href="/admin/tables" 
            class="flex items-center space-x-3 px-3 py-2.5 rounded-xl transition {
              $page.url.pathname === '/admin/tables' 
                ? 'bg-caramel text-white font-bold shadow' 
                : 'text-creamWarm/80 hover:bg-creamWarm/10 hover:text-white'
            }"
          >
            <span>🪑</span>
            <span>Denah Meja & Kartu QR</span>
          </a>

          {#if isAdminOrManager}
            <span class="block px-3 pt-4 pb-2 text-[10px] uppercase font-bold text-creamWarm/40 tracking-wider">
              Manajemen & Laporan
            </span>

            <a 
              href="/admin/menu" 
              class="flex items-center space-x-3 px-3 py-2.5 rounded-xl transition {
                $page.url.pathname === '/admin/menu' 
                  ? 'bg-caramel text-white font-bold shadow' 
                  : 'text-creamWarm/80 hover:bg-creamWarm/10 hover:text-white'
              }"
            >
              <span>📋</span>
              <span>Kelola Menu & Stok</span>
            </a>

            <a 
              href="/admin/reports" 
              class="flex items-center space-x-3 px-3 py-2.5 rounded-xl transition {
                $page.url.pathname === '/admin/reports' 
                  ? 'bg-caramel text-white font-bold shadow' 
                  : 'text-creamWarm/80 hover:bg-creamWarm/10 hover:text-white'
              }"
            >
              <span>📊</span>
              <span>Laporan Laba Rugi & HPP</span>
            </a>
          {/if}
        </nav>
      </div>

      <!-- User Profile & Logout Box -->
      <div class="p-4 border-t border-creamWarm/10 bg-espressoDark/60">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2.5 overflow-hidden">
            <div class="w-8 h-8 rounded-full bg-caramel/30 border border-caramel/50 text-caramel flex items-center justify-center font-bold text-xs uppercase shrink-0">
              {user?.name ? user.name.substring(0, 2) : 'KP'}
            </div>
            <div class="truncate text-xs">
              <p class="font-bold text-creamWarm truncate">{user?.name || 'Petugas'}</p>
              <p class="text-[10px] text-creamWarm/60 uppercase font-semibold truncate">
                {user?.roles?.[0] || 'Staff'}
              </p>
            </div>
          </div>

          <a 
            href="/admin/logout" 
            title="Keluar"
            class="p-1.5 rounded-lg text-creamWarm/60 hover:text-red-400 hover:bg-creamWarm/10 transition"
          >
            🚪
          </a>
        </div>
      </div>
    </aside>

    <!-- MAIN RIGHT CONTENT VIEWPORT -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Top Bar -->
      <header class="h-16 bg-white border-b border-neutral-200 px-6 flex items-center justify-between shadow-sm">
        <div class="flex items-center space-x-3">
          <span class="text-xs font-semibold text-neutral-500">Cabang:</span>
          <span class="text-xs font-bold text-espresso bg-neutral-100 px-2.5 py-1 rounded-full border border-neutral-200">
            KopiRise Jember (Jl. Semeru)
          </span>
        </div>

        <div class="flex items-center space-x-4 text-xs">
          <span class="text-neutral-500 hidden sm:inline">
            Status Sistem: <strong class="text-emerald-600">Terhubung SQLite WAL</strong>
          </span>
          <a 
            href="/" 
            target="_blank"
            class="text-xs font-bold text-caramel hover:text-caramelHover flex items-center space-x-1"
          >
            <span>Buka Web Publik</span>
            <span>↗</span>
          </a>
        </div>
      </header>

      <!-- Content Page -->
      <main class="flex-1 p-6 overflow-y-auto custom-scrollbar">
        {@render children()}
      </main>
    </div>

  </div>
{/if}
