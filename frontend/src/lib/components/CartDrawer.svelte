<script lang="ts">
  import type { CartItem } from '$lib/types';
  import { formatRupiah } from '$lib/utils/format';

  let { 
    isOpen = false, 
    items = [], 
    tableName = '',
    onClose, 
    onUpdateQty, 
    onRemoveItem,
    onCheckout 
  } = $props<{
    isOpen: boolean;
    items: CartItem[];
    tableName?: string;
    onClose: () => void;
    onUpdateQty: (cartItemId: string, delta: number) => void;
    onRemoveItem: (cartItemId: string) => void;
    onCheckout: () => void;
  }>();

  const subtotal = $derived(
    items.reduce((sum, item) => sum + (item.unit_price * item.quantity), 0)
  );

  const taxAmount = $derived(
    Math.round(subtotal * 0.10) // Pajak PB1 10%
  );

  const grandTotal = $derived(
    subtotal + taxAmount
  );
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 overflow-hidden">
    <!-- Backdrop -->
    <div 
      onclick={onClose}
      class="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
    ></div>

    <div class="fixed inset-y-0 right-0 max-w-full flex pl-10">
      <div class="w-screen max-w-md bg-white shadow-2xl flex flex-col justify-between">
        
        <!-- Drawer Header -->
        <div class="p-6 bg-espresso text-creamWarm flex items-center justify-between border-b border-espressoDark">
          <div>
            <div class="flex items-center space-x-2">
              <h3 class="font-heading font-black text-xl text-creamWarm">Keranjang Pesanan</h3>
              {#if tableName}
                <span class="bg-caramel text-white text-[10px] font-bold px-2 py-0.5 rounded-full">
                  {tableName}
                </span>
              {/if}
            </div>
            <p class="text-xs text-creamWarm/70 mt-0.5">{items.length} item dipilih</p>
          </div>
          <button 
            onclick={onClose}
            aria-label="Tutup Keranjang"
            class="w-8 h-8 rounded-full bg-creamWarm/10 hover:bg-creamWarm/20 flex items-center justify-center text-creamWarm text-sm"
          >
            ✕
          </button>
        </div>

        <!-- Items List -->
        <div class="flex-1 overflow-y-auto p-6 space-y-4 custom-scrollbar">
          {#if items.length === 0}
            <div class="h-full flex flex-col items-center justify-center text-neutral-400 space-y-3 py-16 text-center">
              <div class="w-16 h-16 rounded-3xl bg-neutral-100 flex items-center justify-center text-neutral-400">
                <svg class="w-8 h-8 stroke-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                </svg>
              </div>
              <p class="text-sm font-semibold text-neutral-600">Keranjang masih kosong</p>
              <p class="text-xs text-neutral-400 max-w-xs">Pilih sajian kopi atau makanan lezat favorit Anda dari daftar menu kami.</p>
            </div>
          {:else}
            {#each items as item}
              <div class="p-4 bg-neutral-50 rounded-2xl border border-neutral-200/80 space-y-2.5">
                <div class="flex justify-between items-start gap-2">
                  <div>
                    <h4 class="font-heading font-bold text-xs text-espresso">{item.name}</h4>
                    {#if item.variant_name}
                      <span class="inline-block text-[10px] bg-neutral-200 text-neutral-700 px-1.5 py-0.5 rounded font-medium mt-0.5">
                        {item.variant_name}
                      </span>
                    {/if}
                  </div>
                  <button 
                    onclick={() => onRemoveItem(item.cart_item_id)}
                    class="text-neutral-400 hover:text-red-500 transition text-xs"
                    title="Hapus"
                  >
                    🗑
                  </button>
                </div>

                <!-- Modifiers Breakdown -->
                {#if item.modifiers && item.modifiers.length > 0}
                  <ul class="text-[10px] text-neutral-500 space-y-0.5 pl-2 border-l border-neutral-300">
                    {#each item.modifiers as m}
                      <li class="flex justify-between">
                        <span>• {m.option_name}</span>
                        {#if m.price > 0}
                          <span class="text-neutral-600 font-medium">+{formatRupiah(m.price)}</span>
                        {/if}
                      </li>
                    {/each}
                  </ul>
                {/if}

                {#if item.notes}
                  <p class="text-[10px] text-amber-700 italic bg-amber-50 px-2 py-1 rounded">
                    Catatan: {item.notes}
                  </p>
                {/if}

                <!-- Price and Qty Controller -->
                <div class="pt-2 flex items-center justify-between border-t border-neutral-200/60">
                  <span class="font-heading font-black text-xs text-caramel">
                    {formatRupiah(item.unit_price * item.quantity)}
                  </span>

                  <div class="flex items-center space-x-2 bg-white px-2 py-1 rounded-xl border border-neutral-200 shadow-sm">
                    <button 
                      onclick={() => onUpdateQty(item.cart_item_id, -1)}
                      class="w-5 h-5 rounded flex items-center justify-center text-xs font-bold hover:bg-neutral-100"
                    >
                      -
                    </button>
                    <span class="text-xs font-black px-1">{item.quantity}</span>
                    <button 
                      onclick={() => onUpdateQty(item.cart_item_id, 1)}
                      class="w-5 h-5 rounded flex items-center justify-center text-xs font-bold hover:bg-neutral-100"
                    >
                      +
                    </button>
                  </div>
                </div>
              </div>
            {/each}
          {/if}
        </div>

        <!-- Drawer Footer: Totals & Checkout -->
        {#if items.length > 0}
          <div class="p-6 bg-neutral-50 border-t border-neutral-200 space-y-4">
            <div class="space-y-1.5 text-xs">
              <div class="flex justify-between text-neutral-500">
                <span>Subtotal</span>
                <span class="font-semibold text-neutral-800">{formatRupiah(subtotal)}</span>
              </div>
              <div class="flex justify-between text-neutral-500">
                <span>Pajak Restoran (PB1 10%)</span>
                <span class="font-semibold text-neutral-800">{formatRupiah(taxAmount)}</span>
              </div>
              <div class="flex justify-between text-sm font-extrabold text-espresso pt-2 border-t border-neutral-200">
                <span>Total Pembayaran</span>
                <span class="text-caramel font-black text-base">{formatRupiah(grandTotal)}</span>
              </div>
            </div>

            <button
              onclick={onCheckout}
              class="w-full py-3.5 rounded-2xl bg-caramel hover:bg-caramelHover text-white font-heading font-bold text-xs shadow-lg transition transform active:scale-95 text-center flex items-center justify-center space-x-2"
            >
              <span>Kirim Pesanan Sekarang</span>
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
              </svg>
            </button>
          </div>
        {/if}

      </div>
    </div>
  </div>
{/if}
