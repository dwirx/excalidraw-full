# Excalidraw-Complete 架构文档

本文档旨在详细阐述 `excalidraw-complete` 项目的系统架构、技术栈、模块设计和数据流，以便于开发者理解、维护和进行二次开发。

## 1. 概述 (Overview)

`excalidraw-complete` 是一个将优秀的开源白板工具 [Excalidraw](https://github.com/excalidraw/excalidraw) 进行整合与封装的自托管解决方案。其核心目标是简化 Excalidraw 的私有化部署流程，将前端UI、后端数据存储和实时协作服务打包成一个单一的、易于部署的Go二进制文件。

**核心特性:**

- **一体化部署**：将所有服务打包成单个可执行文件，无需复杂的依赖配置。
- **可插拔存储**：通过环境变量支持多种数据持久化方案，包括内存、本地文件系统、SQLite和AWS S3。
- **实时协作**：内置基于 Socket.IO 的实时协作服务器，允许多个用户同时在同一个画板上工作。
- **Firebase 兼容层**：提供一个内存实现的 Firebase API 兼容层，以满足 Excalidraw 前端对 Firebase 的部分依赖。

---

## 2. 技术栈 (Tech Stack)

项目采用了现代化的前后端技术栈。

### 后端 (Backend)

- **语言**: [Go](https://go.dev/) (v1.21+)
- **Web框架**: [Chi (v5)](https://github.com/go-chi/chi) - 一个轻量级、高性能的 Go HTTP 路由器。
- **实时通信**: [Socket.IO for Go](https://github.com/zishang520/socket.io/v2) - 实现了 Socket.IO 协议，用于实时协作。
- **数据库驱动**:
    - [go-sqlite3](https://github.com/mattn/go-sqlite3) - 用于 SQLite 存储。
    - [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) - 用于与 AWS S3 交互。
- **日志**: [Logrus](https://github.com/sirupsen/logrus) - 结构化的日志记录库。
- **ID生成**: [ULID](https://github.com/oklog/ulid) - 用于生成唯一、可排序的文档ID。

### 前端 (Frontend)

- **核心**: [Excalidraw](https://github.com/excalidraw/excalidraw) (作为 Git Submodule)
- **框架**: [React](https://reactjs.org/)
- **构建工具**: [Vite](https://vitejs.dev/)
- **语言**: [TypeScript](https://www.typescriptlang.org/)

### 构建与部署 (Build & Deployment)

- **容器化**: [Docker](https://www.docker.com/) & `Dockerfile`
- **构建自动化**: Go Build Tools, npm/yarn

---

## 3. 系统架构 (System Architecture)

`excalidraw-complete` 是一个典型的**单体架构 (Monolith)**，但内部逻辑分层清晰。

```
+-------------------------------------------------------------------------+
|                                  User                                   |
| (Browser with Excalidraw React App)                                     |
+-------------------------------------------------------------------------+
       |                                      ^
       | HTTP/S (API Calls)                   | HTTP/S (HTML/JS/CSS)
       | WebSocket (Collaboration)            |
       v                                      |
+-------------------------------------------------------------------------+
|                  excalidraw-complete Go Binary                          |
|                                                                         |
|  +-------------------------+      +-----------------------------------+ |
|  |     HTTP Server (Chi)   |      |   Socket.IO Server                | |
|  |-------------------------|      |-----------------------------------| |
|  | - API Routes (/api/v2)  | <--> | - Connection Handling             | |
|  | - Firebase Routes       |      | - Room Management (Join/Leave)    | |
|  | - Static File Serving   |      | - Message Broadcasting            | |
|  +-------------------------+      +-----------------------------------+ |
|               |                                  ^                      |
|               |                                  |                      |
|               v                                  |                      |
|  +-------------------------------------------------------------------+  |
|  |                       Core Logic & Modules                        |  |
|  |-------------------------------------------------------------------|  |
|  |                                |                                  |  |
|  |  +--------------------------+  |  +-----------------------------+  |  |
|  |  |   Handlers (API Logic)   |  |  | Embedded Frontend Assets  |  |  |
|  |  +--------------------------+  |  | (Patched Excalidraw UI)     |  |  |
|  |               |                |  +-----------------------------+  |  |
|  |               v                |                                  |  |
|  |  +--------------------------+  |                                  |  |
|  |  |   Storage Interface      |  |                                  |  |
|  |  |  (core.DocumentStore)    |  |                                  |  |
|  |  +--------------------------+  |                                  |  |
|  |    |      |        |       |   |                                  |  |
|  |----|------|--------|-------|--------------------------------------|  |
|  v    v      v        v       v                                         |
| [S3] [SQLite] [FS] [Memory] (Storage Implementations)                   |
|                                                                         |
+-------------------------------------------------------------------------+
```

**架构说明:**

1.  **Go主程序 (`main.go`)**: 作为应用的入口，它初始化并启动所有服务。
2.  **HTTP服务器**: 使用 `Chi` 路由器来处理所有HTTP请求。这包括：
    - **API服务**: 提供用于创建和获取文档的 RESTful API。
    - **Firebase兼容层**: 模拟 Excalidraw 前端所需的 Firebase API。
    - **静态文件服务**: 将嵌入的、经过修改的 Excalidraw 前端应用（HTML, JS, CSS等）提供给浏览器。
3.  **Socket.IO服务器**: 独立处理 WebSocket 连接，负责所有实时协作功能，如同步绘图数据、光标位置等。
4.  **存储层 (`stores`)**: 通过一个统一的 `core.DocumentStore` 接口，将数据存储逻辑抽象出来。可以根据环境变量在启动时选择不同的实现（S3、SQLite等）。
5.  **嵌入式前端**: 前端 `Excalidraw` UI 作为一个 Git 子模块被包含在内。在构建阶段，它会被编译，并通过 Go 的 `embed` 特性直接嵌入到最终的二进制文件中。

---

## 4. 模块与服务说明 (Modules & Services)

### 4.1. 后端 (Backend)

#### 主应用 (`main.go`)

- **职责**: 应用的启动器和协调器。
- **核心逻辑**:
    - 解析命令行参数 (`-listen`, `-loglevel`)。
    - 根据环境变量初始化存储层 (`stores.GetStore()`)。
    - 设置 `Chi` 路由器 (`setupRouter`)，定义所有API路由。
    - 设置 `Socket.IO` 服务器 (`setupSocketIO`)，定义所有协作事件。
    - 将 `/socket.io/` 路径的请求代理到 Socket.IO 服务器。
    - **动态前端服务 (`handleUI`)**:
        - 使用 Go 的 `embed` 包将编译后的前端文件打包进二进制文件。
        - 在提供前端文件时，动态替换文件内容中的URL（如将 `firestore.googleapis.com` 替换为 `localhost:3002`），以重定向API请求到自身。
    - 监听系统信号以实现优雅停机 (`waitForShutdown`)。

#### 核心模块 (`core/`)

- **`core/entity.go`**: 定义了项目中最核心的数据结构和接口。
    - `Document`: 代表一个画板文档。
    - `DocumentStore`: 一个接口，定义了所有存储后端必须实现的两个方法：`FindID` 和 `Create`。这是实现可插拔存储的关键。

#### 存储层 (`stores/`)

- **`stores/storage.go`**: 工厂模式的实现。`GetStore()` 函数根据环境变量 `STORAGE_TYPE` 的值，创建并返回一个具体的 `DocumentStore` 接口实例。
- **存储实现**:
    - `stores/memory/`: 将文档保存在内存中，服务重启后数据丢失。
    - `stores/filesystem/`: 将每个文档作为单独的文件保存在本地文件系统上。
    - `stores/sqlite/`: 使用 SQLite 数据库来存储文档数据。
    - `stores/aws/`: 使用 AWS S3 对象存储来保存文档。

#### HTTP处理器 (`handlers/`)

- **`handlers/api/documents/`**: 实现了自定义的文档API (`/api/v2`)。
    - `HandleCreate`: 处理文档的创建请求。
    - `HandleGet`: 处理文档的读取请求。
- **`handlers/api/firebase/`**: 一个内存实现的 Firebase API 模拟层。它拦截了原始 Excalidraw 前端对 Firebase 的 `batchGet` 和 `commit` 请求，并在内存中进行处理，以确保前端协作功能可以正常工作，而无需真实的 Firebase 后端。

### 4.2. 前端 (Frontend)

#### Excalidraw UI (`excalidraw/` submodule)

- 项目通过 Git Submodule 引入了官方的 `excalidraw` 仓库。这使得跟踪上游更新变得容易。

#### 前端补丁 (`frontend.patch`)

- 这是一个至关重要的文件。由于我们是自托管，需要修改 Excalidraw 前端的一些硬编码配置。该补丁文件在构建时应用，主要做了以下修改：
    - **重定向API端点**: 将所有对 `excalidraw.com` 官方后端的API请求（如 `VITE_APP_BACKEND_V2_GET_URL`, `VITE_APP_WS_SERVER_URL`）重定向到自托管服务的地址（如 `http://localhost:3002`）。
    - **修改Firebase配置**: 清空部分 Firebase 配置，因为后端已经提供了兼容层。
    - **禁用追踪**: 设置 `VITE_APP_DISABLE_TRACKING=yes` 以禁用官方的数据追踪。

### 4.3. 前端架构分析 (Frontend Architecture)

`excalidraw` 自身是一个复杂的 `monorepo` 项目，其核心是可独立使用的 `@excalidraw/excalidraw` 包和一个完整的Web应用 `excalidraw-app`。我们的项目构建并嵌入的是 `excalidraw-app`。

#### `excalidraw-app` 项目地图 (Project Map)

以下是 `excalidraw/excalidraw-app` 目录的关键结构：

```
excalidraw-app/
├── public/              # 静态资源，如 a-icons, fonts, manifest
├── components/          # 应用的主要React组件
│   ├── AppWelcomeScreen.tsx # 欢迎界面
│   ├── CollabButton.tsx   # 协作按钮
│   ├── Library.tsx        # 元素库UI
│   ├── Tooltip.tsx        # 工具提示组件
│   └── ...                # 其他UI组件
├── data/                # 数据处理与持久化相关的模块
│   ├── localForage.ts     # IndexedDB的封装
│   ├── excalidraw.ts      # Excalidraw核心库的导出与封装
│   └── ...
├── collab/              # 实时协作相关逻辑
│   ├── Collab.tsx         # 协作功能的封装组件
│   ├── Portal.ts          # 管理协作房间和用户
│   └── index.ts           # 协作功能的初始化与管理
├── tests/               # 测试文件
├── App.tsx              # 应用的根React组件，组织所有UI和逻辑
├── index.tsx            # 应用的入口文件，将App组件渲染到DOM中
└── vite.config.mts      # Vite构建配置文件
```

#### 核心组件与逻辑

- **`App.tsx`**: 这是前端的"心脏"。它是一个巨大的组件，负责：
    - 渲染主要的 Excalidraw 画布 (`<Excalidraw />` 组件)。
    - 管理整个应用的状态（如图形元素、应用状态如当前工具、缩放等）。
    - 处理用户输入事件。
    - 初始化并集成协作模块 (`collab`)。

- **`components/`**: 包含了构成 Excalidraw 界面的所有可复用React组件，例如工具栏、菜单、对话框等。这使得UI层具有良好的模块化。

- **`collab/`**: 封装了所有与实时协作相关的功能。
    - 它使用 `socket.io-client` 与后端的 Socket.IO 服务器建立连接。
    - 负责发送和接收绘图数据、光标位置、用户加入/离开等事件。
    - `Portal.ts` 是关键，它维护了当前协作会话的状态。

- **`data/`**: 负责数据的加载和保存。在自托管模式下，它通过 `fetch` API 与我们的Go后端进行通信，以保存和加载画板数据。原始的 Firebase 逻辑被我们后端的兼容层所替代。

**总结**: 前端是一个高度组件化的 React 应用。通过 `frontend.patch`，我们巧妙地将其数据和协作的"后端"从官方服务切换到了我们自己的一体化Go服务器上，实现了完全的自托管。

---

## 5. 构建与部署 (Build & Deployment)

`