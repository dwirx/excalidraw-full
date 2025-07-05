# Excalidraw-Complete 改造计划 (BYOC - Bring Your Own Cloud Edition)

本文档旨在规划和跟踪将 `excalidraw-complete` 升级为一个支持用户认证、多画布管理，并具备前端直连云存储能力的协作平台所需要的开发任务。

**核心思想**: 后端负责 **"身份认证"** 与 **"默认存储"**，前端负责 **"存储适配与执行"**。

---

## ✅ 第一阶段：后端认证与用户体系基础

**目标**：为应用引入用户身份。这是所有个性化功能（如按用户存储画布）的基石。

### 后端 (Go)
- [x] **1.1.1**: 在 `go.mod` 中添加 `golang.org/x/oauth2` 依赖。
- [x] **1.1.2**: 创建新的 HTTP 处理器用于处理 OAuth2 流程。
- [x] **1.1.3**: 在 `main.go` 中添加认证路由:
    - [x] `GET /auth/github/login`
    - [x] `GET /auth/github/callback`
- [x] **1.1.4**: 实现从 GitHub API 获取用户信息的逻辑。
- [x] **1.1.5**: 引入 JWT 库 (e.g., `github.com/golang-jwt/jwt/v5`)。
- [x] **1.1.6**: 实现用户登录成功后生成和颁发 JWT 的逻辑。
- [x] **1.1.7**: 创建 `core/user.go` 定义 `User` 实体。
- [x] **1.1.8**: 创建一个可重用的 JWT 中间件，用于解析 Token 并将用户信息注入请求上下文。

### 前端 (React)
- [x] **1.2.1**: 在 UI 中AppWelcomeScreen中添加"使用 GitHub 登录"按钮。在excalidraw\excalidraw-app\components\AppMainMenu.tsx中添加"登录"按钮。
- [x] **1.2.2**: 添加api层，实现点击按钮后跳转到后端 `/auth/github/login` 的逻辑。
- [x] **1.2.3**: 创建一个用于处理登录回调的组件/页面，能从 URL 中解析出 JWT。
- [x] **1.2.4**: 将获取到的 JWT 安全地存储在 `localStorage` 或 `sessionStorage` 中。
- [x] **1.2.5**: 创建一个全局 API 请求封装（如 Axios 拦截器），为所有请求自动附加 `Authorization` 头。
- [x] **1.2.6**: 建立全局用户状态管理 (e.g., Jotai/Zustand)，并在登录后更新 UI（如显示用户头像）。

---

## ✅ 第二阶段：前端存储抽象层与UI框架

**目标**：在前端建立一个灵活的存储适配器架构和相应的UI，为后续接入多种存储后端做好准备。

### 前端 (React)
- [x] **2.1.1**: 在 `src/data/` 目录下创建 `storage.ts` 文件。
- [x] **2.1.2**: 在 `storage.ts` 中定义 `IStorageAdapter` TypeScript 接口，包含 `listCanvases`, `loadCanvas`, `saveCanvas`, `createCanvas`, `deleteCanvas` 等方法。
- [x] **2.1.3**: 设计并实现一个新的"数据源配置"设置页面或模态框。
- [x] **2.1.4**: 在设置UI中，创建一个下拉菜单，包含未来的存储选项（"默认后端", "Cloudflare KV", "Amazon S3"，"IndexDB"）。
- [x] **2.1.5**: 根据下拉菜单的选择，动态渲染用于输入凭证的表单。
- [x] **2.1.6**: 在 UI 上添加明确的安全警告，告知用户密钥仅存储在浏览器会话中。
- [x] **2.1.7**: 创建全局状态来管理存储配置，将敏感凭证存储在 `sessionStorage`，非敏感配置存储在 `localStorage`。

---

## ✅ 第三阶段：实现后端作为第一个KV存储适配器

**目标**：将项目自身的 Go 后端实现为一个简单的、面向用户的KV存储，作为第一个可用的存储选项。

### 后端 (Go) - KV API 设计
- **API理念**: 放弃复杂的RESTful设计，提供纯粹的KV操作接口，所有权与当前JWT用户绑定。
- **路由规划**:
    - `GET /api/v2/kv`: 列出当前用户所有画布的元信息 (ID, Name, UpdatedAt)。
    - `GET /api/v2/kv/{key}`: 获取单个画布的完整内容。
    - `PUT /api/v2/kv/{key}`: 创建或更新一个画布。
    - `DELETE /api/v2/kv/{key}`: 删除一个画布。

### 后端 (Go) - 执行步骤
- [x] **3.1.1**: 创建新的 `core/canvas.go` 文件，定义 `Canvas` 实体和 `CanvasStore` 接口。此举可避免与用于实时协作的旧 `Document` 模型冲突。
- [x] **3.1.2**: `Canvas` 实体将包含 `ID`, `UserID`, `Name`, `Data`, `CreatedAt`, `UpdatedAt` 字段。
- [x] **3.1.3**: `CanvasStore` 接口将定义 `List`, `Get`, `Save`, `Delete` 方法，所有方法都基于 `UserID` 操作以保证数据隔离。
- [x] **3.1.4**: 更新现有存储实现 (`sqlite`, `filesystem` 等) 以实现新的 `CanvasStore` 接口。
- [x] **3.1.5**: 创建新的 `handlers/api/kv/` 目录和处理器，实现上述KV API路由，并使用JWT中间件进行保护。

### 前端 (React)
- [x] **3.2.1**: 创建 `src/data/BackendStorageAdapter.ts` 文件，并使其实现 `IStorageAdapter` 接口。
- [x] **3.2.2**: 在该适配器内部，实现所有接口方法，使其通过 `fetch` 调用后端的 `/api/v2/kv` 相关 API。
- [x] **3.2.3**: 实现多画布管理的侧边栏 UI。
- [x] **3.2.4**: 将侧边栏 UI 与 `BackendStorageAdapter` 连接，实现一个功能完整的、由后端驱动的多画布管理系统。

---

## ✅ 第四阶段：实现Cloudflare KV客户端适配器

**目标**：实现第一个纯前端的存储选项，数据直接从浏览器发送到用户的Cloudflare KV。

### 前端 (React)
- [ ] **4.1.1**: 创建 `src/data/CloudflareKVAdapter.ts` 文件，并使其实现 `IStorageAdapter` 接口。
- [ ] **4.1.2**: 实现其构造函数，用于接收用户输入的 `accountId`, `namespaceId`, 和 `apiToken`。
- [ ] **4.1.3**: 在适配器内部，使用 `fetch` 实现对 Cloudflare KV 官方 API 的直接调用。
- [ ] **4.1.4**: 设计并在适配器中实现 KV 的键名（Key）管理策略。
- [ ] **4.1.5**: 在主应用逻辑中，当用户在设置中选择并配置了 Cloudflare KV 后，实例化并切换到 `CloudflareKVAdapter`。
- [ ] **4.1.6**: 验证所有画布操作（增删改查）都能在用户的 CF KV 上正确执行。

---

## ✅ 第五阶段：实现Amazon S3客户端适配器与最终打磨

**目标**：添加对S3的支持，并完善整个用户体验。

### 前端 (React)
- [ ] **5.1.1**: 在前端项目中添加 AWS SDK 依赖: `npm install @aws-sdk/client-s3`。
- [ ] **5.1.2**: 创建 `src/data/S3StorageAdapter.ts` 文件，并使其实现 `IStorageAdapter` 接口。
- [ ] **5.1.3**: 实现其构造函数，用于接收用户输入的 `accessKeyId`, `secretAccessKey`, `region`, `bucketName`。
- [ ] **5.1.4**: 在适配器内部，使用 `@aws-sdk/client-s3` 实现对 S3 对象的 `List`, `Get`, `Put`, `Delete` 操作。
- [ ] **5.1.5**: 设计并在适配器中实现 S3 的对象键（Key）管理策略。
- [ ] **5.1.6**: 在主应用逻辑中，当用户在设置中选择并配置了 S3 后，实例化并切换到 `S3StorageAdapter`。

### UX/UI 打磨
- [ ] **5.2.1**: 在每个数据源配置界面添加"测试连接"按钮，提供即时反馈。
- [ ] **5.2.2**: 完善在不同数据源之间切换时的用户体验，如提示保存未保存的更改。
- [ ] **5.2.3**: 在文档和UI中提供详细的指南，说明如何获取各种云服务的API密钥。 