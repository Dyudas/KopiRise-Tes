<script lang="ts">
  import type { PageData } from './$types';
  import MenuCard from '$lib/components/MenuCard.svelte';
  import ModifierModal from '$lib/components/ModifierModal.svelte';
  import CartDrawer from '$lib/components/CartDrawer.svelte';
  import type { MenuItem, CartItem } from '$lib/types';

  let { data } = $props<{ data: PageData }>();

  // Filter & Search State menggunakan Svelte 5 Runes
  let selectedCategorySlug = $state<string>('all');
  let searchQuery = $state<string>('');
  let sortBy = $state<string>('popular');

  // Modal & Cart State
  let selectedItem = $state<MenuItem | null>(null);
  let cartItems = $state<CartItem[]>([]);
  let isCartOpen = $state(false);

  // Derived filtered items
  const filteredItems = $derived.by(() => {
    let list = [...data.items];

    // Filter kategori
    if (selectedCategorySlug !== 'all') {
      const cat = data.categories.find(c => c.slug === selectedCategorySlug);
      if (cat) {
        list = list.filter(i => i.category_id === cat.id);
      }
    }

    // Filter pencarian
    if (searchQuery.trim() !== '') {
      const q = searchQuery.toLowerCase().trim();
      list = list.filter(i => 
        i.name.toLowerCase().includes(q) || 
        (i.description && i.description.toLowerCase().includes(q))
      );
    }

    // Sorting
    if (sortBy === 'price_asc') {
      list.sort((a, b) => a.base_price - b.base_price);
    } else if (sortBy === 'price_desc') {
      list.sort((a, b) => b.base_price - a.base_price);
    } else if (sortBy === 'name') {
      list.sort((a, b) => a.name.localeCompare(b.name));
    } else {
      // Popular (featured first)
      list.sort((a, b) => (b.is_featured ? 1 : 0) - (a.is_featured ? 1 : 0));
    }

    return list;
  });

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
  <title>Daftar Menu & Harga — KopiRise Cafe & Resto Jember</title>
  <meta name="description" content="Katalog menu lengkap KopiRise Jember: Kopi Spesialti Ijen Raung, Kopi Susu Aren, Makanan Utama, Snack, dan Pastry. Harga terjangkau dan transparan." />
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 space-y-8">
  
  <!-- Header Page -->
  <div class="text-center max-w-2xl mx-auto space-y-3">
    <span class="text-xs font-bold text-caramel uppercase tracking-widest">Katalog Hidangan</span>
    <h1 class="font-heading font-black text-3xl sm:text-4xl text-espresso">Daftar Menu KopiRise</h1>
    <p class="text-xs sm:text-sm text-neutral-500">
      Temukan racikan kopi spesialti, minuman segar, hidangan resto nusantara, dan kudapan manis favorit Anda.
    </p>
  </div>

  <!-- Search & Sorting Toolbar -->
  <div class="flex flex-col md:flex-row items-center justify-between gap-4 bg-white p-4 rounded-3xl border border-neutral-200 shadow-sm">
    
    <!-- Search Bar -->
    <div class="relative w-full md:w-96">
      <input 
        type="text" 
        bind:value={searchQuery}
        placeholder="Cari menu (misal: aren, ijen, nasi goreng)..."
        class="w-full pl-10 pr-4 py-2.5 rounded-2xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
      />
      <svg class="w-4 h-4 text-neutral-400 absolute left-3.5 top-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      {#if searchQuery}
        <button 
          onclick={() => searchQuery = ''}
          class="absolute right-3 top-2.5 text-neutral-400 hover:text-neutral-600 text-xs"
        >
          ✕
        </button>
      {/if}
    </div>

    <!-- Sorting Options -->
    <div class="flex items-center space-x-2 w-full md:w-auto justify-end text-xs">
      <span class="text-neutral-500 font-medium shrink-0">Urutkan:</span>
      <select 
        bind:value={sortBy}
        class="px-3.5 py-2 rounded-2xl border border-neutral-200 bg-neutral-50 text-neutral-700 text-xs font-semibold focus:ring-2 focus:ring-caramel focus:outline-none"
      >
        <option value="popular">Paling Populer</option>
        <option value="price_asc">Harga: Rendah ke Tinggi</option>
        <option value="price_desc">Harga: Tinggi ke Rendah</option>
        <option value="name">Nama (A — Z)</option>
      </select>
    </div>

  </div>

  <!-- Category Pills (Horizontal Scrollable) -->
  <div class="flex items-center space-x-2 overflow-x-auto pb-2 custom-scrollbar">
    <button
      onclick={() => selectedCategorySlug = 'all'}
      class="px-4 py-2 rounded-2xl text-xs font-bold transition shrink-0 {
        selectedCategorySlug === 'all'
          ? 'bg-espresso text-creamWarm shadow-md'
          : 'bg-white border border-neutral-200 text-neutral-600 hover:border-caramel'
      }"
    >
      Semua Kategori ({data.items.length})
    </button>
    {#each data.categories as cat}
      {@const count = data.items.filter(i => i.category_id === cat.id).length}
      <button
        onclick={() => selectedCategorySlug = cat.slug}
        class="px-4 py-2 rounded-2xl text-xs font-bold transition shrink-0 {
          selectedCategorySlug === cat.slug
            ? 'bg-espresso text-creamWarm shadow-md'
            : 'bg-white border border-neutral-200 text-neutral-600 hover:border-caramel'
        }"
      >
        {cat.name} ({count})
      </button>
    {/each}
  </div>

  <!-- Menu Grid -->
  {#if filteredItems.length === 0}
    <div class="text-center py-20 bg-white rounded-3xl border border-neutral-200 p-8 space-y-3">
      <div class="w-16 h-16 rounded-3xl bg-neutral-100 mx-auto flex items-center justify-center text-neutral-400">
        🔍
      </div>
      <h3 class="font-heading font-extrabold text-base text-espresso">Menu Tidak Ditemukan</h3>
      <p class="text-xs text-neutral-500 max-w-sm mx-auto">
        Tidak ada menu yang sesuai dengan kata kunci "{searchQuery}". Coba kata kunci lain atau pilih kategori berbeda.
      </p>
      <button 
        onclick={() => { searchQuery = ''; selectedCategorySlug = 'all'; }}
        class="mt-2 px-4 py-2 bg-neutral-100 hover:bg-neutral-200 text-neutral-700 rounded-xl text-xs font-bold"
      >
        Reset Pencarian
      </button>
    </div>
  {:else}
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      {#each filteredItems as item}
        <MenuCard {item} onSelect={openModifier} />
      {/each}
    </div>
  {/if}

</div>

<!-- Floating Cart Button -->
{#if cartItems.length > 0}
  <div class="fixed bottom-6 right-6 z-40">
    <button
      onclick={() => isCartOpen = true}
      class="px-5 py-3.5 rounded-full bg-espresso hover:bg-espressoDark text-creamWarm font-heading font-bold text-xs shadow-2xl flex items-center space-x-3 transition transform hover:scale-105 active:scale-95 border border-creamWarm/10"
    >
      <div class="w-6 h-6 rounded-full bg-caramel text-white flex items-center justify-center text-xs font-black">
        {cartItems.reduce((s, i) => s + i.quantity, 0)}
      </div>
      <span>Lihat Pesanan</span>
      <span class="text-caramel font-black">
        {formatRupiah(cartItems.reduce((s, i) => s + i.item_total, 0))}
      </span>
    </button>
  </div>
{/if}

<!-- Modals -->
<ModifierModal 
  item={selectedItem} 
  onClose={() => selectedItem = null} 
  onAddToCart={handleAddToCart} 
/>

<CartDrawer 
  isOpen={isCartOpen}
  items={cartItems}
  onClose={() => isCartOpen = false}
  onUpdateQty={updateQty}
  onRemoveItem={removeItem}
  onCheckout={() => {
    alert("Pesanan Anda sedang dikirim ke sistem kasir KopiRise...");
  }}
/>
