<script lang="ts">
  import type { PageData } from './$types';
  import { enhance } from '$app/forms';

  let { data } = $props<{ data: PageData }>();

  // Mock KDS items jika backend belum memiliki data aktif
  const defaultKds = [
    {
      id: 'ord-kds-01',
      order_number: 'ORD-20261026-0001',
      table_number: 'IN-03',
      order_type: 'dine_in',
      customer_name: 'Dewi Lestari',
      elapsed: '6 mnt lalu',
      status: 'preparing',
      items: [
        { id: 'item-1', name: 'Kopi Susu Aren Semeru', qty: 2, notes: 'Less sugar, less ice', status: 'preparing' },
        { id: 'item-2', name: 'Tahu Walik Krispi Jember', qty: 1, notes: 'Petis dipisah', status: 'ready' },
        { id: 'item-3', name: 'Nasi Goreng Spesial KopiRise', qty: 1, notes: 'Telur ceplok matang', status: 'preparing' }
      ]
    },
    {
      id: 'ord-kds-02',
      order_number: 'ORD-20261026-0002',
      table_number: 'OUT-04',
      order_type: 'dine_in',
      customer_name: 'Budi (QR Table)',
      elapsed: '2 mnt lalu',
      status: 'pending',
      items: [
        { id: 'item-4', name: 'V60 Ijen Raung Single Origin', qty: 1, notes: 'Air 92C pour over', status: 'pending' },
        { id: 'item-5', name: 'Pain Au Chocolat', qty: 1, notes: 'Hangatkan 1 menit', status: 'pending' }
      ]
    }
  ];

  let kdsOrders = $state(data.orders && data.orders.length > 0 ? data.orders : defaultKds);
  let filterStatus = $state<string>('all');

  function updateLocalStatus(orderId: string, newStatus: string) {
    kdsOrders = kdsOrders.map(o => {
      if (o.id === orderId) {
        return { ...o, status: newStatus };
      }
      return o;
    }).filter(o => o.status !== 'served');
  }

  const filteredOrders = $derived(
    filterStatus === 'all' 
      ? kdsOrders 
      : kdsOrders.filter(o => o.status === filterStatus)
  );

  const preparingCount = $derived(kdsOrders.filter(o => o.status === 'preparing').length);
  const readyCount = $derived(kdsOrders.filter(o => o.status === 'ready').length);
</script>

<svelte:head>
  <title>Kitchen Display System (KDS) — KopiRise</title>
</svelte:head>

<div class="space-y-6">
  
  <!-- KDS TOP STATUS BAR -->
  <div class="bg-espresso text-creamWarm p-6 rounded-3xl shadow-xl flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
    <div>
      <div class="flex items-center space-x-2">
        <span class="w-3 h-3 rounded-full bg-caramel animate-ping"></span>
        <h2 class="font-heading font-black text-2xl text-creamWarm">Antrean Dapur & Barista (KDS)</h2>
      </div>
      <p class="text-xs text-creamWarm/70 mt-1">
        Tiket pesanan real-time dari POS kasir dan pemesanan mandiri QR code meja.
      </p>
    </div>

    <!-- Counters -->
    <div class="flex items-center space-x-3 text-xs">
      <div class="px-4 py-2 rounded-2xl bg-amber-500/20 text-amber-300 border border-amber-500/30 font-bold">
        Sedang Diracik: <strong>{preparingCount}</strong>
      </div>
      <div class="px-4 py-2 rounded-2xl bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 font-bold">
        Siap Saji: <strong>{readyCount}</strong>
      </div>
    </div>
  </div>

  <!-- Filter Pills -->
  <div class="flex items-center space-x-2 text-xs">
    <button 
      onclick={() => filterStatus = 'all'}
      class="px-3.5 py-1.5 rounded-xl font-bold {filterStatus === 'all' ? 'bg-espresso text-creamWarm' : 'bg-white border text-neutral-600'}"
    >
      Semua Antrean ({kdsOrders.length})
    </button>
    <button 
      onclick={() => filterStatus = 'pending'}
      class="px-3.5 py-1.5 rounded-xl font-bold {filterStatus === 'pending' ? 'bg-espresso text-creamWarm' : 'bg-white border text-neutral-600'}"
    >
      Belum Diproses
    </button>
    <button 
      onclick={() => filterStatus = 'preparing'}
      class="px-3.5 py-1.5 rounded-xl font-bold {filterStatus === 'preparing' ? 'bg-espresso text-creamWarm' : 'bg-white border text-neutral-600'}"
    >
      Sedang Diracik ({preparingCount})
    </button>
    <button 
      onclick={() => filterStatus = 'ready'}
      class="px-3.5 py-1.5 rounded-xl font-bold {filterStatus === 'ready' ? 'bg-espresso text-creamWarm' : 'bg-white border text-neutral-600'}"
    >
      Siap Disajikan ({readyCount})
    </button>
  </div>

  <!-- KDS TICKETS GRID -->
  {#if filteredOrders.length === 0}
    <div class="text-center py-20 bg-white rounded-3xl border border-neutral-200 p-8 space-y-2">
      <span class="text-3xl block">✨</span>
      <h3 class="font-heading font-black text-lg text-espresso">Semua Antrean Bersih!</h3>
      <p class="text-xs text-neutral-400">Tidak ada tiket pesanan aktif yang menunggu saat ini.</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
      {#each filteredOrders as order}
        <div class="bg-white rounded-3xl border-2 shadow-sm overflow-hidden flex flex-col justify-between {
          order.status === 'ready' ? 'border-emerald-500' :
          order.status === 'preparing' ? 'border-amber-400' : 'border-neutral-200'
        }">
          
          <!-- Ticket Header -->
          <div class="p-4 bg-neutral-50 border-b border-neutral-200 flex justify-between items-start">
            <div>
              <div class="flex items-center space-x-2">
                <span class="font-heading font-black text-lg text-espresso">
                  {order.table_number ? `Meja ${order.table_number}` : 'Takeaway'}
                </span>
                <span class="text-[10px] font-bold px-2 py-0.5 rounded-full uppercase {
                  order.status === 'ready' ? 'bg-emerald-100 text-emerald-800' :
                  order.status === 'preparing' ? 'bg-amber-100 text-amber-800' : 'bg-neutral-200 text-neutral-700'
                }">
                  {order.status}
                </span>
              </div>
              <p class="text-[11px] text-neutral-500 font-mono mt-0.5">{order.order_number}</p>
            </div>

            <span class="text-[10px] bg-neutral-200 text-neutral-700 font-bold px-2 py-0.5 rounded-md">
              ⏱ {order.elapsed || 'Baru saja'}
            </span>
          </div>

          <!-- Items List -->
          <div class="p-4 space-y-3 flex-1 text-xs">
            {#each order.items as item}
              <div class="p-2.5 rounded-xl bg-neutral-50 border border-neutral-100 space-y-1">
                <div class="flex justify-between items-start">
                  <span class="font-bold text-neutral-900">
                    <strong class="text-caramel font-black mr-1">{item.qty}x</strong> {item.name}
                  </span>
                </div>

                {#if item.notes}
                  <p class="text-[10px] text-amber-800 font-medium bg-amber-50 px-2 py-0.5 rounded border border-amber-200">
                    Catatan: {item.notes}
                  </p>
                {/if}
              </div>
            {/each}
          </div>

          <!-- Ticket Action Buttons -->
          <div class="p-4 bg-neutral-50 border-t border-neutral-200 flex gap-2">
            {#if order.status === 'pending'}
              <button 
                onclick={() => updateLocalStatus(order.id, 'preparing')}
                class="flex-1 py-2.5 bg-amber-500 hover:bg-amber-600 text-white rounded-xl font-bold text-xs shadow transition"
              >
                Mulai Diracik
              </button>
            {:else if order.status === 'preparing'}
              <button 
                onclick={() => updateLocalStatus(order.id, 'ready')}
                class="flex-1 py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold text-xs shadow transition"
              >
                Tandai Siap Saji
              </button>
            {:else if order.status === 'ready'}
              <button 
                onclick={() => updateLocalStatus(order.id, 'served')}
                class="flex-1 py-2.5 bg-espresso hover:bg-espressoDark text-creamWarm rounded-xl font-bold text-xs shadow transition"
              >
                Sudah Diantar ke Tamu ✓
              </button>
            {/if}
          </div>

        </div>
      {/each}
    </div>
  {/if}

</div>
