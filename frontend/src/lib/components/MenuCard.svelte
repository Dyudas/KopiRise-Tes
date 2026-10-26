<script lang="ts">
  import type { MenuItem } from '$lib/types';
  import { formatRupiah } from '$lib/utils/format';

  let { item, onSelect } = $props<{
    item: MenuItem;
    onSelect: (item: MenuItem) => void;
  }>();

  const hasOptions = $derived(
    (item.variants && item.variants.length > 0) ||
    (item.modifier_groups && item.modifier_groups.length > 0)
  );
</script>

<div class="bg-white rounded-3xl border border-neutral-200/80 shadow-sm hover:shadow-xl hover:border-caramel/50 transition duration-300 overflow-hidden flex flex-col justify-between group">
  <div>
    <!-- Image Thumbnail -->
    <div class="relative h-48 sm:h-52 w-full bg-neutral-100 overflow-hidden">
      {#if item.image_url}
        <img 
          src={item.image_url} 
          alt={item.name} 
          class="w-full h-full object-cover group-hover:scale-105 transition duration-500" 
          loading="lazy"
        />
      {:else}
        <div class="w-full h-full flex items-center justify-center bg-espresso/5 text-espresso/30">
          <svg class="w-12 h-12 stroke-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
      {/if}

      <!-- Badges: Featured / Availability -->
      <div class="absolute top-3 left-3 flex flex-col gap-1.5">
        {#if item.is_featured}
          <span class="bg-caramel text-white text-[10px] font-extrabold px-2.5 py-1 rounded-full shadow uppercase tracking-wide">
            Best Seller
          </span>
        {/if}
        {#if !item.is_available}
          <span class="bg-red-600 text-white text-[10px] font-extrabold px-2.5 py-1 rounded-full shadow uppercase tracking-wide">
            Habis
          </span>
        {/if}
      </div>

      {#if item.preparation_time_minutes > 0}
        <div class="absolute bottom-3 right-3 bg-espresso/80 backdrop-blur-sm text-creamWarm text-[10px] font-semibold px-2 py-0.5 rounded-lg flex items-center space-x-1">
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>~{item.preparation_time_minutes} mnt</span>
        </div>
      {/if}
    </div>

    <!-- Content Details -->
    <div class="p-5 space-y-2">
      <div class="flex items-start justify-between gap-2">
        <h3 class="font-heading font-extrabold text-base text-espresso group-hover:text-caramel transition line-clamp-1">
          {item.name}
        </h3>
      </div>
      {#if item.description}
        <p class="text-xs text-neutral-500 leading-relaxed line-clamp-2">
          {item.description}
        </p>
      {/if}
    </div>
  </div>

  <!-- Bottom Price & Order Action -->
  <div class="p-5 pt-0 border-t border-neutral-100 mt-2 flex items-center justify-between">
    <div>
      <span class="text-[10px] text-neutral-400 font-semibold block uppercase tracking-wider">Harga</span>
      <span class="font-heading font-black text-lg text-espresso">
        {formatRupiah(item.base_price)}
      </span>
    </div>

    <button
      onclick={() => onSelect(item)}
      disabled={!item.is_available}
      class="px-4 py-2 rounded-xl text-xs font-bold transition flex items-center space-x-1.5 shadow-sm {
        item.is_available 
          ? 'bg-espresso text-creamWarm hover:bg-caramel hover:text-white active:scale-95' 
          : 'bg-neutral-200 text-neutral-400 cursor-not-allowed'
      }"
    >
      {#if hasOptions}
        <span>Pilih Varian</span>
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      {:else}
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        <span>Tambah</span>
      {/if}
    </button>
  </div>
</div>
