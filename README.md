# go-ai-copilot

基于 Gin + Vue3 开发的 AI 代码助手，支持 RAG 知识库能力。

## 项目简介

go-ai-copilot 是一个可落地的轻量化 AI 开发提效工具，核心功能包括：

- 🤖 **AI 对话**: 支持流式输出、多轮上下文记忆
- 💻 **代码能力**: 代码生成、解释、优化、漏洞检测、单元测试生成
- 📚 **RAG 知识库**: 基于 PostgreSQL + pgvector 实现向量检索
- 🔐 **用户鉴权**: JWT 无状态认证、bcrypt 密码加密
- 💾 **会话管理**: 多会话支持、消息历史持久化

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
├── cmd/
│   └── server/           # 服务入口
│       └── main.go
├── config/              # 配置文件
│   └── config.yaml
├── internal/
│   ├── cache/           # Redis 缓存
│   │   └── cache.go
│   ├── config/          # 配置加载
│   │   └── config.go
│   ├── database/        # 数据库连接
│   │   └── database.go
│   ├── handler/        # 业务处理器
│   │   ├── chat.go
│   │   ├── session.go
│   │   └── user.go
│   ├── middleware/     # 中间件
│   │   └── jwt.go
│   ├── model/          # 数据模型
│   │   ├── session.go
│   │   └── user.go
│   └── router/         # 路由配置
│       └── router.go
├── pkg/
│   ├── ai/             # AI 客户端
│   │   └── client.go
│   └── jwt/            # JWT 工具
│       └── jwt.go
├── docker-compose.yml   # Docker 编排
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

创建 `config.yaml`:

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
  secret: "your-secret-key-change-in-production"
  expire_time: 24h
  issuer: "go-ai-copilot"
```

### 3. 启动数据库

```bash
# 使用 Docker 启动 PostgreSQL 和 Redis
docker-compose up -d

# 或者分别启动
docker run -d --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 pgvector/pgvector:pg15
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

### 4. 启动服务

```bash
# 设置 API Key (DeepSeek 为例)
export AI_API_KEY="your-api-key"

# 运行服务
go run ./cmd/server/

# 或运行编译后的二进制
./server.exe
```

### 5. 测试服务

```bash
# 健康检查
curl http://localhost:8080/health

# 用户注册
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "123456"}'

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
  -d '{"message": "你好，请介绍一下自己"}'
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
| `/api/v1/chat/stream` | POST | 流式对话 | 是 |
| `/api/v1/chat/mode` | POST | 带模式对话 | 是 |

### 对话模式

通过 `mode` 参数选择不同的 AI 能力：

- `chat` - 通用对话（默认）
- `code_generate` - 代码生成
- `code_explain` - 代码解释
- `code_optimize` - 代码优化
- `code_vuln` - 漏洞检测
- `code_test` - 单元测试生成

## 核心亮点

### 1. Go 并发编程

使用 Goroutine + Channel 处理流式响应，异步读取 AI 返回的 Token，实时推送给前端：

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

使用 Context 实现请求超时控制、用户断开连接时自动终止 AI 请求：

```go
// 创建带超时的上下文
ctx, cancel := context.WithCancel(c.Request.Context())
defer cancel()

// AI 调用监听 Context
select {
case <-ctx.Done():
    return // 用户断开，自动终止
}
```

### 3. 完整工程化规范

- 分层架构：API → Handler → Service → Repository
- 统一错误处理
- 统一返回格式 `{code, message, data}`
- 中间件解耦

### 4. RAG 知识库 (开发中)

- 文档上传与解析
- 文本分块
- 向量化存储
- 向量检索
- Prompt 融合

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
- [ ] RAG 知识库模块
- [ ] 前端 Vue3 项目
- [ ] 接口限流与安全

## 许可证

MIT License
