package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakehouse/restaurant-finance/internal/auth"
	"github.com/lakehouse/restaurant-finance/internal/config"
	"github.com/lakehouse/restaurant-finance/internal/exports"
	"github.com/lakehouse/restaurant-finance/internal/imports"
	"github.com/lakehouse/restaurant-finance/internal/kpi"
)

// Server holds all dependencies for the HTTP server
type Server struct {
	router           *chi.Mux
	config           *config.Config
	db               *pgxpool.Pool
	jwtService       *auth.JWTService
	kpiHandler       *KPIHandler
	importHandler    *ImportHandler
	drilldownHandler *DrilldownHandler
	exportHandler    *ExportHandler
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, db *pgxpool.Pool) *Server {
	// Initialize KPI services
	kpiStore := kpi.NewStore(db)
	kpiService := kpi.NewService(kpiStore)

	// Initialize import services
	importPipeline := imports.NewPipeline(db)
	importStore := imports.NewImportStore(db)
	mappingStore := imports.NewMappingStore(db)

	// Initialize export services
	exportService := exports.NewExportService(db)
	exportStore := exports.NewExportStore(db)

	s := &Server{
		router:           chi.NewRouter(),
		config:           cfg,
		db:               db,
		jwtService:       auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpireHours),
		kpiHandler:       NewKPIHandler(kpiService),
		importHandler:    NewImportHandler(importPipeline, importStore, mappingStore),
		drilldownHandler: NewDrilldownHandler(db),
		exportHandler:    NewExportHandler(exportService, exportStore),
	}
	s.setupMiddleware()
	s.setupRoutes()
	return s
}

func (s *Server) setupMiddleware() {
	// Request ID
	s.router.Use(middleware.RequestID)

	// Real IP
	s.router.Use(middleware.RealIP)

	// Logging
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				log.Printf(
					"[%s] %s %s %d %s",
					middleware.GetReqID(r.Context()),
					r.Method,
					r.URL.Path,
					ww.Status(),
					time.Since(start),
				)
			}()
			next.ServeHTTP(ww, r)
		})
	})

	// Recovery
	s.router.Use(middleware.Recoverer)

	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Timeout
	s.router.Use(middleware.Timeout(60 * time.Second))
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.Get("/health", s.handleHealth)

	// API v1
	s.router.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/login", s.handleLogin)

		// Public KPI routes (read-only, for dashboard)
		r.Get("/kpi/daily", s.kpiHandler.HandleDaily)
		r.Get("/kpi/drilldown/sales", s.drilldownHandler.HandleSales)

		// Public export routes (handler checks auth internally)
		r.Route("/exports", func(r chi.Router) {
			r.Get("/", s.exportHandler.HandleList)
			r.Post("/pnl", s.exportHandler.HandlePnL)
			r.Get("/{id}", s.exportHandler.HandleGet)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(s.jwtService))

			// Import routes (accountant or admin only)
			r.Route("/imports", func(r chi.Router) {
				r.Use(auth.RequireRole(auth.RoleOwnerAdmin, auth.RoleAccountant))
				r.Get("/", s.importHandler.HandleList)
				r.Post("/", s.importHandler.HandleCreate)
				r.Get("/{id}", s.importHandler.HandleGet)
			})

			// Mapping profiles
			r.Route("/mappings", func(r chi.Router) {
				r.Get("/", s.importHandler.HandleMappingsGet)
				r.Group(func(r chi.Router) {
					r.Use(auth.RequireRole(auth.RoleOwnerAdmin, auth.RoleAccountant))
					r.Post("/", s.importHandler.HandleMappingCreate)
				})
			})
		})
	})
}

// Router returns the chi router
func (s *Server) Router() *chi.Mux {
	return s.router
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Placeholder handlers - will be implemented in feature phases
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}

	// Query user from database
	var userID uuid.UUID
	var passwordHash string
	var role string
	var locationID uuid.UUID

	err := s.db.QueryRow(r.Context(), `
		SELECT u.id, u.password_hash, u.role, l.id as location_id
		FROM users u
		CROSS JOIN locations l
		WHERE u.email = $1
		LIMIT 1
	`, req.Email).Scan(&userID, &passwordHash, &role, &locationID)

	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	// Verify password (simple comparison for dev - use bcrypt in production)
	if passwordHash != req.Password && !checkPasswordHash(req.Password, passwordHash) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(userID, req.Email, auth.Role(role), locationID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	// Update last login
	s.db.Exec(r.Context(), `UPDATE users SET last_login = NOW() WHERE id = $1`, userID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    userID,
			"email": req.Email,
			"role":  role,
		},
	})
}

// Simple password hash check (for bcrypt hashed passwords)
func checkPasswordHash(password, hash string) bool {
	// For development, also allow plaintext comparison
	if password == hash {
		return true
	}
	// In production, use bcrypt.CompareHashAndPassword
	return false
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
