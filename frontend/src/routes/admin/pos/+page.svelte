<script lang="ts">
  import type { PageData, ActionData } from './$types';
  import { formatRupiah } from '$lib/utils/format';
  import type { MenuItem, CartItem, Table } from '$lib/types';
  import ModifierModal from '$lib/components/ModifierModal.svelte';

  let { data, form } = $props<{ data: PageData; form: ActionData }>();

  // Shift & Order State
  let activeShift = $state(data.currentShift);
  let orderType = $state<'dine_in' | 'takeaway'>('dine_in');
  let selectedTableId = $state<string>(data.tables[0]?.id || '');
  let customerName = $state<string>('');
  let selectedCategorySlug = $state<string>('all');
  let searchQuery = $state<string>('');

  // Cart & Modifier
  let selectedMenuItem = $state<MenuItem | null>(null);
  let posCart = $state<CartItem[]>([]);
  let discountAmount = $state<number>(0);

  // Payment Modal State
  let isPaymentModalOpen = $state(false);
  let paymentMethod = $state<'cash' | 'qris' | 'bank_transfer'>('cash');
  let cashTendered = $state<number>(0);
  let referenceNumber = $state<string>('');
  let isProcessingPayment = $state(false);

  // Receipt Modal State
  let completedReceipt = $state<any | null>(null);

  // Shift Modal State
  let isShiftModalOpen = $state(false);
  let initialCashInput = $state<number>(500000);
  let actualCashInput = $state<number>(0);
  let closingNotes = $state<string>('');

  // Filtered menu items
  const filteredItems = $derived.by(() => {
    let list = [...data.items];
    if (selectedCategorySlug !== 'all') {
      const cat = data.categories.find(c => c.slug === selectedCategorySlug);
      if (cat) list = list.filter(i => i.category_id === cat.id);
    }
    if (searchQuery.trim() !== '') {
      const q = searchQuery.toLowerCase().trim();
      list = list.filter(i => i.name.toLowerCase().includes(q));
    }
    return list;
  });

  // Totals Calculation
  const subtotal = $derived(
    posCart.reduce((sum, item) => sum + (item.unit_price * item.quantity), 0)
  );

  const taxableAmount = $derived(
    Math.max(0, subtotal - discountAmount)
  );

  const taxAmount = $derived(
    Math.round(taxableAmount * 0.10) // PB1 10%
  );

  const totalBill = $derived(
    taxableAmount + taxAmount
  );

  const changeDue = $derived(
    Math.max(0, cashTendered - totalBill)
  );

  function openModifier(item: MenuItem) {
    selectedMenuItem = item;
  }

  function handleAddToCart(item: CartItem) {
    posCart = [...posCart, item];
  }

  function updateQty(cartItemId: string, delta: number) {
    posCart = posCart.map(item => {
      if (item.cart_item_id === cartItemId) {
        const newQty = item.quantity + delta;
        return newQty > 0 ? { ...item, quantity: newQty } : null;
      }
      return item;
    }).filter(Boolean) as CartItem[];
  }

  function removeItem(cartItemId: string) {
    posCart = posCart.filter(item => item.cart_item_id !== cartItemId);
  }

  function openPayment() {
    cashTendered = totalBill;
    isPaymentModalOpen = true;
  }

  async function executePayment() {
    if (paymentMethod === 'cash' && cashTendered < totalBill) {
      alert("Jumlah uang tunai kurang dari total tagihan!");
      return;
    }

    isProcessingPayment = true;

    // Payload simulasi
    const orderNumber = "ORD-" + new Date().toISOString().slice(0,10).replace(/-/g,'') + "-" + Math.floor(Math.random()*9000 + 1000);
    const tbl = data.tables.find(t => t.id === selectedTableId);

    completedReceipt = {
      order_number: orderNumber,
      order_type: orderType === 'dine_in' ? `Meja ${tbl?.table_number || '-'}` : 'Takeaway',
      customer_name: customerName || 'Pelanggan Langsung',
      date: new Date().toLocaleDateString('id-ID', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' }),
      items: [...posCart],
      subtotal,
      discount: discountAmount,
      tax: taxAmount,
      total: totalBill,
      payment_method: paymentMethod.toUpperCase(),
      cash_tendered: cashTendered,
      change_due: changeDue,
      cashier_name: 'Budi Santoso (Kasir)'
    };

    isProcessingPayment = false;
    isPaymentModalOpen = false;
    posCart = [];
    discountAmount = 0;
  }

  function printReceipt() {
    window.print();
  }
</script>

<svelte:head>
  <title>Aplikasi Kasir (POS) — KopiRise</title>
</svelte:head>

<div class="space-y-6">
  
  <!-- CASH SHIFT BAR -->
  <div class="bg-white p-4 rounded-2xl border border-neutral-200 shadow-sm flex flex-wrap items-center justify-between gap-4 text-xs">
    <div class="flex items-center space-x-3">
      {#if activeShift}
        <span class="w-3 h-3 rounded-full bg-emerald-500 animate-pulse"></span>
        <span>Kasir: <strong>Budi Santoso</strong></span>
        <span class="text-neutral-300">|</span>
        <span>Shift: <code class="font-mono bg-neutral-100 px-2 py-0.5 rounded font-bold">{activeShift.shift_number || 'SHF-ACTIVE'}</code></span>
        <span class="text-neutral-300">|</span>
        <span>Kas Awal: <strong>{formatRupiah(activeShift.initial_cash || 500000)}</strong></span>
      {:else}
        <span class="w-3 h-3 rounded-full bg-amber-500"></span>
        <span class="font-bold text-amber-700">Shift Kasir Belum Dibuka</span>
      {/if}
    </div>

    <div>
      {#if activeShift}
        <button 
          onclick={() => isShiftModalOpen = true}
          class="px-3.5 py-1.5 rounded-xl bg-red-50 hover:bg-red-100 text-red-700 font-bold border border-red-200 transition"
        >
          Tutup Shift Kasir & Rekonsiliasi
        </button>
      {:else}
        <button 
          onclick={() => isShiftModalOpen = true}
          class="px-3.5 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold transition shadow"
        >
          Buka Shift Kasir Baru
        </button>
      {/if}
    </div>
  </div>

  <!-- MAIN POS INTERFACE -->
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
    
    <!-- LEFT 7 COLS: MENU BROWSER & SELECTOR -->
    <div class="lg:col-span-7 xl:col-span-8 space-y-4">
      
      <!-- Order Mode & Table Selector -->
      <div class="bg-white p-4 rounded-2xl border border-neutral-200 shadow-sm flex flex-wrap items-center justify-between gap-4 text-xs">
        <div class="flex items-center space-x-2">
          <button 
            onclick={() => orderType = 'dine_in'}
            class="px-4 py-2 rounded-xl font-bold transition {orderType === 'dine_in' ? 'bg-espresso text-creamWarm' : 'bg-neutral-100 text-neutral-600'}"
          >
            🍽️ Dine-In (Makan di Tempat)
          </button>
          <button 
            onclick={() => orderType = 'takeaway'}
            class="px-4 py-2 rounded-xl font-bold transition {orderType === 'takeaway' ? 'bg-espresso text-creamWarm' : 'bg-neutral-100 text-neutral-600'}"
          >
            🥡 Takeaway (Bungkus)
          </button>
        </div>

        {#if orderType === 'dine_in'}
          <div class="flex items-center space-x-2">
            <span class="font-bold text-neutral-600">Pilih Meja:</span>
            <select 
              bind:value={selectedTableId}
              class="px-3 py-1.5 rounded-xl border border-neutral-300 font-bold text-xs bg-neutral-50"
            >
              {#each data.tables as tbl}
                <option value={tbl.id}>
                  Meja {tbl.table_number} ({tbl.area_name || 'Area'}) [{tbl.status}]
                </option>
              {/each}
            </select>
          </div>
        {/if}
      </div>

      <!-- Categories & Search -->
      <div class="flex items-center space-x-2 overflow-x-auto pb-1 custom-scrollbar text-xs">
        <button
          onclick={() => selectedCategorySlug = 'all'}
          class="px-3.5 py-1.5 rounded-xl font-bold shrink-0 transition {selectedCategorySlug === 'all' ? 'bg-espresso text-creamWarm shadow' : 'bg-white border text-neutral-600'}"
        >
          Semua
        </button>
        {#each data.categories as c}
          <button
            onclick={() => selectedCategorySlug = c.slug}
            class="px-3.5 py-1.5 rounded-xl font-bold shrink-0 transition {selectedCategorySlug === c.slug ? 'bg-espresso text-creamWarm shadow' : 'bg-white border text-neutral-600'}"
          >
            {c.name}
          </button>
        {/each}
      </div>

      <!-- Menu Items Grid -->
      <div class="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 gap-4">
        {#each filteredItems as item}
          <div 
            onclick={() => openModifier(item)}
            class="bg-white rounded-2xl border border-neutral-200 p-3 shadow-sm hover:border-caramel hover:shadow-md cursor-pointer transition flex flex-col justify-between"
          >
            <div class="h-24 bg-neutral-100 rounded-xl overflow-hidden relative mb-2">
              {#if item.image_url}
                <img src={item.image_url} alt={item.name} class="w-full h-full object-cover" />
              {:else}
                <div class="w-full h-full flex items-center justify-center text-neutral-300 text-2xl">☕</div>
              {/if}
              {#if item.is_featured}
                <span class="absolute top-1.5 left-1.5 bg-caramel text-white text-[9px] font-black px-1.5 py-0.5 rounded shadow">
                  Star
                </span>
              {/if}
            </div>
            <div>
              <h4 class="font-heading font-bold text-xs text-espresso line-clamp-1">{item.name}</h4>
              <span class="font-heading font-black text-xs text-caramel mt-1 block">
                {formatRupiah(item.base_price)}
              </span>
            </div>
          </div>
        {/each}
      </div>

    </div>

    <!-- RIGHT 5 COLS: KASIR CART & BILLING -->
    <div class="lg:col-span-5 xl:col-span-4 bg-white rounded-3xl border border-neutral-200 shadow-sm p-6 space-y-4">
      <div class="flex items-center justify-between pb-3 border-b border-neutral-100">
        <h3 class="font-heading font-black text-base text-espresso">Tagihan Kasir</h3>
        <span class="text-xs bg-neutral-100 text-neutral-600 font-bold px-2 py-0.5 rounded">
          {posCart.length} Item
        </span>
      </div>

      <!-- Customer Name Input -->
      <div>
        <label for="pos-cust-name" class="block text-[10px] text-neutral-400 font-bold uppercase mb-1">Nama Pelanggan</label>
        <input 
          id="pos-cust-name"
          type="text" 
          bind:value={customerName}
          placeholder="Nama pembeli..."
          class="w-full px-3 py-2 rounded-xl border border-neutral-200 text-xs focus:ring-2 focus:ring-caramel focus:outline-none"
        />
      </div>

      <!-- Items List -->
      <div class="max-h-64 overflow-y-auto space-y-3 custom-scrollbar text-xs">
        {#if posCart.length === 0}
          <div class="py-12 text-center text-neutral-400 space-y-1">
            <span class="text-2xl block">🛒</span>
            <p class="font-medium text-xs">Keranjang masih kosong</p>
          </div>
        {:else}
          {#each posCart as item}
            <div class="p-3 bg-neutral-50 rounded-xl border border-neutral-100 space-y-1.5">
              <div class="flex justify-between items-start">
                <span class="font-bold text-neutral-800">{item.name}</span>
                <button onclick={() => removeItem(item.cart_item_id)} class="text-neutral-400 hover:text-red-500">✕</button>
              </div>

              {#if item.modifiers && item.modifiers.length > 0}
                <div class="text-[10px] text-neutral-500 pl-2">
                  {#each item.modifiers as m}
                    <span>+ {m.option_name} </span>
                  {/each}
                </div>
              {/if}

              <div class="flex items-center justify-between pt-1">
                <span class="font-black text-xs text-caramel">{formatRupiah(item.unit_price * item.quantity)}</span>
                <div class="flex items-center space-x-2 bg-white px-2 py-0.5 rounded-lg border border-neutral-200">
                  <button onclick={() => updateQty(item.cart_item_id, -1)} class="font-bold text-xs">-</button>
                  <span class="font-black text-xs px-1">{item.quantity}</span>
                  <button onclick={() => updateQty(item.cart_item_id, 1)} class="font-bold text-xs">+</button>
                </div>
              </div>
            </div>
          {/each}
        {/if}
      </div>

      <!-- Bill Breakdown -->
      {#if posCart.length > 0}
        <div class="pt-3 border-t border-neutral-100 space-y-2 text-xs">
          <div class="flex justify-between text-neutral-500">
            <span>Subtotal</span>
            <span class="font-bold text-neutral-800">{formatRupiah(subtotal)}</span>
          </div>

          <div class="flex justify-between text-neutral-500">
            <span>Pajak Restoran (PB1 10%)</span>
            <span class="font-bold text-neutral-800">{formatRupiah(taxAmount)}</span>
          </div>

          <div class="flex justify-between text-base font-heading font-black text-espresso pt-2 border-t border-neutral-200">
            <span>Total Bayar</span>
            <span class="text-caramel">{formatRupiah(totalBill)}</span>
          </div>

          <button
            onclick={openPayment}
            class="w-full mt-3 py-3.5 rounded-2xl bg-caramel hover:bg-caramelHover text-white font-heading font-extrabold text-xs shadow-lg transition transform active:scale-95 text-center"
          >
            Proses Pembayaran ({formatRupiah(totalBill)})
          </button>
        </div>
      {/if}

    </div>

  </div>

</div>

<!-- PAYMENT MODAL -->
{#if isPaymentModalOpen}
  <div class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="bg-white rounded-3xl max-w-md w-full shadow-2xl overflow-hidden border border-neutral-200 space-y-6 p-6">
      <div class="flex justify-between items-center pb-3 border-b border-neutral-100">
        <h3 class="font-heading font-black text-lg text-espresso">Pilih Metode Pembayaran</h3>
        <button onclick={() => isPaymentModalOpen = false} class="text-neutral-400 hover:text-neutral-600 font-bold">✕</button>
      </div>

      <div class="text-center py-2 bg-neutral-50 rounded-2xl border border-neutral-200">
        <span class="text-xs text-neutral-500 font-medium block">Total Tagihan</span>
        <span class="font-heading font-black text-2xl text-espresso">{formatRupiah(totalBill)}</span>
      </div>

      <!-- Payment Method Tabs -->
      <div class="grid grid-cols-3 gap-2 text-xs">
        <button 
          onclick={() => paymentMethod = 'cash'}
          class="p-3 rounded-xl border text-center font-bold {paymentMethod === 'cash' ? 'border-caramel bg-caramel/10 text-espresso' : 'border-neutral-200 text-neutral-600'}"
        >
          💵 Tunai
        </button>
        <button 
          onclick={() => paymentMethod = 'qris'}
          class="p-3 rounded-xl border text-center font-bold {paymentMethod === 'qris' ? 'border-caramel bg-caramel/10 text-espresso' : 'border-neutral-200 text-neutral-600'}"
        >
          📱 QRIS
        </button>
        <button 
          onclick={() => paymentMethod = 'bank_transfer'}
          class="p-3 rounded-xl border text-center font-bold {paymentMethod === 'bank_transfer' ? 'border-caramel bg-caramel/10 text-espresso' : 'border-neutral-200 text-neutral-600'}"
        >
          🏦 Bank BCA
        </button>
      </div>

      {#if paymentMethod === 'cash'}
        <div class="space-y-3 text-xs">
          <div>
            <label for="cash-tendered" class="block font-bold text-neutral-700 mb-1">Uang Diterima dari Pelanggan</label>
            <input 
              id="cash-tendered"
              type="number" 
              bind:value={cashTendered}
              class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-base font-bold focus:ring-2 focus:ring-caramel focus:outline-none"
            />
          </div>

          <!-- Quick Cash Buttons -->
          <div class="flex gap-2">
            <button onclick={() => cashTendered = totalBill} class="px-2.5 py-1.5 rounded-lg bg-neutral-100 font-bold hover:bg-neutral-200 text-[11px]">
              Uang Pas
            </button>
            <button onclick={() => cashTendered = 50000} class="px-2.5 py-1.5 rounded-lg bg-neutral-100 font-bold hover:bg-neutral-200 text-[11px]">
              50.000
            </button>
            <button onclick={() => cashTendered = 100000} class="px-2.5 py-1.5 rounded-lg bg-neutral-100 font-bold hover:bg-neutral-200 text-[11px]">
              100.000
            </button>
            <button onclick={() => cashTendered = 200000} class="px-2.5 py-1.5 rounded-lg bg-neutral-100 font-bold hover:bg-neutral-200 text-[11px]">
              200.000
            </button>
          </div>

          <div class="p-3 bg-neutral-50 rounded-xl flex justify-between items-center text-sm font-bold">
            <span class="text-neutral-600">Uang Kembalian:</span>
            <span class="font-black text-caramel text-base">{formatRupiah(changeDue)}</span>
          </div>
        </div>
      {:else if paymentMethod === 'qris'}
        <div class="text-center p-4 bg-neutral-50 rounded-2xl border border-neutral-200 space-y-2 text-xs">
          <div class="w-32 h-32 bg-white border border-neutral-300 mx-auto rounded-xl flex items-center justify-center font-bold text-neutral-400">
            [ QRIS MANDIRI ]
          </div>
          <p class="font-bold text-neutral-700">NMID: ID102008988KOPIRISE</p>
          <p class="text-[10px] text-neutral-500">Minta pelanggan scan kode QR melalui BCA, GoPay, OVO, atau ShopeePay.</p>
        </div>
      {:else}
        <div class="p-4 bg-blue-50/60 rounded-2xl border border-blue-200 text-xs space-y-1 text-blue-900">
          <p class="font-bold">BCA Bisnis KopiRise Jember</p>
          <p class="font-mono text-sm font-black">128-099-8877</p>
          <p class="text-[10px] text-blue-700">Pastikan mutasi dana telah masuk sebelum menyelesaikan transaksi.</p>
        </div>
      {/if}

      <button
        onclick={executePayment}
        disabled={isProcessingPayment}
        class="w-full py-3.5 rounded-2xl bg-emerald-600 hover:bg-emerald-700 text-white font-heading font-extrabold text-xs shadow-lg transition"
      >
        Konfirmasi Pembayaran Selesai
      </button>
    </div>
  </div>
{/if}

<!-- THERMAL RECEIPT MODAL (READY FOR 58mm / 80mm PRINTING) -->
{#if completedReceipt}
  <div class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="bg-white rounded-3xl max-w-sm w-full shadow-2xl p-6 space-y-6">
      
      <!-- Printable Area -->
      <div id="receipt-print-area" class="font-mono text-[11px] text-neutral-900 leading-tight space-y-3 p-4 border border-dashed border-neutral-300 rounded-xl bg-neutral-50">
        <div class="text-center space-y-1">
          <h2 class="font-black text-sm uppercase">KopiRise Jember</h2>
          <p class="text-[9px]">Jl. Semeru, Kloncing, Sumbersari, Jember</p>
          <p class="text-[9px]">WA: 0812-3456-7890</p>
        </div>

        <div class="border-t border-b border-dashed border-neutral-400 py-1 space-y-0.5 text-[10px]">
          <div class="flex justify-between">
            <span>No: {completedReceipt.order_number}</span>
            <span>{completedReceipt.date}</span>
          </div>
          <div class="flex justify-between">
            <span>Mode: {completedReceipt.order_type}</span>
            <span>Kasir: {completedReceipt.cashier_name}</span>
          </div>
        </div>

        <div class="space-y-1 text-[10px]">
          {#each completedReceipt.items as item}
            <div class="flex justify-between">
              <span>{item.quantity}x {item.name}</span>
              <span>{formatRupiah(item.unit_price * item.quantity)}</span>
            </div>
          {/each}
        </div>

        <div class="border-t border-dashed border-neutral-400 pt-1 space-y-0.5 text-[10px]">
          <div class="flex justify-between">
            <span>Subtotal</span>
            <span>{formatRupiah(completedReceipt.subtotal)}</span>
          </div>
          <div class="flex justify-between">
            <span>Pajak Restoran PB1 (10%)</span>
            <span>{formatRupiah(completedReceipt.tax)}</span>
          </div>
          <div class="flex justify-between font-black text-xs pt-1 border-t border-neutral-300">
            <span>TOTAL</span>
            <span>{formatRupiah(completedReceipt.total)}</span>
          </div>
          <div class="flex justify-between pt-1">
            <span>Bayar ({completedReceipt.payment_method})</span>
            <span>{formatRupiah(completedReceipt.cash_tendered)}</span>
          </div>
          <div class="flex justify-between">
            <span>Kembalian</span>
            <span>{formatRupiah(completedReceipt.change_due)}</span>
          </div>
        </div>

        <div class="text-center pt-2 border-t border-dashed border-neutral-400 space-y-0.5 text-[9px] text-neutral-500">
          <p class="font-bold text-neutral-800">WiFi: KopiRise_Guest / semerujember</p>
          <p>Terima kasih atas kunjungan Anda!</p>
        </div>
      </div>

      <div class="flex gap-2">
        <button 
          onclick={printReceipt}
          class="flex-1 py-3 bg-espresso hover:bg-espressoDark text-creamWarm rounded-xl font-bold text-xs flex items-center justify-center space-x-1"
        >
          <span>🖨️ Cetak Struk Thermal</span>
        </button>
        <button 
          onclick={() => completedReceipt = null}
          class="px-4 py-3 bg-neutral-100 hover:bg-neutral-200 text-neutral-700 rounded-xl font-bold text-xs"
        >
          Selesai
        </button>
      </div>

    </div>
  </div>
{/if}

<!-- SHIFT RECONCILIATION MODAL -->
{#if isShiftModalOpen}
  <div class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="bg-white rounded-3xl max-w-md w-full shadow-2xl p-6 space-y-6">
      <div class="flex justify-between items-center pb-3 border-b border-neutral-100">
        <h3 class="font-heading font-black text-lg text-espresso">
          {activeShift ? 'Tutup Shift Kasir & Rekonsiliasi' : 'Buka Shift Kasir Baru'}
        </h3>
        <button onclick={() => isShiftModalOpen = false} class="text-neutral-400 font-bold">✕</button>
      </div>

      {#if activeShift}
        <form method="POST" action="?/closeShift" class="space-y-4 text-xs">
          <div class="p-4 bg-neutral-50 rounded-2xl border border-neutral-200 space-y-2">
            <div class="flex justify-between">
              <span>Modal Awal Kas:</span>
              <strong>{formatRupiah(activeShift.initial_cash || 500000)}</strong>
            </div>
            <div class="flex justify-between">
              <span>Penjualan Tunai Shift Ini:</span>
              <strong class="text-emerald-700">{formatRupiah(activeShift.cash_sales || 0)}</strong>
            </div>
            <div class="flex justify-between pt-2 border-t border-neutral-200 font-bold">
              <span>Kas Fisik yang Seharusnya:</span>
              <strong class="text-caramel">{formatRupiah((activeShift.initial_cash || 500000) + (activeShift.cash_sales || 0))}</strong>
            </div>
          </div>

          <div class="space-y-1.5">
            <label for="actual-cash" class="block font-bold text-neutral-700">Hitung Uang Fisik Aktual di Laci (Rp)</label>
            <input 
              id="actual-cash"
              name="actual_cash" 
              type="number" 
              required
              bind:value={actualCashInput}
              class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-sm font-bold"
            />
          </div>

          <div class="space-y-1.5">
            <label for="close-notes" class="block font-bold text-neutral-700">Catatan Rekonsiliasi (Opsional)</label>
            <input 
              id="close-notes"
              name="closing_notes" 
              type="text" 
              bind:value={closingNotes}
              placeholder="Keterangan jika terdapat selisih..."
              class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-xs"
            />
          </div>

          <button 
            type="submit" 
            class="w-full py-3.5 bg-red-600 hover:bg-red-700 text-white rounded-xl font-bold text-xs shadow"
          >
            Selesaikan & Tutup Shift Kasir
          </button>
        </form>
      {:else}
        <form method="POST" action="?/openShift" class="space-y-4 text-xs">
          <div class="space-y-1.5">
            <label for="init-cash" class="block font-bold text-neutral-700">Modal Kas Awal di Laci (Rp)</label>
            <input 
              id="init-cash"
              name="initial_cash" 
              type="number" 
              required
              bind:value={initialCashInput}
              class="w-full px-4 py-2.5 rounded-xl border border-neutral-200 text-sm font-bold"
            />
          </div>

          <button 
            type="submit" 
            class="w-full py-3.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold text-xs shadow"
          >
            Konfirmasi Buka Shift
          </button>
        </form>
      {/if}
    </div>
  </div>
{/if}

<!-- Modals -->
<ModifierModal 
  item={selectedMenuItem} 
  onClose={() => selectedMenuItem = null} 
  onAddToCart={handleAddToCart} 
/>
