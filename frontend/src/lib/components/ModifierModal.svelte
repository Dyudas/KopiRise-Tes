<script lang="ts">
  import type { MenuItem, CartItem, CartModifier } from '$lib/types';
  import { formatRupiah } from '$lib/utils/format';

  let { item, onClose, onAddToCart } = $props<{
    item: MenuItem | null;
    onClose: () => void;
    onAddToCart: (cartItem: CartItem) => void;
  }>();

  // State menggunakan Svelte 5 Runes
  let selectedVariantId = $state<string>('');
  let selectedModifiers = $state<Record<string, string[]>>({}); // groupId -> array of optionIds
  let quantity = $state<number>(1);
  let notes = $state<string>('');

  // Inisialisasi default options saat item berubah
  $effect(() => {
    if (item) {
      quantity = 1;
      notes = '';
      
      // Pilih varian pertama jika ada
      if (item.variants && item.variants.length > 0) {
        selectedVariantId = item.variants[0].id;
      } else {
        selectedVariantId = '';
      }

      // Pilih default modifier options
      const initialMods: Record<string, string[]> = {};
      if (item.modifier_groups) {
        for (const grp of item.modifier_groups) {
          initialMods[grp.id] = [];
          for (const opt of grp.options) {
            if (opt.is_default && opt.is_available) {
              initialMods[grp.id].push(opt.id);
            }
          }
          // Jika single selection dan belum ada yang terpilih, pilih opsi pertama yang tersedia
          if (grp.selection_type === 'single' && initialMods[grp.id].length === 0 && grp.options.length > 0) {
            const firstAvail = grp.options.find(o => o.is_available);
            if (firstAvail) initialMods[grp.id].push(firstAvail.id);
          }
        }
      }
      selectedModifiers = initialMods;
    }
  });

  // Kalkulasi Harga Live menggunakan Runes $derived
  const selectedVariant = $derived(
    item?.variants?.find(v => v.id === selectedVariantId)
  );

  const variantDiff = $derived(
    selectedVariant?.price_difference || 0
  );

  const modifiersCost = $derived.by(() => {
    if (!item?.modifier_groups) return 0;
    let cost = 0;
    for (const grp of item.modifier_groups) {
      const selectedIds = selectedModifiers[grp.id] || [];
      for (const optId of selectedIds) {
        const opt = grp.options.find(o => o.id === optId);
        if (opt) cost += opt.additional_price;
      }
    }
    return cost;
  });

  const unitPrice = $derived(
    (item?.base_price || 0) + variantDiff + modifiersCost
  );

  const itemTotal = $derived(
    unitPrice * quantity
  );

  function handleSingleSelect(groupId: string, optionId: string) {
    selectedModifiers[groupId] = [optionId];
  }

  function handleMultiSelect(groupId: string, optionId: string) {
    const current = selectedModifiers[groupId] || [];
    if (current.includes(optionId)) {
      selectedModifiers[groupId] = current.filter(id => id !== optionId);
    } else {
      selectedModifiers[groupId] = [...current, optionId];
    }
  }

  function handleConfirm() {
    if (!item) return;

    const flattenedModifiers: CartModifier[] = [];
    if (item.modifier_groups) {
      for (const grp of item.modifier_groups) {
        const ids = selectedModifiers[grp.id] || [];
        for (const id of ids) {
          const opt = grp.options.find(o => o.id === id);
          if (opt) {
            flattenedModifiers.push({
              option_id: opt.id,
              group_name: grp.name,
              option_name: opt.name,
              price: opt.additional_price
            });
          }
        }
      }
    }

    const cartItem: CartItem = {
      cart_item_id: 'cart-' + Math.random().toString(36).substring(2, 9),
      menu_item_id: item.id,
      name: item.name,
      image_url: item.image_url,
      variant_id: selectedVariant?.id,
      variant_name: selectedVariant?.name,
      unit_price: unitPrice,
      quantity,
      notes,
      modifiers: flattenedModifiers,
      item_total: itemTotal
    };

    onAddToCart(cartItem);
    onClose();
  }
</script>

{#if item}
  <div class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="bg-white rounded-3xl max-w-lg w-full shadow-2xl overflow-hidden border border-neutral-200 animate-in fade-in zoom-in-95 duration-200">
      
      <!-- Modal Header -->
      <div class="relative bg-espresso p-6 text-creamWarm flex items-center justify-between">
        <div>
          <span class="text-[10px] uppercase tracking-wider text-caramel font-bold">Kustomisasi Pesanan</span>
          <h3 class="font-heading font-black text-xl text-creamWarm leading-tight mt-0.5">{item.name}</h3>
        </div>
        <button 
          onclick={onClose}
          aria-label="Tutup"
          class="w-9 h-9 rounded-full bg-creamWarm/10 hover:bg-creamWarm/20 flex items-center justify-center text-creamWarm transition"
        >
          ✕
        </button>
      </div>

      <!-- Scrollable Body -->
      <div class="p-6 max-h-[60vh] overflow-y-auto space-y-6 custom-scrollbar text-xs">
        
        <!-- Varian Menu (jika ada) -->
        {#if item.variants && item.variants.length > 0}
          <div class="space-y-2.5">
            <label class="font-heading font-extrabold text-sm text-espresso block">
              Pilih Varian Ukuran / Sajian <span class="text-red-500">*</span>
            </label>
            <div class="grid grid-cols-2 gap-2">
              {#each item.variants as v}
                <button
                  type="button"
                  disabled={!v.is_available}
                  onclick={() => selectedVariantId = v.id}
                  class="p-3 rounded-2xl border text-left transition flex flex-col justify-between {
                    selectedVariantId === v.id 
                      ? 'border-caramel bg-caramel/10 text-espresso font-bold ring-2 ring-caramel/20' 
                      : 'border-neutral-200 bg-neutral-50 hover:bg-neutral-100 text-neutral-700'
                  }"
                >
                  <span class="text-xs">{v.name}</span>
                  <span class="text-[11px] text-neutral-500 font-semibold mt-1">
                    {v.price_difference > 0 ? `+${formatRupiah(v.price_difference)}` : 'Harga Dasar'}
                  </span>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Modifier Groups (Suhu, Manis, Topping, dll) -->
        {#if item.modifier_groups && item.modifier_groups.length > 0}
          {#each item.modifier_groups as grp}
            <div class="space-y-2.5 pt-4 border-t border-neutral-100">
              <div class="flex justify-between items-center">
                <span class="font-heading font-extrabold text-sm text-espresso">
                  {grp.name}
                  {#if grp.is_required}
                    <span class="text-red-500">*</span>
                  {/if}
                </span>
                <span class="text-[10px] text-neutral-400">
                  {grp.selection_type === 'single' ? 'Pilih salah satu' : 'Bisa pilih lebih dari satu'}
                </span>
              </div>

              <div class="space-y-2">
                {#each grp.options as opt}
                  {@const isChecked = (selectedModifiers[grp.id] || []).includes(opt.id)}
                  <button
                    type="button"
                    disabled={!opt.is_available}
                    onclick={() => grp.selection_type === 'single' ? handleSingleSelect(grp.id, opt.id) : handleMultiSelect(grp.id, opt.id)}
                    class="w-full p-2.5 rounded-xl border flex items-center justify-between text-left transition {
                      isChecked 
                        ? 'border-caramel bg-caramel/10 text-espresso font-bold' 
                        : 'border-neutral-200 bg-white hover:bg-neutral-50 text-neutral-700'
                    }"
                  >
                    <div class="flex items-center space-x-2.5">
                      <div class="w-4 h-4 rounded-full border flex items-center justify-center {
                        isChecked ? 'border-caramel bg-caramel text-white' : 'border-neutral-300'
                      }">
                        {#if isChecked}
                          <span class="text-[9px]">✓</span>
                        {/if}
                      </div>
                      <span class="text-xs">{opt.name}</span>
                    </div>

                    <span class="text-[11px] text-caramel font-semibold">
                      {opt.additional_price > 0 ? `+${formatRupiah(opt.additional_price)}` : 'Gratis'}
                    </span>
                  </button>
                {/each}
              </div>
            </div>
          {/each}
        {/if}

        <!-- Catatan Khusus -->
        <div class="space-y-2 pt-4 border-t border-neutral-100">
          <label for="modifier-notes" class="font-heading font-extrabold text-sm text-espresso block">
            Catatan Khusus Barista / Dapur
          </label>
          <input 
            id="modifier-notes"
            type="text"
            bind:value={notes}
            placeholder="Contoh: Sedikit es, gula dipisah, saus ekstra..."
            class="w-full px-3.5 py-2.5 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
          />
        </div>

        <!-- Quantity Counter -->
        <div class="flex items-center justify-between pt-4 border-t border-neutral-100">
          <span class="font-heading font-extrabold text-sm text-espresso">Jumlah Pesanan</span>
          <div class="flex items-center space-x-3 bg-neutral-100 p-1 rounded-2xl border border-neutral-200">
            <button 
              type="button"
              onclick={() => quantity = Math.max(1, quantity - 1)}
              class="w-8 h-8 rounded-xl bg-white shadow-sm flex items-center justify-center font-black text-neutral-700 hover:bg-neutral-50 active:scale-95"
            >
              -
            </button>
            <span class="font-heading font-black text-sm px-2 text-espresso">{quantity}</span>
            <button 
              type="button"
              onclick={() => quantity++}
              class="w-8 h-8 rounded-xl bg-white shadow-sm flex items-center justify-center font-black text-neutral-700 hover:bg-neutral-50 active:scale-95"
            >
              +
            </button>
          </div>
        </div>

      </div>

      <!-- Modal Footer -->
      <div class="p-6 bg-neutral-50 border-t border-neutral-200 flex items-center justify-between gap-4">
        <div>
          <span class="text-[10px] text-neutral-500 block uppercase font-bold">Total Harga</span>
          <span class="font-heading font-black text-xl text-espresso">{formatRupiah(itemTotal)}</span>
        </div>

        <button
          type="button"
          onclick={handleConfirm}
          class="flex-1 max-w-xs py-3 px-6 rounded-2xl bg-caramel hover:bg-caramelHover text-white font-bold text-xs shadow-md transition transform active:scale-95 text-center"
        >
          Tambah ke Keranjang
        </button>
      </div>

    </div>
  </div>
{/if}
