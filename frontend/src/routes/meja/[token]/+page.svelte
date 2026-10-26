<script lang="ts">
  import type { PageData } from './$types';
  import MenuCard from '$lib/components/MenuCard.svelte';
  import ModifierModal from '$lib/components/ModifierModal.svelte';
  import CartDrawer from '$lib/components/CartDrawer.svelte';
  import type { MenuItem, CartItem, CreateOrderRequest } from '$lib/types';
  import { fetchApi } from '$lib/api/client';

  let { data } = $props<{ data: PageData }>();

  // State
  let customerName = $state<string>('');
  let selectedCategorySlug = $state<string>('all');
  let searchQuery = $state<string>('');
  
  // Cart & Modals
  let selectedItem = $state<MenuItem | null>(null);
  let cartItems = $state<CartItem[]>([]);
  let isCartOpen = $state(false);
  let isSubmitting = $state(false);
  let completedOrderNumber = $state<string | null>(null);

  const filteredItems = $derived.by(() => {
    let list = [...data.items];

    if (selectedCategorySlug !== 'all') {
      const cat = data.categories.find(c => c.slug === selectedCategorySlug);
      if (cat) list = list.filter(i => i.category_id === cat.id);
    }

    if (searchQuery.trim() !== '') {
      const q = searchQuery.toLowerCase().trim();
      list = list.filter(i => 
        i.name.toLowerCase().includes(q) || 
        (i.description && i.description.toLowerCase().includes(q))
      );
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

  async function handleCheckout() {
    if (cartItems.length === 0) return;
    isSubmitting = true;

    const payload: CreateOrderRequest = {
      order_type: 'dine_in',
      table_id: data.table.id,
      customer_name: customerName || `Pelanggan Meja ${data.table.table_number}`,
      idempotency_key: 'qr-order-' + Math.random().toString(36).substring(2, 12),
      items: cartItems.map(item => ({
        menu_item_id: item.menu_item_id,
        variant_id: item.variant_id,
        quantity: item.quantity,
        notes: item.notes,
        modifier_option_ids: item.modifiers.map(m => m.option_id)
      }))
    };

    try {
      const res = await fetchApi<{ id: string; order_number: string }>('/public/orders', {
        method: 'POST',
        body: JSON.stringify(payload)
      });

      completedOrderNumber = res.order_number;
      cartItems = [];
      isCartOpen = false;
    } catch (err: any) {
      alert(`Gagal mengirim pesanan: ${err.message || 'Silakan hubungi kasir.'}`);
    } finally {
      isSubmitting = false;
    }
  }
</script>

<svelte:head>
  <title>Pesan dari Meja {data.table.table_number} — KopiRise Jember</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
  
  <!-- Active Table Header Card -->
  <div class="bg-gradient-to-r from-espresso to-espressoDark rounded-3xl p-6 sm:p-8 text-creamWarm shadow-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
    <div>
      <div class="inline-flex items-center space-x-2 bg-caramel/20 border border-caramel/40 px-3 py-1 rounded-full text-xs font-bold text-caramel mb-2">
        <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
        <span>QR Code Meja Terverifikasi</span>
      </div>
      <h1 class="font-heading font-black text-2xl sm:text-3xl text-creamWarm">
        Pemesanan Meja <span class="text-caramel">{data.table.table_number}</span>
      </h1>
      <p class="text-xs text-creamWarm/70 mt-1">
        Area: <strong>{data.table.area_name || 'Resto & Cafe Area'}</strong> • Kapasitas {data.table.capacity} Tamu
      </p>
    </div>

    <!-- Optional Customer Name -->
    <div class="w-full sm:w-auto">
      <label for="qr-customer-name" class="block text-[10px] text-creamWarm/60 uppercase font-bold mb-1">Nama Pemesan (Opsional)</label>
      <input 
        id="qr-customer-name"
        type="text" 
        bind:value={customerName}
        placeholder="Atas nama siapa?"
        class="w-full sm:w-48 px-3.5 py-2 rounded-xl bg-creamWarm/10 border border-creamWarm/20 text-creamWarm text-xs focus:ring-2 focus:ring-caramel focus:outline-none placeholder:text-creamWarm/40"
      />
    </div>
  </div>

  {#if completedOrderNumber}
    <!-- Order Submitted Success View -->
    <div class="bg-white rounded-3xl border border-emerald-300 p-8 text-center space-y-6 shadow-xl animate-in zoom-in-95 duration-200">
      <div class="w-16 h-16 rounded-3xl bg-emerald-100 text-emerald-700 flex items-center justify-center font-black text-2xl mx-auto">
        ✓
      </div>
      <div class="space-y-2">
        <span class="text-xs font-bold text-emerald-700 uppercase tracking-wider">Pesanan Berhasil Dikirim ke Barista & Dapur!</span>
        <h2 class="font-heading font-black text-2xl text-espresso">
          Nomor Pesanan: <span class="font-mono text-caramel">{completedOrderNumber}</span>
        </h2>
        <p class="text-xs text-neutral-600 max-w-md mx-auto">
          Silakan bersantai sejenak di Meja <strong>{data.table.table_number}</strong>. Barista dan juru masak kami sedang menyiapkan sajian terbaik untuk Anda.
        </p>
      </div>

      <div class="pt-4 flex justify-center gap-3">
        <button 
          onclick={() => completedOrderNumber = null}
          class="px-6 py-3 rounded-2xl bg-espresso hover:bg-caramel text-creamWarm font-bold text-xs shadow transition"
        >
          Pesan Menu Tambahan
        </button>
      </div>
    </div>
  {:else}
    <!-- Search & Categories Bar -->
    <div class="space-y-4">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-3 bg-white p-3.5 rounded-2xl border border-neutral-200 shadow-sm">
        <div class="relative w-full sm:w-80">
          <input 
            type="text" 
            bind:value={searchQuery}
            placeholder="Cari menu pesanan..."
            class="w-full pl-9 pr-4 py-2 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
          />
          <span class="absolute left-3 top-2.5 text-neutral-400 text-xs">🔍</span>
        </div>

        <div class="text-xs text-neutral-500 font-semibold text-right w-full sm:w-auto">
          {filteredItems.length} menu tersedia
        </div>
      </div>

      <!-- Categories Scroll -->
      <div class="flex items-center space-x-2 overflow-x-auto pb-2 custom-scrollbar">
        <button
          onclick={() => selectedCategorySlug = 'all'}
          class="px-3.5 py-1.5 rounded-xl text-xs font-bold transition shrink-0 {
            selectedCategorySlug === 'all'
              ? 'bg-espresso text-creamWarm shadow'
              : 'bg-white border border-neutral-200 text-neutral-600'
          }"
        >
          Semua Menu
        </button>
        {#each data.categories as cat}
          <button
            onclick={() => selectedCategorySlug = cat.slug}
            class="px-3.5 py-1.5 rounded-xl text-xs font-bold transition shrink-0 {
              selectedCategorySlug === cat.slug
                ? 'bg-espresso text-creamWarm shadow'
                : 'bg-white border border-neutral-200 text-neutral-600'
            }"
          >
            {cat.name}
          </button>
        {/each}
      </div>

      <!-- Menu Grid -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {#each filteredItems as item}
          <MenuCard {item} onSelect={openModifier} />
        {/each}
      </div>
    </div>
  {/if}

</div>

<!-- Floating Cart Button for Table Order -->
{#if cartItems.length > 0 && !completedOrderNumber}
  <div class="fixed bottom-6 right-6 z-40">
    <button
      onclick={() => isCartOpen = true}
      class="px-5 py-3.5 rounded-full bg-caramel hover:bg-caramelHover text-white font-heading font-bold text-xs shadow-2xl flex items-center space-x-3 transition transform hover:scale-105 active:scale-95"
    >
      <div class="w-6 h-6 rounded-full bg-espresso text-creamWarm flex items-center justify-center text-xs font-black">
        {cartItems.reduce((s, i) => s + i.quantity, 0)}
      </div>
      <span>Kirim Pesanan Meja {data.table.table_number}</span>
      <span class="font-black">
        →
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
  tableName={`Meja ${data.table.table_number}`}
  onClose={() => isCartOpen = false}
  onUpdateQty={updateQty}
  onRemoveItem={removeItem}
  onCheckout={handleCheckout}
/>
