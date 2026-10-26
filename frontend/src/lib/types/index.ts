export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  display_order: number;
  is_active: boolean;
}

export interface ModifierOption {
  id: string;
  group_id: string;
  name: string;
  additional_price: number; // INTEGER Rupiah
  is_default: boolean;
  is_available: boolean;
}

export interface ModifierGroup {
  id: string;
  name: string;
  selection_type: 'single' | 'multiple';
  min_selection: number;
  max_selection: number;
  is_required: boolean;
  options: ModifierOption[];
}

export interface MenuVariant {
  id: string;
  menu_item_id: string;
  name: string;
  price_difference: number; // INTEGER Rupiah
  is_available: boolean;
}

export interface MenuItem {
  id: string;
  category_id: string;
  category_name?: string;
  name: string;
  slug: string;
  description?: string;
  base_price: number; // INTEGER Rupiah
  image_url?: string;
  is_available: boolean;
  is_featured: boolean;
  preparation_time_minutes: number;
  variants?: MenuVariant[];
  modifier_groups?: ModifierGroup[];
}

export interface Area {
  id: string;
  name: string;
  description?: string;
  is_smoking: boolean;
}

export interface Table {
  id: string;
  area_id: string;
  area_name?: string;
  table_number: string;
  capacity: number;
  status: 'available' | 'occupied' | 'reserved';
  qr_token?: string;
}

export interface Reservation {
  id: string;
  reservation_number: string;
  customer_name: string;
  customer_phone: string;
  customer_email?: string;
  table_id?: string;
  table_number?: string;
  area_name?: string;
  reservation_date: string;
  reservation_time: string;
  guest_count: number;
  duration_minutes: number;
  status: 'pending' | 'confirmed' | 'seated' | 'completed' | 'cancelled' | 'no_show';
  notes?: string;
  created_at: string;
}

export interface CartModifier {
  option_id: string;
  group_name: string;
  option_name: string;
  price: number;
}

export interface CartItem {
  cart_item_id: string; // Unique for cart line
  menu_item_id: string;
  name: string;
  image_url?: string;
  variant_id?: string;
  variant_name?: string;
  unit_price: number;
  quantity: number;
  notes?: string;
  modifiers: CartModifier[];
  item_total: number;
}

export interface CreateOrderRequest {
  order_type: 'dine_in' | 'takeaway' | 'reservation';
  table_id?: string;
  customer_name?: string;
  customer_phone?: string;
  notes?: string;
  idempotency_key: string;
  items: {
    menu_item_id: string;
    variant_id?: string;
    quantity: number;
    notes?: string;
    modifier_option_ids?: string[];
  }[];
}

export interface ApiResponse<T> {
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  meta?: {
    current_page: number;
    per_page: number;
    total_items: number;
    total_pages: number;
  };
}
