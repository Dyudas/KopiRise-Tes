<script lang="ts">
  import { page } from '$app/stores';

  let { cartCount = 0, onOpenCart } = $props<{
    cartCount?: number;
    onOpenCart?: () => void;
  }>();

  let mobileMenuOpen = $state(false);

  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }
</script>

<header class="sticky top-0 z-40 bg-espresso text-creamWarm shadow-md border-b border-espressoDark">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="flex items-center justify-between h-20">
      
      <!-- Brand Logo -->
      <a href="/" class="flex items-center space-x-3 group">
        <div class="w-11 h-11 rounded-2xl bg-caramel flex items-center justify-center text-white shadow-md group-hover:bg-caramelHover transition transform group-hover:scale-105">
          <svg class="w-6 h-6 fill-current" viewBox="0 0 24 24">
            <path d="M2,21H20V19H2M20,8H18V5H20M20,3H4V13A4,4 0 0,0 8,17H14A4,4 0 0,0 18,13V10H20A2,2 0 0,0 22,8V5C22,3.89 21.1,3 20,3Z" />
          </svg>
        </div>
        <div>
          <div class="flex items-center space-x-2">
            <span class="font-heading font-black text-2xl tracking-tight text-creamWarm">KopiRise</span>
            <span class="text-[10px] uppercase tracking-wider bg-caramel/30 text-caramel border border-caramel/40 px-2 py-0.5 rounded-full font-bold">Jember</span>
          </div>
          <p class="text-[11px] text-creamWarm/70 hidden sm:block">Jl. Semeru, Sumbersari • Cafe & Resto</p>
        </div>
      </a>

      <!-- Desktop Navigation Links -->
      <nav class="hidden md:flex items-center space-x-8 text-sm font-semibold">
        <a 
          href="/" 
          class="transition hover:text-caramel {$page.url.pathname === '/' ? 'text-caramel font-bold' : 'text-creamWarm/80'}"
        >
          Beranda
        </a>
        <a 
          href="/menu" 
          class="transition hover:text-caramel {$page.url.pathname.startsWith('/menu') ? 'text-caramel font-bold' : 'text-creamWarm/80'}"
        >
          Daftar Menu
        </a>
        <a 
          href="/reservasi" 
          class="transition hover:text-caramel {$page.url.pathname.startsWith('/reservasi') ? 'text-caramel font-bold' : 'text-creamWarm/80'}"
        >
          Reservasi Meja
        </a>
        <a 
          href="/reservasi/status" 
          class="transition hover:text-caramel {$page.url.pathname.startsWith('/reservasi/status') ? 'text-caramel font-bold' : 'text-creamWarm/80'}"
        >
          Cek Reservasi
        </a>
      </nav>

      <!-- Action Buttons -->
      <div class="flex items-center space-x-3">
        <!-- Floating Cart Trigger (if cart active) -->
        {#if onOpenCart}
          <button 
            onclick={onOpenCart}
            aria-label="Keranjang Pesanan"
            class="relative p-2.5 rounded-xl bg-espressoDark border border-creamWarm/10 text-creamWarm hover:text-caramel transition shadow-inner"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
            </svg>
            {#if cartCount > 0}
              <span class="absolute -top-1.5 -right-1.5 bg-caramel text-white text-[11px] font-black w-5 h-5 rounded-full flex items-center justify-center shadow">
                {cartCount}
              </span>
            {/if}
          </button>
        {/if}

        <!-- Book Table CTA -->
        <a 
          href="/reservasi" 
          class="hidden sm:inline-flex items-center space-x-2 px-4 py-2.5 rounded-xl bg-caramel hover:bg-caramelHover text-white text-xs font-bold shadow-md transition transform active:scale-95"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <span>Reservasi Meja</span>
        </a>

        <!-- Portal Karyawan & Admin Link -->
        <a 
          href="/admin/login" 
          class="inline-flex items-center space-x-1.5 px-3 py-2 rounded-xl bg-amber-600/30 hover:bg-amber-600 text-amber-200 hover:text-white text-xs font-bold border border-amber-500/40 transition shadow-sm"
          title="Login Karyawan / Admin POS Kasir & Laporan"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <span class="hidden lg:inline">Portal Admin</span>
        </a>

        <!-- Mobile Menu Toggle Button -->
        <button 
          onclick={toggleMobileMenu}
          aria-label="Buka Menu"
          class="md:hidden p-2.5 rounded-xl bg-espressoDark text-creamWarm hover:text-caramel"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            {#if mobileMenuOpen}
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            {:else}
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            {/if}
          </svg>
        </button>
      </div>

    </div>
  </div>

  <!-- Mobile Drawer Menu -->
  {#if mobileMenuOpen}
    <div class="md:hidden bg-espressoDark border-t border-creamWarm/10 px-4 pt-3 pb-6 space-y-3">
      <a 
        href="/" 
        onclick={() => mobileMenuOpen = false}
        class="block py-2 text-base font-semibold text-creamWarm hover:text-caramel"
      >
        Beranda
      </a>
      <a 
        href="/menu" 
        onclick={() => mobileMenuOpen = false}
        class="block py-2 text-base font-semibold text-creamWarm hover:text-caramel"
      >
        Daftar Menu & Kategori
      </a>
      <a 
        href="/reservasi" 
        onclick={() => mobileMenuOpen = false}
        class="block py-2 text-base font-semibold text-creamWarm hover:text-caramel"
      >
        Reservasi Meja
      </a>
      <a 
        href="/reservasi/status" 
        onclick={() => mobileMenuOpen = false}
        class="block py-2 text-base font-semibold text-creamWarm hover:text-caramel"
      >
        Cek Status Reservasi
      </a>
      <a 
        href="/admin/login" 
        onclick={() => mobileMenuOpen = false}
        class="block py-2 text-base font-semibold text-amber-300 hover:text-amber-200"
      >
        🔐 Portal Karyawan & Dashboard Admin
      </a>
      <div class="pt-3 border-t border-creamWarm/10">
        <a 
          href="/reservasi" 
          onclick={() => mobileMenuOpen = false}
          class="w-full text-center block py-3 rounded-xl bg-caramel text-white font-bold text-sm shadow"
        >
          Reservasi Meja Sekarang
        </a>
      </div>
    </div>
  {/if}
</header>
