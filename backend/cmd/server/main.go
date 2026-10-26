package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"kopirise/backend/internal/config"
	"kopirise/backend/internal/database"
	"kopirise/backend/internal/handler"
	"kopirise/backend/internal/middleware"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
	"kopirise/backend/internal/service"
)

func main() {
	// Inisialisasi Structured Logger (JSON di prod, Text di dev)
	cfg := config.Load()

	var logHandler slog.Handler
	if cfg.AppEnv == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	slog.Info("Memulai KopiRise Backend Service...",
		"env", cfg.AppEnv,
		"port", cfg.Port,
		"db", cfg.DBPath,
	)

	// Inisialisasi Database SQLite
	db, err := database.Connect(cfg.DBPath)
	if err != nil {
		slog.Error("Gagal menghubungkan ke database SQLite", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Inisialisasi Layer Repositori
	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	tableRepo := repository.NewTableRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	financeRepo := repository.NewFinanceRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)

	// Inisialisasi Layer Layanan (Service)
	authService := service.NewAuthService(cfg, userRepo, auditRepo)
	userService := service.NewUserService(cfg, userRepo, auditRepo)
	menuService := service.NewMenuService(menuRepo, auditRepo)
	tableService := service.NewTableService(tableRepo, auditRepo)
	reservationService := service.NewReservationService(reservationRepo, tableRepo, auditRepo)
	orderService := service.NewOrderService(orderRepo, menuRepo, tableRepo, auditRepo)
	paymentService := service.NewPaymentService(paymentRepo, orderRepo, auditRepo)
	financeService := service.NewFinanceService(financeRepo, auditRepo)
	inventoryService := service.NewInventoryService(inventoryRepo, auditRepo)
	exportService := service.NewExportService()

	// Inisialisasi Layer Handler HTTP
	authHandler := handler.NewAuthHandler(cfg, authService)
	userHandler := handler.NewUserHandler(userService)
	menuHandler := handler.NewMenuHandler(cfg, menuService)
	tableHandler := handler.NewTableHandler(tableService)
	reservationHandler := handler.NewReservationHandler(reservationService)
	orderHandler := handler.NewOrderHandler(orderService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	financeHandler := handler.NewFinanceHandler(financeService, inventoryService, exportService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	// Inisialisasi Router Utama Chi
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.StructuredLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.BodySizeLimiter(cfg.MaxRequestBodyBytes))
	r.Use(chimw.Compress(5))
	r.Use(middleware.SessionAuth(cfg, authService))

	// Static uploads file serving (Safe read-only)
	uploadsDir := "data/uploads"
	_ = os.MkdirAll(uploadsDir, 0755)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// Health Check & Readiness Endpoints
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		handler.RespondSuccess(w, map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "kopirise-backend",
		})
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			handler.RespondError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database belum siap menerima koneksi", nil)
			return
		}

		handler.RespondSuccess(w, map[string]interface{}{
			"status":    "ready",
			"database":  "connected",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Kelompok REST API v1
	r.Route("/api/v1", func(v1 chi.Router) {
		// --- 1. PUBLIK (Tanpa Login) ---
		v1.Group(func(pub chi.Router) {
			// Katalog Menu Publik
			pub.Get("/public/menu", menuHandler.ListMenuItems)
			pub.Get("/public/menu/categories", menuHandler.ListCategories)
			pub.Get("/public/menu/{id}", menuHandler.GetMenuItem)

			// Reservasi Meja Publik
			pub.Post("/public/reservations", reservationHandler.CreatePublicReservation)
			pub.Get("/public/reservations/{number}", reservationHandler.GetByNumber)

			// Scan QR Meja & Order Publik dari Meja
			pub.Get("/public/qr/{token}", tableHandler.ResolvePublicQR)
			pub.Post("/public/orders", orderHandler.CreateOrder)
			pub.Get("/public/orders/{id}", orderHandler.GetOrder)
		})

		// --- 2. AUTHENTICATION ---
		v1.Group(func(auth chi.Router) {
			auth.Use(httprate.LimitByIP(10, time.Minute)) // Maksimal 10 percobaan login per menit per IP
			auth.Post("/auth/login", authHandler.Login)
		})
		v1.Post("/auth/logout", authHandler.Logout)

		v1.Group(func(protected chi.Router) {
			protected.Use(middleware.RequireAuth)
			protected.Get("/auth/me", authHandler.Me)
		})

		// --- 3. MENU & KATEGORI (Admin & Manager) ---
		v1.Route("/menu", func(m chi.Router) {
			m.Use(middleware.RequireAuth)
			m.Use(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager))

			// Kategori
			m.Get("/categories", menuHandler.ListCategories)
			m.Post("/categories", menuHandler.CreateCategory)
			m.Put("/categories/{id}", menuHandler.UpdateCategory)
			m.Delete("/categories/{id}", menuHandler.DeleteCategory)

			// Item Menu
			m.Get("/items", menuHandler.ListMenuItems)
			m.Get("/items/{id}", menuHandler.GetMenuItem)
			m.Post("/items", menuHandler.CreateMenuItem)
			m.Put("/items/{id}", menuHandler.UpdateMenuItem)
			m.Delete("/items/{id}", menuHandler.DeleteMenuItem)
			m.Patch("/items/{id}/availability", menuHandler.ToggleAvailability)

			// Upload Foto Menu
			m.Post("/upload-photo", menuHandler.UploadPhoto)
		})

		// --- 4. MEJA & QR TOKEN (Admin, Manager, Kasir, Waiter) ---
		v1.Route("/tables", func(tbl chi.Router) {
			tbl.Use(middleware.RequireAuth)

			// Kasir, Barista, Pelayan, Manager, Admin dapat melihat daftar meja
			tbl.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier, models.RoleWaiter)).Get("/", tableHandler.ListTables)
			tbl.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier, models.RoleWaiter)).Get("/areas", tableHandler.ListAreas)
			tbl.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier, models.RoleWaiter)).Get("/{id}", tableHandler.GetTable)

			// Hanya Admin & Manager yang dapat membuat meja atau mengubah status & QR
			tbl.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager)).Post("/", tableHandler.CreateTable)
			tbl.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager)).Put("/{id}", tableHandler.UpdateTable)
			tbl.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager)).Post("/{id}/qr/generate", tableHandler.GenerateQR)
			tbl.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager)).Post("/{id}/qr/revoke", tableHandler.RevokeQR)
		})

		// --- 5. RESERVASI MEJA ADMIN (Manager & Waiter) ---
		v1.Route("/reservations", func(rsv chi.Router) {
			rsv.Use(middleware.RequireAuth)
			rsv.Use(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleWaiter))

			rsv.Get("/", reservationHandler.ListReservations)
			rsv.Patch("/{id}/status", reservationHandler.UpdateStatus)
		})

		// --- 6. USER & EMPLOYEE MANAGEMENT (Admin Only) ---
		v1.Route("/users", func(users chi.Router) {
			users.Use(middleware.RequireAuth)

			// Manager dan Admin dapat melihat daftar pengguna
			users.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager)).Get("/", userHandler.List)

			// Hanya Admin/Pemilik yang dapat membuat, mengubah status, atau reset password akun
			users.With(middleware.RequireRole(models.RoleAdmin)).Post("/", userHandler.Create)
			users.With(middleware.RequireRole(models.RoleAdmin)).Put("/{id}", userHandler.Update)
			users.With(middleware.RequireRole(models.RoleAdmin)).Post("/{id}/reset-password", userHandler.ResetPassword)
		})

		// --- 7. ORDER MANAGEMENT & POS ---
		v1.Route("/orders", func(o chi.Router) {
			o.Use(middleware.RequireAuth)

			// Melihat daftar order & detail order (Staff kafe)
			o.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier, models.RoleBarista, models.RoleWaiter)).Get("/", orderHandler.ListOrders)
			o.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier, models.RoleBarista, models.RoleWaiter)).Get("/{id}", orderHandler.GetOrder)

			// Antrean KDS Kitchen (Barista & Dapur)
			o.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleBarista, models.RoleCashier)).Get("/kitchen", orderHandler.ListKitchenOrders)

			// Struk Thermal
			o.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier)).Get("/{id}/receipt", orderHandler.GetReceipt)

			// Pembuatan order via POS
			o.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier, models.RoleWaiter)).Post("/", orderHandler.CreateOrder)

			// Perubahan status order keseluruhan & status per item
			o.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier, models.RoleBarista, models.RoleWaiter)).Patch("/{id}/status", orderHandler.UpdateOrderStatus)
			o.With(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleBarista)).Patch("/items/{itemId}/status", orderHandler.UpdateOrderItemStatus)
		})

		// --- 8. PEMBAYARAN & KASIR ---
		v1.Route("/payments", func(p chi.Router) {
			p.Use(middleware.RequireAuth)
			p.Use(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier))

			p.Post("/", paymentHandler.ProcessPayment)
		})

		// --- 9. REFUND TER-AUDIT (Admin & Manager) ---
		v1.Route("/refunds", func(rf chi.Router) {
			rf.Use(middleware.RequireAuth)
			rf.Use(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager))

			rf.Post("/", paymentHandler.ProcessRefund)
		})

		// --- 10. SHIFT KASIR & KAS OPERASIONAL ---
		v1.Route("/cash-shifts", func(cs chi.Router) {
			cs.Use(middleware.RequireAuth)
			cs.Use(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager, models.RoleCashier))

			cs.Get("/current", paymentHandler.GetCurrentShift)
			cs.Post("/open", paymentHandler.OpenShift)
			cs.Post("/close", paymentHandler.CloseShift)
			cs.Post("/movements", paymentHandler.CreateMovement)
		})

		// --- 11. KEUANGAN, BEBAN OPERASIONAL, & LAPORAN LABA RUGI ---
		v1.Route("/finance", func(fin chi.Router) {
			fin.Use(middleware.RequireAuth)
			fin.Use(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager))

			// Bagan Akun Keuangan & Transfer Dana
			fin.Get("/accounts", financeHandler.ListAccounts)
			fin.Post("/transfers", financeHandler.CreateTransfer)
			fin.Get("/transfers", financeHandler.ListTransfers)

			// Kategori Beban & Pengeluaran Operasional
			fin.Get("/categories", financeHandler.ListExpenseCategories)
			fin.Get("/expenses", financeHandler.ListExpenses)
			fin.Post("/expenses", financeHandler.CreateExpense)
			fin.Get("/expenses/{id}", financeHandler.GetExpense)
			fin.Patch("/expenses/{id}/status", financeHandler.UpdateExpenseStatus)

			// Laporan Laba Rugi Komprehensif & Ekspor
			fin.Get("/report/profit-loss", financeHandler.GetProfitLossReport)
			fin.Get("/export/excel", financeHandler.ExportExcel)
			fin.Get("/export/csv", financeHandler.ExportCSV)
		})

		// --- 12. PERSEDIAAN, BAHAN BAKU, SUPPLIER, & RESEP HPP ---
		v1.Route("/inventory", func(inv chi.Router) {
			inv.Use(middleware.RequireAuth)
			inv.Use(middleware.RequireAnyRole(models.RoleAdmin, models.RoleManager))

			// Bahan Baku
			inv.Get("/ingredients", inventoryHandler.ListIngredients)
			inv.Post("/ingredients", inventoryHandler.CreateIngredient)
			inv.Put("/ingredients/{id}", inventoryHandler.UpdateIngredient)
			inv.Get("/ingredients/low-stock", inventoryHandler.ListLowStock)

			// Supplier
			inv.Get("/suppliers", inventoryHandler.ListSuppliers)
			inv.Post("/suppliers", inventoryHandler.CreateSupplier)

			// Pembelian / Pengadaan Bahan
			inv.Get("/purchases", inventoryHandler.ListPurchases)
			inv.Post("/purchases", inventoryHandler.CreatePurchase)

			// Resep (Bill of Materials) & Kalkulasi HPP
			inv.Get("/recipes/{menuItemId}", inventoryHandler.GetRecipe)
			inv.Post("/recipes", inventoryHandler.SaveRecipe)
		})

		// Roles lookup
		v1.Get("/roles", func(w http.ResponseWriter, r *http.Request) {
			roles := []map[string]string{
				{"id": models.RoleAdmin, "name": "Admin / Pemilik"},
				{"id": models.RoleManager, "name": "Store Manager"},
				{"id": models.RoleCashier, "name": "Kasir"},
				{"id": models.RoleBarista, "name": "Barista / Dapur"},
				{"id": models.RoleWaiter, "name": "Pelayan"},
				{"id": models.RoleCustomer, "name": "Pelanggan"},
			}
			handler.RespondSuccess(w, roles)
		})
	})

	// Konfigurasi HTTP Server dengan Timeout
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Menjalankan server dalam goroutine terpisah
	go func() {
		slog.Info("Server HTTP KopiRise aktif dan mendengarkan", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Gagal menjalankan HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// Menunggu sinyal OS untuk Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Menerima sinyal shutdown, mematikan server secara anggun...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		slog.Error("Shutdown server terpaksa dilakukan", "error", err)
	}

	slog.Info("Server KopiRise berhasil dimatikan dengan aman.")
}
