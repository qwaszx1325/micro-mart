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

## 開發工作流程

### 設定開發環境
1. 複製儲存庫
2. 安裝依賴套件: `go mod download`
3. 啟動基礎設施服務: `cd docker && docker-compose up -d`
4. 配置環境變數（請參考配置檔案了解必要變數）
5. 執行服務: `go run services/<service-name>/main.go`

### 新增新服務
1. 在 `/services/` 中建立新目錄
2. 遵循現有的服務結構 (application, domain, infrastructure)
3. 使用 FX 依賴注入建立 main.go 檔案
4. 在 config/config.go 中定義配置

### 修改現有服務
1. 尊重乾淨架構的邊界
2. 在領域層更新領域模型
3. 在基礎設施層實作介面
4. 使用 FX 在 main.go 中連接依賴關係

## 測試
- 使用標準的 Go 測試套件
- 對於資料庫測試，使用 `enttest` 套件中的 Ent 測試工具
- 使用模擬介面進行單元測試
- 整合測試應使用 docker-compose 處理依賴關係

## 部署
- 服務透過環境變數進行配置
- 使用 Docker 容器進行部署
- 基礎設施服務（PostgreSQL、Redis、Jaeger）在 docker-compose.yml 中定義

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