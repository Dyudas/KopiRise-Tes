<script lang="ts">
  import type { PageData } from './$types';
  import MenuCard from '$lib/components/MenuCard.svelte';
  import ModifierModal from '$lib/components/ModifierModal.svelte';
  import CartDrawer from '$lib/components/CartDrawer.svelte';
  import type { MenuItem, CartItem } from '$lib/types';

  let { data } = $props<{ data: PageData }>();

  // State kustomisasi modal & keranjang
  let selectedItem = $state<MenuItem | null>(null);
  let cartItems = $state<CartItem[]>([]);
  let isCartOpen = $state(false);

  function openModifier(item: MenuItem) {
    selectedItem = item;
  }

  function handleAddToCart(cartItem: CartItem) {
    cartItems = [...cartItems, cartItem];
    isCartOpen = true;
  }

  function updateQty(cartItemId: string, delta: number) {
    cartItems = cartItems.map(item => {
      if (item.cart_item_id === cartItemId) {
        const newQty = item.quantity + delta;
        return newQty > 0 ? { ...item, quantity: newQty } : null;
      }
      return item;
    }).filter(Boolean) as CartItem[];
  }

  function removeItem(cartItemId: string) {
    cartItems = cartItems.filter(item => item.cart_item_id !== cartItemId);
  }
</script>

<svelte:head>
  <title>KopiRise Jember — Cafe & Resto Spesialti di Jl. Semeru</title>
  <meta name="description" content="KopiRise Cafe & Resto Jember. Sajian kopi spesialti Ijen Raung, aneka makanan resto otentik, reservasi meja rapat VIP, dan pemesanan online mudah." />
</svelte:head>

<div class="space-y-16 pb-20">
  
  <!-- HERO SECTION -->
  <section class="relative bg-espresso text-creamWarm overflow-hidden min-h-[580px] flex items-center">
    <div 
      class="absolute inset-0 bg-cover bg-center opacity-25 mix-blend-overlay scale-105 transform transition duration-1000"
      style="background-image: url('https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?auto=format&fit=crop&w=1600&q=80');"
    ></div>
    <div class="absolute inset-0 bg-gradient-to-r from-espressoDark via-espresso/95 to-transparent"></div>

    <div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 lg:py-28">
      <div class="max-w-2xl space-y-6">
        <div class="inline-flex items-center space-x-2 bg-caramel/20 border border-caramel/40 px-3.5 py-1.5 rounded-full text-xs font-bold text-caramel">
          <span>✨ Spesialti Kopi & Kuliner Otentik di Jember</span>
        </div>

        <h1 class="font-heading font-black text-4xl sm:text-5xl lg:text-6xl tracking-tight text-creamWarm leading-[1.15]">
          Secangkir Ketenangan, <span class="text-caramel">Berjuta Cerita</span> di KopiRise.
        </h1>

        <p class="text-sm sm:text-base text-creamWarm/80 leading-relaxed font-normal">
          Menghadirkan harmoni seduhan biji kopi terbaik dari lereng Gunung Ijen & Raung, berpadu dengan hidangan resto pilihan dalam suasana hangat nan asri di Jl. Semeru, Sumbersari, Jember.
        </p>

        <div class="pt-4 flex flex-wrap items-center gap-4">
          <a 
            href="/menu" 
            class="px-6 py-3.5 rounded-2xl bg-caramel hover:bg-caramelHover text-white font-heading font-bold text-xs sm:text-sm shadow-xl transition transform hover:-translate-y-0.5 active:scale-95 flex items-center space-x-2"
          >
            <span>Jelajahi Menu</span>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
            </svg>
          </a>

          <a 
            href="/reservasi" 
            class="px-6 py-3.5 rounded-2xl bg-creamWarm/10 hover:bg-creamWarm/20 border border-creamWarm/20 text-creamWarm font-heading font-bold text-xs sm:text-sm backdrop-blur-sm transition"
          >
            Reservasi Meja
          </a>

          <a 
            href="https://maps.google.com/?q=Jl.+Semeru+Karangrejo+Sumbersari+Jember" 
            target="_blank" 
            rel="noopener noreferrer"
            class="px-4 py-3.5 rounded-2xl text-creamWarm/70 hover:text-white text-xs font-semibold flex items-center space-x-1.5 transition"
          >
            <svg class="w-4 h-4 text-caramel" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <span>Petunjuk Lokasi</span>
          </a>
        </div>
      </div>
    </div>
  </section>

  <!-- HIGHLIGHT VALUES -->
  <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
      <div class="p-8 rounded-3xl bg-white border border-neutral-200/80 shadow-sm space-y-3">
        <div class="w-12 h-12 rounded-2xl bg-caramel/10 text-caramel flex items-center justify-center font-black text-xl">
          ☕
        </div>
        <h3 class="font-heading font-extrabold text-lg text-espresso">Biji Kopi Spesialti Lokal</h3>
        <p class="text-xs text-neutral-500 leading-relaxed">
          Dipetik langsung dari perkebunan lereng Ijen & Raung, di-roasting dengan presisi demi profil rasa fruity, floral, dan keasaman seimbang.
        </p>
      </div>

      <div class="p-8 rounded-3xl bg-white border border-neutral-200/80 shadow-sm space-y-3">
        <div class="w-12 h-12 rounded-2xl bg-amber-500/10 text-amber-600 flex items-center justify-center font-black text-xl">
          🍲
        </div>
        <h3 class="font-heading font-extrabold text-lg text-espresso">Resto Otentik & Higienis</h3>
        <p class="text-xs text-neutral-500 leading-relaxed">
          Dari Nasi Goreng Spesial KopiRise hingga Tahu Walik Krispi khas Jember, seluruh sajian diracik segar dari bahan baku bermutu tinggi.
        </p>
      </div>

      <div class="p-8 rounded-3xl bg-white border border-neutral-200/80 shadow-sm space-y-3">
        <div class="w-12 h-12 rounded-2xl bg-emerald-500/10 text-emerald-600 flex items-center justify-center font-black text-xl">
          🌿
        </div>
        <h3 class="font-heading font-extrabold text-lg text-espresso">4 Pilihan Suasana Nyaman</h3>
        <p class="text-xs text-neutral-500 leading-relaxed">
          Indoor AC Non-Smoking, Outdoor Garden rindang, Mezzanine santai lantai 2, serta Ruang VIP ber-AC lengkap untuk rapat dan acara keluarga.
        </p>
      </div>
    </div>
  </section>

  <!-- MENU UNGGULAN (FEATURED ITEMS) -->
  <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8">
    <div class="flex flex-col sm:flex-row items-start sm:items-end justify-between gap-4">
      <div>
        <span class="text-xs font-bold text-caramel uppercase tracking-widest">Favorit Pelanggan</span>
        <h2 class="font-heading font-black text-2xl sm:text-3xl text-espresso mt-1">Menu Unggulan KopiRise</h2>
      </div>
      <a href="/menu" class="text-xs font-bold text-caramel hover:text-caramelHover flex items-center space-x-1">
        <span>Lihat Semua Menu</span>
        <span>→</span>
      </a>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each data.featuredItems as item}
        <MenuCard {item} onSelect={openModifier} />
      {/each}
    </div>
  </section>

  <!-- SUASANA AREA KAFE & RESTO -->
  <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8">
    <div>
      <span class="text-xs font-bold text-caramel uppercase tracking-widest">Galeri & Fasilitas</span>
      <h2 class="font-heading font-black text-2xl sm:text-3xl text-espresso mt-1">Suasana Kafe & Resto</h2>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="rounded-3xl overflow-hidden bg-white border border-neutral-200 shadow-sm group">
        <div class="h-44 bg-cover bg-center group-hover:scale-105 transition duration-500" style="background-image: url('https://images.unsplash.com/photo-1554118811-1e0d58224f24?auto=format&fit=crop&w=600&q=80')"></div>
        <div class="p-4 space-y-1">
          <h4 class="font-heading font-bold text-sm text-espresso">Indoor AC Non-Smoking</h4>
          <p class="text-[11px] text-neutral-500">Sejuk, tenang, ideal untuk bekerja (WFC) & tugas kuliah.</p>
        </div>
      </div>

      <div class="rounded-3xl overflow-hidden bg-white border border-neutral-200 shadow-sm group">
        <div class="h-44 bg-cover bg-center group-hover:scale-105 transition duration-500" style="background-image: url('https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?auto=format&fit=crop&w=600&q=80')"></div>
        <div class="p-4 space-y-1">
          <h4 class="font-heading font-bold text-sm text-espresso">Outdoor Garden Terrace</h4>
          <p class="text-[11px] text-neutral-500">Pemandangan asri tanaman hijau, sejuk di sore hari.</p>
        </div>
      </div>

      <div class="rounded-3xl overflow-hidden bg-white border border-neutral-200 shadow-sm group">
        <div class="h-44 bg-cover bg-center group-hover:scale-105 transition duration-500" style="background-image: url('https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=600&q=80')"></div>
        <div class="p-4 space-y-1">
          <h4 class="font-heading font-bold text-sm text-espresso">Mezzanine Floor (Lt 2)</h4>
          <p class="text-[11px] text-neutral-500">Suasana santai pemandangan terbuka, cocok untuk nongkrong.</p>
        </div>
      </div>

      <div class="rounded-3xl overflow-hidden bg-white border border-neutral-200 shadow-sm group">
        <div class="h-44 bg-cover bg-center group-hover:scale-105 transition duration-500" style="background-image: url('https://images.unsplash.com/photo-1497366216548-37526070297c?auto=format&fit=crop&w=600&q=80')"></div>
        <div class="p-4 space-y-1">
          <h4 class="font-heading font-bold text-sm text-espresso">VIP Room Semeru</h4>
          <p class="text-[11px] text-neutral-500">Ruangan privat ber-AC dengan TV HDMI, kapasitas hingga 18 tamu.</p>
        </div>
      </div>
    </div>
  </section>

  <!-- BANNER RESERVASI CTA -->
  <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="rounded-3xl bg-gradient-to-r from-espresso to-espressoDark p-8 sm:p-12 text-creamWarm shadow-2xl relative overflow-hidden flex flex-col md:flex-row items-center justify-between gap-8">
      <div class="space-y-3 max-w-xl">
        <span class="text-xs font-bold text-caramel uppercase tracking-widest">Bebas Antre & Pasti Dapat Meja</span>
        <h3 class="font-heading font-black text-2xl sm:text-3xl text-creamWarm">Rencanakan Kunjungan Anda Sekarang</h3>
        <p class="text-xs sm:text-sm text-creamWarm/80 leading-relaxed">
          Pesan meja santai, tempat reuni keluarga, atau ruang rapat bisnis VIP tanpa biaya pemesanan awal.
        </p>
      </div>
      <a 
        href="/reservasi"
        class="shrink-0 px-8 py-4 rounded-2xl bg-caramel hover:bg-caramelHover text-white font-heading font-extrabold text-sm shadow-xl transition transform hover:scale-105 active:scale-95"
      >
        Form Reservasi Online
      </a>
    </div>
  </section>

</div>

<!-- Modal Kustomisasi Varian -->
<ModifierModal 
  item={selectedItem} 
  onClose={() => selectedItem = null} 
  onAddToCart={handleAddToCart} 
/>

<!-- Drawer Keranjang -->
<CartDrawer 
  isOpen={isCartOpen}
  items={cartItems}
  onClose={() => isCartOpen = false}
  onUpdateQty={updateQty}
  onRemoveItem={removeItem}
  onCheckout={() => {
    alert("Pesanan keranjang Anda akan diproses ke kasir POS!");
  }}
/>
