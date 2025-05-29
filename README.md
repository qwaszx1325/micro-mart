# Micro-Mart 開發指南

## 專案概述
Micro-Mart 是一個基於微服務架構的電子商務平台，使用 Go 語言開發。它遵循乾淨架構原則，具有清晰的關注點分離。

## 架構與結構

### 服務組織
- **服務**: 位於 `/services/` 目錄下，每個服務都是獨立的微服務
  - `user`: 用戶管理和身份驗證
  - `transaction-orchestrator`: 處理交易流程
  - `frontend_api`: 前端客戶端的 API 網關
  - `auth`: 身份驗證服務
  - `order`: 訂單管理
  - `product`: 產品目錄管理

### 套件結構
每個服務都遵循乾淨架構模式：
- `application/`: 應用服務，協調領域邏輯
- `domain/`: 核心業務邏輯、實體和介面
  - `model/` 或 `aggregate/`: 領域模型
  - `repository/`: 資料存取介面
  - `service/`: 領域服務
- `infrastructure/`: 實作細節
  - `db_impl/`: 資料庫連線設定
  - `ent_impl/`: Ent ORM 實作
  - `grpc_impl/`: gRPC 伺服器實作
  - `redis_impl/`: Redis 實作
- `config/`: 配置結構

### 共享套件
- `pkg/`: 共享工具和函式庫
  - `cfgloader/`: 配置載入工具
  - `db/`: 資料庫工具
  - `mmotel/`: OpenTelemetry 儀表化
  - `pb/`: Protocol buffer 定義
  - `responder/`: HTTP 回應工具
  - `utils/`: 一般工具

## 技術堆疊
- **程式語言**: Go 1.24
- **依賴注入**: Uber FX
- **API**: gRPC 與 Protocol Buffers
- **網頁框架**: Gin (用於 HTTP APIs)
- **資料庫**: PostgreSQL 搭配 Ent ORM
- **快取**: Redis
- **可觀測性**: OpenTelemetry 搭配 Jaeger
- **身份驗證**: JWT

## 用戶註冊流程

### 註冊流程概述
用戶註冊是透過多個微服務協作完成的，遵循交易協調模式確保資料一致性：

1. **前端 API 接收請求**：
   - 前端 API 服務接收用戶的註冊請求
   - 驗證請求資料的格式和完整性
   - 將請求轉發給交易協調器服務

2. **交易協調器處理註冊流程**：
   - 創建一個唯一的交易記錄，包含兩個步驟：用戶註冊和生成令牌
   - 調用用戶服務進行用戶註冊
   - 如果用戶註冊成功，調用身份驗證服務生成訪問令牌和刷新令牌
   - 如果任何步驟失敗，執行補償操作（例如，刪除已創建的用戶）
   - 更新交易狀態並返回結果

3. **用戶服務處理註冊**：
   - 驗證用戶資料（電子郵件、用戶名、密碼）
   - 對密碼進行雜湊處理
   - 在資料庫中創建用戶記錄
   - 返回用戶資訊

4. **身份驗證服務生成令牌**：
   - 根據用戶資訊生成 JWT 訪問令牌和刷新令牌
   - 將刷新令牌存儲在 Redis 中
   - 返回令牌資訊

5. **前端 API 處理響應**：
   - 設置刷新令牌作為 HttpOnly Cookie
   - 返回訪問令牌給客戶端

### 實作細節

#### 前端 API (frontend_api)
前端 API 服務使用 Gin 框架處理 HTTP 請求：
```go
// Register 處理用戶註冊請求
func (h *UserHandler) Register(c *gin.Context) {
    // 解析請求
    var req request.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 錯誤處理...
        return
    }

    // 使用交易協調器處理用戶註冊
    resp, err := h.transactionClient.RegisterUserTransaction(ctx, &req)
    if err != nil {
        // 錯誤處理...
        return
    }

    // 設定 HttpOnly Cookie 存儲刷新令牌
    c.SetCookie("refresh_token", resp.RefreshToken, 30*24*60*60, "/", "", false, true)

    // 回傳訪問令牌
    responder.Ok(response.GetAccessTokenResponse{
        Success:     true,
        AccessToken: resp.AccessToken,
    }).WithContext(c)
}
```

#### 交易協調器 (transaction-orchestrator)
交易協調器服務負責協調整個註冊流程：
```go
// RegisterUserTransaction 協調用戶註冊交易
func (s *TransactionService) RegisterUserTransaction(ctx context.Context, req *RegisterRequest) (*RegisterUserResponse, error) {
    // 生成唯一交易 ID
    txID := uuid.New().String()

    // 創建交易記錄
    tx := &model.Transaction{
        ID:     txID,
        Type:   "USER_REGISTRATION",
        Status: model.StatusPending,
        Steps: []model.TransactionStep{
            // 用戶註冊步驟
            {
                ServiceName: "user",
                Operation:   "Register",
                Status:      "PENDING",
                // ...
            },
            // 生成令牌步驟
            {
                ServiceName: "auth",
                Operation:   "GenerateTokens",
                Status:      "PENDING",
                // ...
            },
        },
        // ...
    }

    // 保存交易記錄
    err := s.transactionRepo.SaveTransaction(ctx, tx.ToAggregate())
    // 錯誤處理...

    // 執行第一步：用戶註冊
    userResp, err := s.userClient.Register(ctx, req)
    // 錯誤處理與狀態更新...

    // 執行第二步：生成令牌
    authResp, err := s.authClient.GenerateTokens(ctx, userResp.UserId, userResp.Username, userResp.Email, userResp.Role)
    // 錯誤處理與狀態更新...

    // 返回完整結果
    return &RegisterUserResponse{
        UserInfo: userResp,
        AuthInfo: authResp,
    }, nil
}
```

#### 用戶服務 (user)
用戶服務處理實際的用戶註冊邏輯：
```go
func (s *UserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.LoginResponse, error) {
    // 驗證請求
    if req.GetEmail() == "" || req.GetUsername() == "" || req.GetPassword() == "" {
        // 錯誤處理...
    }

    // 密碼雜湊
    hashPassword, _ := utils.HashPassword(req.GetPassword())

    // 創建用戶資料
    profile := entity.Profile{
        Email:    req.GetEmail(),
        Name:     req.GetUsername(),
        Password: hashPassword,
    }

    u := &aggregate.User{
        Profile: profile,
    }

    // 開始資料庫交易
    ctx, err := s.db.Begin(ctx)
    // 錯誤處理...

    // 註冊用戶
    userProfile, mmErr := s.userService.Register(ctx, u)
    // 錯誤處理與交易回滾...

    // 提交交易
    _, commitErr := s.db.Commit(ctx)
    // 錯誤處理...

    // 返回結果
    return &user.LoginResponse{
        Message:  "Registration successful",
        Success:  true,
        UserId:   userProfile.ID.String(),
        Username: userProfile.Profile.Name,
        Email:    userProfile.Profile.Email,
        Role:     "user",
    }, nil
}
```

## 開發工作流程

### 設定開發環境
1. 複製儲存庫
2. 安裝依賴套件: `go mod download`
3. 啟動基礎設施服務: `cd docker && docker-compose up -d`
4. 配置環境變數（請參考配置檔案了解必要變數）
5. 執行服務: `go run services/<service-name>/main.go`

### 測試註冊功能
1. 啟動所有必要的服務：
   ```bash
   go run services/user/main.go
   go run services/auth/main.go
   go run services/transaction-orchestrator/main.go
   go run services/frontend_api/main.go
   ```

2. 發送註冊請求：
   ```bash
   curl -X POST http://localhost:8080/api/v1/register \
     -H "Content-Type: application/json" \
     -d '{"username":"test_user","email":"test@example.com","password":"password123"}'
   ```

3. 檢查響應中的訪問令牌和 Cookie 中的刷新令牌

## 最佳實務
1. **乾淨架構**: 維持領域、應用和基礎設施之間的分離
2. **依賴注入**: 使用 Uber FX 進行依賴連接
3. **配置**: 遵循十二因子應用方法論，使用環境變數
4. **錯誤處理**: 使用 mm_error 套件進行一致的錯誤處理
5. **日誌記錄**: 使用 zap 進行結構化日誌記錄
6. **追蹤**: 使用 OpenTelemetry 對程式碼進行儀表化以提升可觀測性
7. **測試**: 為所有新功能撰寫測試
8. **文檔**: 為公開 API 和複雜邏輯撰寫文檔
9. **程式碼風格**: 遵循標準 Go 慣例並使用 gofmt
10. **版本控制**: 對 API 使用語義化版本控制
11. **語言規範**: 遵循專案語言使用標準

## 語言規範 (Language Standards)
- **程式碼註解**: 使用繁體中文撰寫所有程式碼註解
- **文檔說明**: API 文檔、README、技術文件均使用繁體中文
- **變數與函數命名**: 使用英文命名，但要有清楚的意義
- **錯誤訊息**: 所有用戶面向的錯誤訊息使用繁體中文
- **日誌記錄**: 結構化日誌中的描述性訊息使用繁體中文
- **測試案例**: 測試函數名稱使用英文，但測試描述和註解使用繁體中文