# go-ai-copilot

基于 Gin + Vue3 开发的 AI 代码助手，支持 RAG 知识库能力。

## 项目简介

go-ai-copilot 是一个可落地的轻量化 AI 开发提效工具，核心功能包括：

- 🤖 **AI 对话**: 支持流式输出、多轮上下文记忆
- 💻 **代码能力**: 代码生成、解释、优化、漏洞检测、单元测试生成
- 📚 **RAG 知识库**: 基于 PostgreSQL + pgvector 实现向量检索
- 🔐 **用户鉴权**: JWT 无状态认证、bcrypt 密码加密
- 💾 **会话管理**: 多会话支持、消息历史持久化 + Redis 缓存

## 技术栈

### 后端

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.22+ | 核心语言 |
| Gin | v1.9+ | Web 框架 |
| GORM | v2.0+ | ORM 框架 |
| go-openai | v1.17+ | OpenAI 兼容 SDK |
| pgvector | - | PostgreSQL 向量插件 |
| Redis | 7.0+ | 缓存数据库 |
| JWT | v5.0+ | 用户鉴权 |

### 前端

| 技术 | 说明 |
|------|------|
| Vue 3 | 核心框架 |
| Vite | 构建工具 |
| Element Plus | UI 组件库 |
| Pinia | 状态管理 |
| Axios | HTTP 请求 |

## 项目结构

```
go-ai-copilot/
├── cmd/server/                    # ========== 服务入口 ==========
│   └── main.go                    # 程序入口，加载配置、初始化各模块、启动服务
│
├── config/
│   └── config.yaml                # 主配置文件（数据库、Redis、AI模型等）
│
├── .env                           # 环境变量（API Keys）
│
├── internal/                      # ========== 核心业务模块 ==========
│   │
│   ├── config/                    # 配置加载模块
│   │   └── config.go             # 解析 config.yaml，加载 AI API Key
│   │
│   ├── database/                  # 数据库模块
│   │   └── database.go           # PostgreSQL 连接、AutoMigrate、pgvector 扩展
│   │
│   ├── cache/                     # Redis 缓存模块
│   │   └── cache.go              # 会话历史缓存、连接管理
│   │
│   ├── model/                     # 数据模型（ORM）
│   │   ├── user.go               # 用户模型（ID、用户名、密码、昵称）
│   │   ├── session.go            # 会话模型（用户ID、标题、创建时间）
│   │   └── rag.go                # RAG模型（文档、分块、向量化）
│   │
│   ├── handler/                   # 业务处理器（API 逻辑）
│   │   ├── user.go               # 用户注册/登录/信息更新
│   │   ├── chat.go               # AI 对话（普通/流式/多模式）
│   │   ├── session.go            # 会话 CRUD、历史管理
│   │   └── rag.go                # RAG 文档上传、向量化、检索
│   │
│   ├── middleware/                # 中间件
│   │   └── jwt.go                # JWT 认证中间件
│   │
│   ├── router/                    # 路由配置
│   │   └── router.go             # 所有 API 路由定义
│   │
│   └── rag/                       # RAG 核心逻辑
│       └── text_splitter.go      # 文本分块（1024字符/块，256重叠）
│
├── pkg/                           # ========== 公共工具包 ==========
│   ├── ai/                        # AI 客户端
│   │   ├── client.go             # OpenAI 兼容客户端（对话/流式）
│   │   └── embedding.go          # 向量化客户端
│   │
│   └── jwt/                       # JWT 工具
│       └── jwt.go                # Token 生成、解析、验证
│
├── web/                          # ========== 前端 Vue3 项目 ==========
│   ├── src/
│   │   ├── api/                  # API 请求封装
│   │   │   └── chat.ts          # 对话/用户/会话 API
│   │   ├── stores/               # Pinia 状态管理
│   │   │   ├── user.ts          # 用户状态
│   │   │   └── chat.ts          # 会话/消息状态
│   │   ├── views/                # 页面组件
│   │   │   ├── Login.vue        # 登录页面
│   │   │   └── Home.vue         # 主对话页面
│   │   ├── router/               # 路由配置
│   │   │   └── index.ts
│   │   ├── main.ts              # Vue 入口
│   │   └── App.vue              # 根组件
│   ├── index.html
│   ├── vite.config.ts
│   └── package.json
│
├── docker-compose.yml             # Docker 编排（PostgreSQL + Redis）
├── go.mod
├── go.sum
└── README.md
```

## 快速开始

### 1. 环境要求

- Go 1.22+
- PostgreSQL 15+ (带 pgvector 扩展)
- Redis 7.0+

### 2. 配置文件

创建 `.env` 文件（可选，也可以使用环境变量）：

```env
# AI API 配置
AI_API_KEY=your_ai_api_key_here
EMBEDDING_API_KEY=your_embedding_api_key_here
```

配置文件 `config.yaml` 已包含默认配置，按需修改：

```yaml
server:
  port: 8080
  mode: debug

ai:
  base_url: "https://api.deepseek.com"
  model: "deepseek-chat"
  temperature: 0.7
  max_tokens: 2000
  timeout: 120
  embedding_model: "text-embedding-3-small"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "go_ai_copilot"
  sslmode: "disable"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

jwt:
  secret: "go-ai-copilot-secret-key-change-in-production"
  expire_time: 24h
  issuer: "go-ai-copilot"
```

### 3. 启动数据库

```bash
# 使用 Docker 启动 PostgreSQL 和 Redis
docker-compose up -d
```

### 4. 启动服务

```bash
# 方式1：直接运行（自动加载 .env 文件）
go run ./cmd/server/

# 方式2：编译后运行
go build -o server ./cmd/server/
./server
```

### 5. 启动前端

```bash
cd web
npm install
npm run dev
```

### 6. 测试服务

```bash
# 健康检查
curl http://localhost:8080/health

# 用户注册
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "123456", "nickname": "测试用户"}'

# 用户登录
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "123456"}'

# 创建会话 (需要 Token)
curl -X POST http://localhost:8080/api/v1/session \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"title": "我的第一个会话"}'

# 发送消息 (需要 Token)
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"message": "你好，请介绍一下自己", "session_id": 1}'
```

## API 文档

### 用户认证

| 接口 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/api/v1/user/register` | POST | 用户注册 | 否 |
| `/api/v1/user/login` | POST | 用户登录 | 否 |
| `/api/v1/user/info` | GET | 获取用户信息 | 是 |
| `/api/v1/user/info` | PUT | 更新用户信息 | 是 |
| `/api/v1/user/password` | PUT | 修改密码 | 是 |

### 会话管理

| 接口 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/api/v1/session` | POST | 创建会话 | 是 |
| `/api/v1/session/list` | GET | 获取会话列表 | 是 |
| `/api/v1/session/:id` | GET | 获取会话 | 是 |
| `/api/v1/session/:id` | PUT | 更新会话 | 是 |
| `/api/v1/session/:id` | DELETE | 删除会话 | 是 |
| `/api/v1/session/:id/history` | GET | 获取历史消息 | 是 |

### AI 对话

| 接口 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/api/v1/chat` | POST | 普通对话 | 是 |
| `/api/v1/chat/stream` | POST | 流式对话 (SSE) | 是 |
| `/api/v1/chat/mode` | POST | 带模式对话 | 是 |

### RAG 知识库

| 接口 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/api/v1/rag/upload` | POST | 上传文档 | 是 |
| `/api/v1/rag/list` | GET | 文档列表 | 是 |
| `/api/v1/rag/:id` | GET | 文档详情 | 是 |
| `/api/v1/rag/:id` | DELETE | 删除文档 | 是 |
| `/api/v1/rag/search` | POST | 向量检索 | 是 |
| `/api/v1/rag/chat` | POST | RAG 对话 | 是 |

### 对话模式

通过 `/api/v1/chat/mode` 的 `mode` 参数选择：

- `chat` - 通用对话（默认）
- `code_generate` - 代码生成
- `code_explain` - 代码解释
- `code_optimize` - 代码优化
- `code_vuln` - 漏洞检测
- `code_test` - 单元测试生成

## 核心架构

### 1. 流式响应实现 (Goroutine + Channel + SSE)

```go
// 启动 Goroutine 调用 AI 流式接口
go func() {
    err := h.client.StreamChat(ctx, messages, func(chunk string) error {
        tokenChan <- chunk  // 通过 Channel 推送
        return nil
    })
    close(tokenChan)
}()

// 主 Goroutine 监听 Channel 和 Context
for {
    select {
    case <-ctx.Done():
        return // 用户断开，终止流
    case token := <-tokenChan:
        c.SSEvent("message", token)
    }
}
```

### 2. Context 精准控制

- 请求超时控制
- 用户断开连接时自动终止 AI 请求
- 资源清理

### 3. RAG 完整流程

```
文档上传 → 文本分块 → 向量化 → 存储向量 → 相似度检索 → Prompt 融合 → AI 回答
```

### 4. 分层架构

```
请求 → Router → Middleware → Handler → Model/Database
         ↓
      缓存层 (Redis)
```

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t go-ai-copilot .

# 运行
docker run -d -p 8080:8080 -e AI_API_KEY=your-key go-ai-copilot
```

### Docker Compose 部署

```bash
# 一键启动所有服务
docker-compose up -d
```

## 面试讲解要点

1. **流式响应实现**: 展示 Goroutine + Channel + SSE 的使用
2. **Context 最佳实践**: 展示如何优雅地处理用户断开
3. **工程化设计**: 分层架构、中间件、统一的错误处理
4. **AI 工程化**: RAG 完整流程、向量检索
5. **数据库设计**: 用户隔离、会话管理、消息历史

## 开发计划

- [x] MVP - 基础 HTTP 服务 + AI 对话
- [x] 用户鉴权模块
- [x] 会话与对话管理
- [x] RAG 知识库模块
- [x] 前端 Vue3 项目
- [x] Docker 部署

## 许可证

MIT License
