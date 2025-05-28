package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go" // JWT 库
	"github.com/gorilla/mux"      // 路由器
	"github.com/rs/cors"          // CORS 中间件
	"golang.org/x/crypto/bcrypt"  // 密码哈希
	"gorm.io/driver/sqlite"       // SQLite 驱动
	"gorm.io/gorm"                // Gorm ORM
)

// ============================================================================
// 全局变量和初始化
// ============================================================================

var (
	db          *gorm.DB      // Gorm 数据库连接实例
	jwtKey      = []byte("your_secret_key") // JWT 签名密钥，生产环境请使用更复杂的密钥并从环境变量获取
	router      *mux.Router   // Gorilla Mux 路由器实例
	corsHandler http.Handler  // 包含 CORS 逻辑的 HTTP 处理器
)

// init 函数在包被导入时执行一次。
// 在 Vercel Serverless Function 中，它会在函数冷启动时执行，用于初始化数据库连接、路由器和 CORS 配置。
func init() {
	// 初始化日志
	log.SetOutput(os.Stdout)
	log.Println("Initializing Go API service...")

	// 1. 初始化数据库 (SQLite)
	// 注意：在 Vercel Serverless Function 中使用 SQLite 意味着数据是非持久化的。
	// 每次函数冷启动时，数据库文件可能会被重置，导致数据丢失。
	// 生产环境请使用外部的持久化数据库服务，如 PostgreSQL, MySQL 等。
	var err error
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected successfully.")

	// 自动迁移数据库模型
	err = db.AutoMigrate(&User{}, &Product{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	log.Println("Database auto-migration completed.")

	// 2. 初始化 Gorilla Mux 路由器
	r := mux.NewRouter()

	// 3. 定义路由
	// 认证路由 (无需 Token)
	r.HandleFunc("/register", RegisterHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")

	// 产品 CRUD 路由 (需要 Token 认证)
	// 创建子路由器并应用认证中间件
	authRouter := r.PathPrefix("/api").Subrouter()
	authRouter.Use(AuthMiddleware) // 将认证中间件应用于所有 /api 路径下的路由

	authRouter.HandleFunc("/products", CreateProduct).Methods("POST")
	authRouter.HandleFunc("/products", GetProducts).Methods("GET")
	authRouter.HandleFunc("/products/{id}", GetProduct).Methods("GET")
	authRouter.HandleFunc("/products/{id}", UpdateProduct).Methods("PUT")
	authRouter.HandleFunc("/products/{id}", DeleteProduct).Methods("DELETE")

	// 将路由器赋值给全局变量
	router = r

	// 4. 配置 CORS
	// 生产环境请将 AllowedOrigins 替换为你的前端实际域名，例如 "https://your-frontend-domain.com"
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:5173", "https://*.vercel.app"}, // 允许的源，可以添加多个
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // 允许的 HTTP 方法
		AllowedHeaders:   []string{"Content-Type", "Authorization"},          // 允许的请求头部
		AllowCredentials: true,                                               // 允许携带凭据 (如 cookies, HTTP authentication)
		MaxAge:           86400,                                              // 预检请求的缓存时间 (秒，这里是 24 小时)
		Debug:            false,                                              // 生产环境通常设置为 false，开发时可设为 true 查看 CORS 日志
	})

	// 将 CORS 中间件包装到路由器上，赋值给全局变量
	corsHandler = c.Handler(router)
	log.Println("CORS configured.")
	log.Println("API service initialization complete.")
}

// ============================================================================
// 数据模型
// ============================================================================

// User 用户模型
type User struct {
	gorm.Model // Gorm 提供的 ID, CreatedAt, UpdatedAt, DeletedAt 字段
	Username   string `gorm:"unique;not null"` // 用户名，唯一且非空
	Password   string `gorm:"not null"`        // 密码 (存储哈希值)
}

// Product 产品模型
type Product struct {
	gorm.Model // Gorm 提供的 ID, CreatedAt, UpdatedAt, DeletedAt 字段
	Name        string  `json:"name" gorm:"not null"`        // 产品名称
	Description string  `json:"description"`                 // 产品描述
	Price       float64 `json:"price" gorm:"not null"`       // 产品价格
}

// Claims JWT 的自定义声明结构
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// ============================================================================
// JWT 辅助函数
// ============================================================================

// GenerateToken 生成 JWT token
func GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token 24 小时后过期
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return "", fmt.Errorf("could not sign token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken 验证 JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		log.Printf("Token validation error: %v", err)
		return nil, err
	}

	if !token.Valid {
		log.Println("Invalid token.")
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// ============================================================================
// 认证中间件
// ============================================================================

// AuthMiddleware 是一个用于验证 JWT token 的中间件
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("AuthMiddleware: Incoming request for %s %s", r.Method, r.URL.Path)

		// 从请求头部获取 Authorization 字段
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("AuthMiddleware: Missing Authorization header.")
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// 检查 Authorization 头部是否以 "Bearer " 开头
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:] // 提取 token 字符串
		} else {
			log.Println("AuthMiddleware: Invalid Authorization header format.")
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// 验证 token
		claims, err := ValidateToken(tokenString)
		if err != nil {
			log.Printf("AuthMiddleware: Token validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Printf("AuthMiddleware: Token validated for user: %s", claims.Username)

		// 将用户信息存储到请求上下文中，以便后续处理函数访问 (可选)
		ctx := context.WithValue(r.Context(), "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx)) // 将带有用户信息的上下文传递给下一个处理器
	})
}

// ============================================================================
// 认证处理器
// ============================================================================

// RegisterRequest 注册请求体结构
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterHandler 处理用户注册请求
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("RegisterHandler: Processing registration request.")
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("RegisterHandler: Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 检查用户名是否已存在
	var existingUser User
	if db.Where("username = ?", req.Username).First(&existingUser).Error == nil {
		log.Printf("RegisterHandler: Username '%s' already exists.", req.Username)
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("RegisterHandler: Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 创建新用户
	user := User{Username: req.Username, Password: string(hashedPassword)}
	result := db.Create(&user)
	if result.Error != nil {
		log.Printf("RegisterHandler: Error creating user in DB: %v", result.Error)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	log.Printf("RegisterHandler: User '%s' registered successfully.", req.Username)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// LoginRequest 登录请求体结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应体结构
type LoginResponse struct {
	Token string `json:"token"`
}

// LoginHandler 处理用户登录请求
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler: Processing login request.")
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("LoginHandler: Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 查找用户
	var user User
	result := db.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		log.Printf("LoginHandler: User '%s' not found or DB error: %v", req.Username, result.Error)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Printf("LoginHandler: Password mismatch for user '%s'.", req.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 签发 token
	tokenString, err := GenerateToken(user.Username)
	if err != nil {
		log.Printf("LoginHandler: Error generating token for user '%s': %v", req.Username, err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("LoginHandler: User '%s' logged in successfully, token issued.", req.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
}

// ============================================================================
// 产品 CRUD 处理器 (需要认证)
// ============================================================================

// CreateProduct 处理创建产品请求
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateProduct: Processing create product request.")
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		log.Printf("CreateProduct: Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := db.Create(&product)
	if result.Error != nil {
		log.Printf("CreateProduct: Error creating product in DB: %v", result.Error)
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	log.Printf("CreateProduct: Product '%s' created successfully (ID: %d).", product.Name, product.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// GetProducts 处理获取所有产品请求
func GetProducts(w http.ResponseWriter, r *http.Request) {
	log.Println("GetProducts: Processing get all products request.")
	var products []Product
	result := db.Find(&products)
	if result.Error != nil {
		log.Printf("GetProducts: Error fetching products from DB: %v", result.Error)
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	log.Printf("GetProducts: Retrieved %d products.", len(products))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetProduct 处理获取单个产品请求
func GetProduct(w http.ResponseWriter, r *http.Request) {
	log.Println("GetProduct: Processing get single product request.")
	vars := mux.Vars(r)
	id := vars["id"]

	var product Product
	result := db.First(&product, id)
	if result.Error != nil {
		log.Printf("GetProduct: Product with ID '%s' not found or DB error: %v", id, result.Error)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	log.Printf("GetProduct: Retrieved product (ID: %d, Name: %s).", product.ID, product.Name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// UpdateProduct 处理更新产品请求
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	log.Println("UpdateProduct: Processing update product request.")
	vars := mux.Vars(r)
	id := vars["id"]

	var product Product
	result := db.First(&product, id)
	if result.Error != nil {
		log.Printf("UpdateProduct: Product with ID '%s' not found or DB error: %v", id, result.Error)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	var updatedProduct Product
	err := json.NewDecoder(r.Body).Decode(&updatedProduct)
	if err != nil {
		log.Printf("UpdateProduct: Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 更新产品字段
	product.Name = updatedProduct.Name
	product.Description = updatedProduct.Description
	product.Price = updatedProduct.Price

	result = db.Save(&product)
	if result.Error != nil {
		log.Printf("UpdateProduct: Error updating product in DB (ID: %d): %v", product.ID, result.Error)
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	log.Printf("UpdateProduct: Product (ID: %d) updated successfully.", product.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// DeleteProduct 处理删除产品请求
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteProduct: Processing delete product request.")
	vars := mux.Vars(r)
	id := vars["id"]

	var product Product
	result := db.First(&product, id)
	if result.Error != nil {
		log.Printf("DeleteProduct: Product with ID '%s' not found or DB error: %v", id, result.Error)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	result = db.Delete(&product)
	if result.Error != nil {
		log.Printf("DeleteProduct: Error deleting product from DB (ID: %d): %v", product.ID, result.Error)
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	log.Printf("DeleteProduct: Product (ID: %d) deleted successfully.", product.ID)
	w.WriteHeader(http.StatusNoContent) // 204 No Content 表示成功删除但无响应体
}

// ============================================================================
// Vercel Serverless Function 入口点
// ============================================================================

// Handler 是 Vercel Go Serverless Function 的主要入口点。
// Vercel 会将所有 HTTP 请求传递给这个函数。
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handler: Received request - Method: %s, Path: %s", r.Method, r.URL.Path)
	// 将请求传递给预先初始化好的 CORS 处理器，它会进一步调用我们的 Mux 路由器。
	corsHandler.ServeHTTP(w, r)
}

// 注意：在本地测试时，你可以在 main 函数中启动一个 HTTP 服务器来模拟 Vercel 环境。
// 但在 Vercel 部署时，main 函数不会被直接调用，只有 Handler 函数会被调用。
/*
func main() {
	log.Println("Starting local server for testing...")
	// init() 函数会自动执行，所以不需要在这里再次调用
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
*/
