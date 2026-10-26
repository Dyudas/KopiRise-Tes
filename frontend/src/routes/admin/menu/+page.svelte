<script lang="ts">
  import type { PageData } from './$types';
  import { formatRupiah } from '$lib/utils/format';

  let { data } = $props<{ data: PageData }>();

  let items = $state(data.items);
  let searchQuery = $state('');

  function toggleLocalAvailability(id: string) {
    items = items.map(i => {
      if (i.id === id) {
        return { ...i, is_available: !i.is_available };
      }
      return i;
    });
  }

  const filteredItems = $derived(
    searchQuery.trim() === '' 
      ? items 
      : items.filter(i => i.name.toLowerCase().includes(searchQuery.toLowerCase()))
  );
</script>

<svelte:head>
  <title>Kelola Menu & Ketersediaan — KopiRise</title>
</svelte:head>

<div class="space-y-6">
  
  <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 bg-white p-6 rounded-3xl border border-neutral-200 shadow-sm">
    <div>
      <h2 class="font-heading font-black text-2xl text-espresso">Manajemen Menu & Ketersediaan</h2>
      <p class="text-xs text-neutral-500 mt-1">
        Kelola 32 item menu, ubah status habis/tersedia, dan pantau varian sajian.
      </p>
    </div>

    <div class="w-full sm:w-72">
      <input 
        type="text" 
        bind:value={searchQuery}
        placeholder="Cari menu..."
        class="w-full px-4 py-2 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
      />
    </div>
  </div>

  <!-- Menu Table -->
  <div class="bg-white rounded-3xl border border-neutral-200 shadow-sm overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead class="bg-neutral-50 border-b border-neutral-200 text-neutral-500 font-bold uppercase tracking-wider text-[10px]">
          <tr>
            <th class="p-4">Foto</th>
            <th class="p-4">Nama Menu</th>
            <th class="p-4">Kategori</th>
            <th class="p-4">Harga Dasar</th>
            <th class="p-4">Status</th>
            <th class="p-4 text-right">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-neutral-100">
          {#each filteredItems as item}
            <tr class="hover:bg-neutral-50/60 transition">
              <td class="p-4">
                <div class="w-10 h-10 rounded-xl bg-neutral-100 overflow-hidden">
                  {#if item.image_url}
                    <img src={item.image_url} alt={item.name} class="w-full h-full object-cover" />
                  {:else}
                    <span class="w-full h-full flex items-center justify-center text-xs">☕</span>
                  {/if}
                </div>
              </td>
              <td class="p-4">
                <span class="font-bold text-neutral-900 block">{item.name}</span>
                {#if item.is_featured}
                  <span class="text-[9px] bg-caramel/20 text-caramel font-bold px-1.5 py-0.5 rounded">Best Seller</span>
                {/if}
              </td>
              <td class="p-4 text-neutral-500">{item.category_name || 'Kategori'}</td>
              <td class="p-4 font-black text-espresso">{formatRupiah(item.base_price)}</td>
              <td class="p-4">
                <span class="px-2.5 py-1 rounded-full text-[10px] font-bold uppercase {
                  item.is_available ? 'bg-emerald-100 text-emerald-800' : 'bg-red-100 text-red-800'
                }">
                  {item.is_available ? 'Tersedia' : 'Habis'}
                </span>
              </td>
              <td class="p-4 text-right">
                <button
                  type="button"
                  onclick={() => toggleLocalAvailability(item.id)}
                  class="px-3 py-1.5 rounded-xl text-xs font-bold transition {
                    item.is_available 
                      ? 'bg-red-50 hover:bg-red-100 text-red-700 border border-red-200' 
                      : 'bg-emerald-50 hover:bg-emerald-100 text-emerald-700 border border-emerald-200'
                  }"
                >
                  {item.is_available ? 'Set Habis' : 'Set Tersedia'}
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>

</div>
