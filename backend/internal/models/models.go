package models

import "time"

// Konstanta Peran Sistem (RBAC)
const (
	RoleAdmin    = "admin"
	RoleManager  = "manager"
	RoleCashier  = "cashier"
	RoleBarista  = "barista"
	RoleWaiter   = "waiter"
	RoleCustomer = "customer"
)

// User merepresentasikan entitas akun pengguna sistem.
type User struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Email        string           `json:"email"`
	Phone        string           `json:"phone,omitempty"`
	PasswordHash string           `json:"-"` // Tidak boleh pernah diekspos ke JSON
	Status       string           `json:"status"` // 'active', 'inactive', 'suspended'
	Roles        []string         `json:"roles"`
	Profile      *EmployeeProfile `json:"profile,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	DeletedAt    *time.Time       `json:"deleted_at,omitempty"`
}

// HasRole memeriksa apakah pengguna memiliki peran tertentu.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole memeriksa apakah pengguna memiliki salah satu peran dari daftar yang diberikan.
func (u *User) HasAnyRole(roles ...string) bool {
	for _, requiredRole := range roles {
		if u.HasRole(requiredRole) {
			return true
		}
	}
	return false
}

// Session merepresentasikan sesi aktif yang tercatat di database.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// EmployeeProfile merepresentasikan data kepegawaian internal.
type EmployeeProfile struct {
	UserID       string    `json:"user_id"`
	EmployeeCode string    `json:"employee_code"`
	IDCardNumber string    `json:"id_card_number,omitempty"`
	DateOfBirth  string    `json:"date_of_birth,omitempty"`
	JoinedDate   string    `json:"joined_date"`
	HourlyRate   int64     `json:"hourly_rate"` // INTEGER Rupiah
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ActivityLog mencatat jejak audit mutasi penting sistem.
type ActivityLog struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"user_id,omitempty"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   *string   `json:"entity_id,omitempty"`
	Details    string    `json:"details,omitempty"` // JSON string
	IPAddress  string    `json:"ip_address,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Setting merepresentasikan konfigurasi kafe yang disimpan di database.
type Setting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Category    string    `json:"category"`
	Description string    `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- DTO Request & Response ---

// LoginRequest request payload untuk login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse response payload login.
type LoginResponse struct {
	User      *User     `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateUserRequest request pembuatan akun baru oleh admin.
type CreateUserRequest struct {
	Name         string   `json:"name" validate:"required,min=2,max=100"`
	Email        string   `json:"email" validate:"required,email"`
	Phone        string   `json:"phone" validate:"omitempty,min=8,max=20"`
	Password     string   `json:"password" validate:"required,min=8"`
	Roles        []string `json:"roles" validate:"required,min=1"`
	EmployeeCode string   `json:"employee_code" validate:"omitempty"`
	HourlyRate   int64    `json:"hourly_rate" validate:"gte=0"`
	JoinedDate   string   `json:"joined_date" validate:"omitempty"`
	Notes        string   `json:"notes" validate:"omitempty"`
}

// UpdateUserRequest request update data pengguna.
type UpdateUserRequest struct {
	Name         string    `json:"name" validate:"required,min=2,max=100"`
	Phone        string    `json:"phone" validate:"omitempty,min=8,max=20"`
	Status       string    `json:"status" validate:"required,oneof=active inactive suspended"`
	Roles        []string  `json:"roles" validate:"omitempty"`
	HourlyRate   *int64    `json:"hourly_rate" validate:"omitempty,gte=0"`
	Notes        *string   `json:"notes" validate:"omitempty"`
}

// ResetPasswordRequest request reset password.
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// APIResponse struktur response standar konsisten seluruh endpoint.
type APIResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data,omitempty"`
	Error   *APIErrorDetail `json:"error,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

// APIErrorDetail detail error untuk klien.
type APIErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginationMeta metadata paginasi server-side.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// ContextKey tipe khusus kunci konteks.
type ContextKey string

const (
	ContextKeyUser      ContextKey = "authenticated_user"
	ContextKeyRequestID ContextKey = "request_id"
)

// --- MODEL MENU & MODIFIER ---

type MenuCategory struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MenuItem struct {
	ID           string           `json:"id"`
	CategoryID   string           `json:"category_id"`
	CategoryName string           `json:"category_name,omitempty"`
	Name         string           `json:"name"`
	Slug         string           `json:"slug"`
	Description  string           `json:"description,omitempty"`
	BasePrice    int64            `json:"base_price"` // INTEGER Rupiah
	IsAvailable  bool             `json:"is_available"`
	IsBestSeller bool             `json:"is_best_seller"`
	IsFeatured   bool             `json:"is_featured"`
	ImageURL     string           `json:"image_url,omitempty"`
	SortOrder    int              `json:"sort_order"`
	Variants     []*MenuVariant   `json:"variants,omitempty"`
	Modifiers    []*ModifierGroup `json:"modifiers,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type MenuVariant struct {
	ID              string `json:"id"`
	MenuItemID      string `json:"menu_item_id"`
	Name            string `json:"name"`
	PriceAdjustment int64  `json:"price_adjustment"` // INTEGER Rupiah
	IsAvailable     bool   `json:"is_available"`
	SortOrder       int    `json:"sort_order"`
}

type ModifierGroup struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	IsRequired bool              `json:"is_required"`
	MinSelect  int               `json:"min_select"`
	MaxSelect  int               `json:"max_select"`
	Options    []*ModifierOption `json:"options,omitempty"`
}

type ModifierOption struct {
	ID              string `json:"id"`
	ModifierGroupID string `json:"modifier_group_id"`
	Name            string `json:"name"`
	AdditionalPrice int64  `json:"additional_price"` // INTEGER Rupiah
	IsAvailable     bool   `json:"is_available"`
	SortOrder       int    `json:"sort_order"`
}

type MenuPhoto struct {
	ID           string    `json:"id"`
	MenuItemID   string    `json:"menu_item_id"`
	PhotoURL     string    `json:"photo_url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	IsPrimary    bool      `json:"is_primary"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// Request DTOs Menu
type CreateCategoryRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=100"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=100"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"`
}

type CreateMenuItemRequest struct {
	CategoryID       string            `json:"category_id" validate:"required"`
	Name             string            `json:"name" validate:"required,min=2,max=150"`
	Description      string            `json:"description"`
	BasePrice        int64             `json:"base_price" validate:"gte=0"`
	IsAvailable      bool              `json:"is_available"`
	IsBestSeller     bool              `json:"is_best_seller"`
	IsFeatured       bool              `json:"is_featured"`
	ImageURL         string            `json:"image_url"`
	SortOrder        int               `json:"sort_order"`
	Variants         []CreateVariant   `json:"variants,omitempty"`
	ModifierGroupIDs []string          `json:"modifier_group_ids,omitempty"`
}

type CreateVariant struct {
	Name            string `json:"name" validate:"required"`
	PriceAdjustment int64  `json:"price_adjustment"`
	IsAvailable     bool   `json:"is_available"`
	SortOrder       int    `json:"sort_order"`
}

type UpdateMenuItemRequest struct {
	CategoryID   string `json:"category_id" validate:"required"`
	Name         string `json:"name" validate:"required,min=2,max=150"`
	Description  string `json:"description"`
	BasePrice    int64  `json:"base_price" validate:"gte=0"`
	IsAvailable  bool   `json:"is_available"`
	IsBestSeller bool   `json:"is_best_seller"`
	IsFeatured   bool   `json:"is_featured"`
	ImageURL     string `json:"image_url"`
	SortOrder    int    `json:"sort_order"`
}

// --- MODEL MEJA & QR TOKEN ---

type CafeArea struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order"`
}

type CafeTable struct {
	ID          string    `json:"id"`
	AreaID      string    `json:"area_id"`
	AreaName    string    `json:"area_name,omitempty"`
	TableNumber string    `json:"table_number"`
	Capacity    int       `json:"capacity"`
	Status      string    `json:"status"` // 'available', 'occupied', 'reserved', 'maintenance'
	IsActive    bool      `json:"is_active"`
	ActiveToken string    `json:"active_token,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TableQRToken struct {
	ID        string     `json:"id"`
	TableID   string     `json:"table_id"`
	Token     string     `json:"token"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type CreateTableRequest struct {
	AreaID      string `json:"area_id" validate:"required"`
	TableNumber string `json:"table_number" validate:"required,min=1,max=20"`
	Capacity    int    `json:"capacity" validate:"required,gt=0"`
}

type UpdateTableRequest struct {
	AreaID      string `json:"area_id" validate:"required"`
	TableNumber string `json:"table_number" validate:"required,min=1,max=20"`
	Capacity    int    `json:"capacity" validate:"required,gt=0"`
	Status      string `json:"status" validate:"required,oneof=available occupied reserved maintenance"`
	IsActive    bool   `json:"is_active"`
}

type PublicTableQRResponse struct {
	TableNumber string `json:"table_number"`
	AreaName    string `json:"area_name"`
	Capacity    int    `json:"capacity"`
	Status      string `json:"status"`
	Token       string `json:"token"`
}

// --- MODEL RESERVASI ---

type Reservation struct {
	ID                string                      `json:"id"`
	ReservationNumber string                      `json:"reservation_number"`
	CustomerName      string                      `json:"customer_name"`
	CustomerPhone     string                      `json:"customer_phone"`
	CustomerEmail     string                      `json:"customer_email,omitempty"`
	GuestCount        int                         `json:"guest_count"`
	ReservationDate   string                      `json:"reservation_date"` // YYYY-MM-DD
	ReservationTime   string                      `json:"reservation_time"` // HH:MM:SS
	DurationMinutes   int                         `json:"duration_minutes"`
	AreaID            *string                     `json:"area_id,omitempty"`
	AreaName          string                      `json:"area_name,omitempty"`
	TableID           *string                     `json:"table_id,omitempty"`
	TableNumber       string                      `json:"table_number,omitempty"`
	Status            string                      `json:"status"` // 'pending', 'confirmed', 'seated', 'completed', 'cancelled', 'no_show'
	Notes             string                      `json:"notes,omitempty"`
	History           []*ReservationStatusHistory `json:"history,omitempty"`
	CreatedAt         time.Time                   `json:"created_at"`
	UpdatedAt         time.Time                   `json:"updated_at"`
}

type ReservationStatusHistory struct {
	ID                string    `json:"id"`
	ReservationID     string    `json:"reservation_id"`
	FromStatus        string    `json:"from_status,omitempty"`
	ToStatus          string    `json:"to_status"`
	Notes             string    `json:"notes,omitempty"`
	ChangedByUserID   *string   `json:"changed_by_user_id,omitempty"`
	ChangedByUserName string    `json:"changed_by_user_name,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type CreateReservationRequest struct {
	CustomerName    string  `json:"customer_name" validate:"required,min=2,max=100"`
	CustomerPhone   string  `json:"customer_phone" validate:"required,min=8,max=20"`
	CustomerEmail   string  `json:"customer_email" validate:"omitempty,email"`
	GuestCount      int     `json:"guest_count" validate:"required,gt=0"`
	ReservationDate string  `json:"reservation_date" validate:"required"` // YYYY-MM-DD
	ReservationTime string  `json:"reservation_time" validate:"required"` // HH:MM
	AreaID          *string `json:"area_id,omitempty"`
	TableID         *string `json:"table_id,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	WebsiteTrap     string  `json:"website_trap,omitempty"` // Honeypot anti-spam
}

type UpdateReservationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed seated completed cancelled no_show"`
	Notes  string `json:"notes,omitempty"`
}

// --- MODEL PEMESANAN (ORDER) & KASIR ---

type Order struct {
	ID                  string       `json:"id"`
	OrderNumber         string       `json:"order_number"`
	IdempotencyKey      *string      `json:"idempotency_key,omitempty"`
	OrderType           string       `json:"order_type"` // 'dine_in', 'takeaway', 'reservation'
	TableID             *string      `json:"table_id,omitempty"`
	TableNumber         string       `json:"table_number,omitempty"`
	CustomerName        string       `json:"customer_name,omitempty"`
	CustomerPhone       string       `json:"customer_phone,omitempty"`
	Status              string       `json:"status"` // 'pending', 'confirmed', 'preparing', 'ready', 'served', 'completed', 'cancelled'
	PaymentStatus       string       `json:"payment_status"` // 'unpaid', 'partially_paid', 'paid', 'refunded', 'partially_refunded'
	Subtotal            int64        `json:"subtotal"`
	DiscountAmount      int64        `json:"discount_amount"`
	TaxAmount           int64        `json:"tax_amount"`
	ServiceChargeAmount int64        `json:"service_charge_amount"`
	TotalAmount         int64        `json:"total_amount"`
	PaidAmount          int64        `json:"paid_amount"`
	PromoID             *string      `json:"promo_id,omitempty"`
	PromoCode           string       `json:"promo_code,omitempty"`
	CreatedByUserID     *string      `json:"created_by_user_id,omitempty"`
	CashShiftID         *string      `json:"cash_shift_id,omitempty"`
	Notes               string       `json:"notes,omitempty"`
	Items               []*OrderItem `json:"items,omitempty"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

type OrderItem struct {
	ID                    string               `json:"id"`
	OrderID               string               `json:"order_id"`
	MenuItemID            *string              `json:"menu_item_id,omitempty"`
	MenuVariantID         *string              `json:"menu_variant_id,omitempty"`
	ItemNameSnapshot      string               `json:"item_name_snapshot"`
	VariantNameSnapshot   string               `json:"variant_name_snapshot,omitempty"`
	UnitBasePriceSnapshot int64                `json:"unit_base_price_snapshot"`
	Quantity              int                  `json:"quantity"`
	Subtotal              int64                `json:"subtotal"`
	Notes                 string               `json:"notes,omitempty"`
	Status                string               `json:"status"` // 'pending', 'preparing', 'ready', 'served', 'cancelled'
	Modifiers             []*OrderItemModifier `json:"modifiers,omitempty"`
	CreatedAt             time.Time            `json:"created_at"`
}

type OrderItemModifier struct {
	ID                       string `json:"id"`
	OrderItemID              string `json:"order_item_id"`
	ModifierOptionID         *string `json:"modifier_option_id,omitempty"`
	GroupNameSnapshot        string `json:"group_name_snapshot"`
	OptionNameSnapshot       string `json:"option_name_snapshot"`
	AdditionalPriceSnapshot  int64  `json:"additional_price_snapshot"`
}

type CreateOrderRequest struct {
	IdempotencyKey string                   `json:"idempotency_key" validate:"required"`
	OrderType      string                   `json:"order_type" validate:"required,oneof=dine_in takeaway reservation"`
	TableID        *string                  `json:"table_id,omitempty"`
	CustomerName   string                   `json:"customer_name,omitempty"`
	CustomerPhone  string                   `json:"customer_phone,omitempty"`
	PromoCode      string                   `json:"promo_code,omitempty"`
	CashShiftID    *string                  `json:"cash_shift_id,omitempty"`
	Notes          string                   `json:"notes,omitempty"`
	Items          []CreateOrderItemRequest `json:"items" validate:"required,min=1"`
}

type CreateOrderItemRequest struct {
	MenuItemID        string   `json:"menu_item_id" validate:"required"`
	MenuVariantID     *string  `json:"menu_variant_id,omitempty"`
	Quantity          int      `json:"quantity" validate:"required,gt=0"`
	Notes             string   `json:"notes,omitempty"`
	ModifierOptionIDs []string `json:"modifier_option_ids,omitempty"`
}

// --- MODEL PEMBAYARAN (PAYMENT) & REFUND ---

type Payment struct {
	ID                 string    `json:"id"`
	OrderID            string    `json:"order_id"`
	CashShiftID        *string   `json:"cash_shift_id,omitempty"`
	PaymentNumber      string    `json:"payment_number"`
	IdempotencyKey     *string   `json:"idempotency_key,omitempty"`
	PaymentMethod      string    `json:"payment_method"` // 'cash', 'qris', 'bank_transfer', 'other'
	Amount             int64     `json:"amount"`
	CashTendered       *int64    `json:"cash_tendered,omitempty"`
	ChangeDue          *int64    `json:"change_due,omitempty"`
	ReferenceNumber    string    `json:"reference_number,omitempty"`
	ProcessedByUserID  string    `json:"processed_by_user_id"`
	ProcessedByName    string    `json:"processed_by_name,omitempty"`
	FinancialAccountID string    `json:"financial_account_id"`
	Status             string    `json:"status"` // 'completed', 'voided', 'refunded'
	Notes              string    `json:"notes,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type CreatePaymentRequest struct {
	IdempotencyKey     string  `json:"idempotency_key" validate:"required"`
	OrderID            string  `json:"order_id" validate:"required"`
	PaymentMethod      string  `json:"payment_method" validate:"required,oneof=cash qris bank_transfer other"`
	Amount             int64   `json:"amount" validate:"required,gt=0"`
	CashTendered       *int64  `json:"cash_tendered,omitempty"`
	ReferenceNumber    string  `json:"reference_number,omitempty"`
	FinancialAccountID string  `json:"financial_account_id" validate:"required"`
	Notes              string  `json:"notes,omitempty"`
}

type Refund struct {
	ID                 string    `json:"id"`
	RefundNumber       string    `json:"refund_number"`
	PaymentID          string    `json:"payment_id"`
	OrderID            string    `json:"order_id"`
	CashShiftID        *string   `json:"cash_shift_id,omitempty"`
	Amount             int64     `json:"amount"`
	Reason             string    `json:"reason"`
	RefundMethod       string    `json:"refund_method"`
	FinancialAccountID string    `json:"financial_account_id"`
	ApprovedByUserID   string    `json:"approved_by_user_id"`
	ApprovedByName     string    `json:"approved_by_name,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type CreateRefundRequest struct {
	PaymentID          string `json:"payment_id" validate:"required"`
	OrderID            string `json:"order_id" validate:"required"`
	Amount             int64  `json:"amount" validate:"required,gt=0"`
	Reason             string `json:"reason" validate:"required,min=5"`
	RefundMethod       string `json:"refund_method" validate:"required,oneof=cash qris bank_transfer other"`
	FinancialAccountID string `json:"financial_account_id" validate:"required"`
}

// --- MODEL SHIFT KAS (CASH SHIFT) ---

type CashShift struct {
	ID            string          `json:"id"`
	CashierUserID string          `json:"cashier_user_id"`
	CashierName   string          `json:"cashier_name,omitempty"`
	ShiftNumber   string          `json:"shift_number"`
	OpenedAt      time.Time       `json:"opened_at"`
	ClosedAt      *time.Time      `json:"closed_at,omitempty"`
	InitialCash   int64           `json:"initial_cash"`
	CashSales     int64           `json:"cash_sales"`
	CashIn        int64           `json:"cash_in"`
	CashOut       int64           `json:"cash_out"`
	CashRefunds   int64           `json:"cash_refunds"`
	ExpectedCash  int64           `json:"expected_cash"`
	ActualCash    *int64          `json:"actual_cash,omitempty"`
	Discrepancy   *int64          `json:"discrepancy,omitempty"`
	ClosingNotes  string          `json:"closing_notes,omitempty"`
	Status        string          `json:"status"` // 'open', 'closed'
	Movements     []*CashMovement `json:"movements,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type OpenShiftRequest struct {
	InitialCash int64 `json:"initial_cash" validate:"gte=0"`
}

type CloseShiftRequest struct {
	ActualCash   int64  `json:"actual_cash" validate:"gte=0"`
	ClosingNotes string `json:"closing_notes,omitempty"`
}

type CashMovement struct {
	ID                 string    `json:"id"`
	CashShiftID        string    `json:"cash_shift_id"`
	Type               string    `json:"type"` // 'cash_in', 'cash_out'
	Amount             int64     `json:"amount"`
	Reason             string    `json:"reason"`
	AuthorizedByUserID *string   `json:"authorized_by_user_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type CreateCashMovementRequest struct {
	Type   string `json:"type" validate:"required,oneof=cash_in cash_out"`
	Amount int64  `json:"amount" validate:"required,gt=0"`
	Reason string `json:"reason" validate:"required,min=3"`
}

// ==========================================
// TAHAP 5: KEUANGAN, BIAYA, & AKUN
// ==========================================

// FinancialAccount merepresentasikan bagan akun keuangan (Kas POS, Brankas, Rekening Bank, QRIS).
type FinancialAccount struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	AccountType    string    `json:"account_type"` // 'cash', 'bank', 'electronic', 'other'
	AccountNumber  string    `json:"account_number,omitempty"`
	CurrentBalance int64     `json:"current_balance"` // INTEGER Rupiah
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AccountTransfer perpindahan dana antar rekening/kas (misal: setor kasir ke brankas/bank).
type AccountTransfer struct {
	ID                 string    `json:"id"`
	TransferNumber     string    `json:"transfer_number"`
	FromAccountID      string    `json:"from_account_id"`
	FromAccountName    string    `json:"from_account_name,omitempty"`
	ToAccountID        string    `json:"to_account_id"`
	ToAccountName      string    `json:"to_account_name,omitempty"`
	Amount             int64     `json:"amount"` // INTEGER Rupiah
	TransferDate       string    `json:"transfer_date"`
	Description        string    `json:"description,omitempty"`
	CreatedByUserID    string    `json:"created_by_user_id"`
	CreatedByUserName  string    `json:"created_by_user_name,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type CreateTransferRequest struct {
	FromAccountID string `json:"from_account_id" validate:"required"`
	ToAccountID   string `json:"to_account_id" validate:"required"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
	TransferDate  string `json:"transfer_date" validate:"required"` // YYYY-MM-DD
	Description   string `json:"description" validate:"required,min=3"`
}

// ExpenseCategory kategori operasional kafe (Listrik, Bahan, Gaji, Sewa, Promosi).
type ExpenseCategory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Expense pencatatan biaya operasional riil.
type Expense struct {
	ID                  string               `json:"id"`
	ExpenseNumber       string               `json:"expense_number"`
	ExpenseDate         string               `json:"expense_date"`
	CategoryID          string               `json:"category_id"`
	CategoryName        string               `json:"category_name,omitempty"`
	FinancialAccountID  string               `json:"financial_account_id"`
	AccountName         string               `json:"account_name,omitempty"`
	Amount              int64                `json:"amount"` // INTEGER Rupiah
	Description         string               `json:"description"`
	Vendor              string               `json:"vendor,omitempty"`
	ReferenceNumber     string               `json:"reference_number,omitempty"`
	Status              string               `json:"status"` // 'pending', 'approved', 'rejected', 'voided'
	CreatedByUserID     string               `json:"created_by_user_id"`
	CreatedByUserName   string               `json:"created_by_user_name,omitempty"`
	ApprovedByUserID    *string              `json:"approved_by_user_id,omitempty"`
	ApprovedByUserName  *string              `json:"approved_by_user_name,omitempty"`
	Notes               string               `json:"notes,omitempty"`
	Attachments         []*ExpenseAttachment `json:"attachments,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

type ExpenseAttachment struct {
	ID        string    `json:"id"`
	ExpenseID string    `json:"expense_id"`
	FilePath  string    `json:"file_path"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateExpenseRequest struct {
	ExpenseDate        string `json:"expense_date" validate:"required"` // YYYY-MM-DD
	CategoryID         string `json:"category_id" validate:"required"`
	FinancialAccountID string `json:"financial_account_id" validate:"required"`
	Amount             int64  `json:"amount" validate:"required,gt=0"`
	Description        string `json:"description" validate:"required,min=3"`
	Vendor             string `json:"vendor,omitempty"`
	ReferenceNumber    string `json:"reference_number,omitempty"`
	Notes              string `json:"notes,omitempty"`
}

type ApproveExpenseRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected voided"`
	Notes  string `json:"notes,omitempty"`
}

// ==========================================
// TAHAP 5: BAHAN BAKU, SUPPLIER, & RESEP (HPP)
// ==========================================

// Ingredient bahan baku persediaan (kopi beans, susu, sirup, gula aren, cup).
type Ingredient struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Unit         string    `json:"unit"` // 'kg', 'g', 'l', 'ml', 'pcs', 'pack'
	CurrentStock float64   `json:"current_stock"`
	MinimumStock float64   `json:"minimum_stock"`
	CostPerUnit  int64     `json:"cost_per_unit"` // Biaya rata-rata tertimbang per satuan (INTEGER Rupiah)
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateIngredientRequest struct {
	Name         string  `json:"name" validate:"required,min=2"`
	Unit         string  `json:"unit" validate:"required,oneof=kg g l ml pcs pack"`
	MinimumStock float64 `json:"minimum_stock" validate:"gte=0"`
	CostPerUnit  int64   `json:"cost_per_unit" validate:"gte=0"`
}

// Supplier pemasok bahan baku.
type Supplier struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone,omitempty"`
	Email         string    `json:"email,omitempty"`
	Address       string    `json:"address,omitempty"`
	ContactPerson string    `json:"contact_person,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateSupplierRequest struct {
	Name          string `json:"name" validate:"required,min=2"`
	Phone         string `json:"phone,omitempty"`
	Email         string `json:"email,omitempty,omitempty"`
	Address       string `json:"address,omitempty"`
	ContactPerson string `json:"contact_person,omitempty"`
}

// IngredientPurchase transaksi pengadaan/pembelian stok bahan baku.
type IngredientPurchase struct {
	ID                 string                    `json:"id"`
	PurchaseNumber     string                    `json:"purchase_number"`
	SupplierID         *string                   `json:"supplier_id,omitempty"`
	SupplierName       *string                   `json:"supplier_name,omitempty"`
	FinancialAccountID string                    `json:"financial_account_id"`
	AccountName        string                    `json:"account_name,omitempty"`
	PurchaseDate       string                    `json:"purchase_date"`
	TotalAmount        int64                     `json:"total_amount"` // INTEGER Rupiah
	ExpenseID          *string                   `json:"expense_id,omitempty"`
	CreatedByUserID    string                    `json:"created_by_user_id"`
	CreatedByUserName  string                    `json:"created_by_user_name,omitempty"`
	Notes              string                    `json:"notes,omitempty"`
	Items              []*IngredientPurchaseItem `json:"items,omitempty"`
	CreatedAt          time.Time                 `json:"created_at"`
}

type IngredientPurchaseItem struct {
	ID             string  `json:"id"`
	PurchaseID     string  `json:"purchase_id"`
	IngredientID   string  `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name,omitempty"`
	Unit           string  `json:"unit,omitempty"`
	Quantity       float64 `json:"quantity"`
	UnitCost       int64   `json:"unit_cost"` // INTEGER Rupiah
	Subtotal       int64   `json:"subtotal"`
}

type CreatePurchaseItemRequest struct {
	IngredientID string  `json:"ingredient_id" validate:"required"`
	Quantity     float64 `json:"quantity" validate:"required,gt=0"`
	UnitCost     int64   `json:"unit_cost" validate:"required,gte=0"`
}

type CreatePurchaseRequest struct {
	SupplierID         *string                     `json:"supplier_id,omitempty"`
	FinancialAccountID string                      `json:"financial_account_id" validate:"required"`
	PurchaseDate       string                      `json:"purchase_date" validate:"required"`
	Notes              string                      `json:"notes,omitempty"`
	Items              []CreatePurchaseItemRequest `json:"items" validate:"required,min=1,dive"`
}

// Recipe Bill of Materials (BOM) per menu item untuk kalkulasi HPP otomatis.
type Recipe struct {
	ID           string        `json:"id"`
	MenuItemID   string        `json:"menu_item_id"`
	MenuItemName string        `json:"menu_item_name,omitempty"`
	TotalHPP     int64         `json:"total_hpp"` // Total HPP dihitung dari item
	SellingPrice int64         `json:"selling_price,omitempty"`
	MarginRp     int64         `json:"margin_rp,omitempty"`
	MarginPct    float64       `json:"margin_pct,omitempty"`
	Notes        string        `json:"notes,omitempty"`
	Items        []*RecipeItem `json:"items,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type RecipeItem struct {
	ID             string  `json:"id"`
	RecipeID       string  `json:"recipe_id"`
	IngredientID   string  `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name,omitempty"`
	Unit           string  `json:"unit,omitempty"`
	Quantity       float64 `json:"quantity"` // Kebutuhan per porsi
	CostPerUnit    int64   `json:"cost_per_unit,omitempty"`
	SubtotalCost   int64   `json:"subtotal_cost,omitempty"`
}

type CreateRecipeItemRequest struct {
	IngredientID string  `json:"ingredient_id" validate:"required"`
	Quantity     float64 `json:"quantity" validate:"required,gt=0"`
}

type CreateRecipeRequest struct {
	MenuItemID string                    `json:"menu_item_id" validate:"required"`
	Notes      string                    `json:"notes,omitempty"`
	Items      []CreateRecipeItemRequest `json:"items" validate:"required,min=1,dive"`
}

// InventoryMovement kartu stok bahan baku.
type InventoryMovement struct {
	ID                 string    `json:"id"`
	IngredientID       string    `json:"ingredient_id"`
	IngredientName     string    `json:"ingredient_name,omitempty"`
	MovementType       string    `json:"movement_type"` // 'purchase', 'sale_consumption', 'waste_loss', 'adjustment'
	Quantity           float64   `json:"quantity"`
	PreviousStock      float64   `json:"previous_stock"`
	NewStock           float64   `json:"new_stock"`
	ReferenceType      string    `json:"reference_type,omitempty"`
	ReferenceID        string    `json:"reference_id,omitempty"`
	Reason             string    `json:"reason,omitempty"`
	PerformedByUserID  *string   `json:"performed_by_user_id,omitempty"`
	PerformedByUserName *string  `json:"performed_by_user_name,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// ==========================================
// LAPORAN KEUANGAN & LABA RUGI (PROFIT & LOSS)
// ==========================================

type ProfitLossReport struct {
	PeriodStart      string                   `json:"period_start"`
	PeriodEnd        string                   `json:"period_end"`
	GrossSales       int64                    `json:"gross_sales"`       // Penjualan kotor
	Discounts        int64                    `json:"discounts"`         // Total potongan
	NetSales         int64                    `json:"net_sales"`         // Penjualan bersih
	TaxCollected     int64                    `json:"tax_collected"`     // Pajak PB1 10%
	COGS             int64                    `json:"cogs"`              // HPP Bahan Baku
	GrossProfit      int64                    `json:"gross_profit"`      // Net Sales - COGS
	GrossMarginPct   float64                  `json:"gross_margin_pct"`  // (Gross Profit / Net Sales) * 100
	OperatingExpenses int64                   `json:"operating_expenses"` // Total Biaya Operasional
	NetProfit        int64                    `json:"net_profit"`        // Gross Profit - Operating Expenses
	NetMarginPct     float64                  `json:"net_margin_pct"`    // (Net Profit / Net Sales) * 100
	ExpenseBreakdown []*ExpenseCategorySummary `json:"expense_breakdown"`
}

type ExpenseCategorySummary struct {
	CategoryName string `json:"category_name"`
	TotalAmount  int64  `json:"total_amount"`
	Percentage   float64 `json:"percentage"`
}



