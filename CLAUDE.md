## 项目目标
构建中医智能助手（TCM Agent），支持基于中医古籍和现代文献的 Agentic RAG 问答，重点满足中医知识检索、智能诊断、养生建议、药材查询、引用溯源和流式对话体验。传承中医智慧，赋能健康生活。

## 开发原则
- 优先实现 MVP 闭环
- 先做最小可用功能，再逐步增强
- 保持代码简洁，避免过度抽象
- 优先复用现有结构，减少无关改动
- 改动后先做最小验证
- 沟通尽量简洁，减少 token 消耗

## 当前实现状态（2026-05-26）

### 已完成
- Go + Fiber 后端框架，所有路由和中间件就绪
- DeepSeek API 封装（同步 + SSE 流式）
- RAG Pipeline：Embedding → Qdrant 检索(召回 topK×2) → Reranker 重排序 → 上下文拼接 → LLM 生成
- Chat SSE 流式输出（token / citations / done / error 事件）
- Agent 模式：QueryAnalyzer + TaskPlanner + ReactExecutor + 5 个 TCM 工具 + 最终 LLM 合成
- Agent SSE 流式输出（meta / thought / task_progress / tool_call / tool_result / final_answer 事件）
- Redis 会话存储（Lua 脚本原子追加 + 自动裁剪到 50 条 + TTL 24h）
- Qdrant 向量检索（gRPC 客户端）
- Embedding / Reranker HTTP 客户端
- JWT 中间件（MVP 阶段免登录放行）
- CORS / 日志 / 限流中间件
- React 前端：ChatContainer / MessageList / ChatInput / MessageBubble / MarkdownRenderer
- 引用卡片（CitationCard + CitationPanel，可展开原文预览）
- Agent 可视化组件（ReasoningSteps / TaskProgressBar / ToolInvocationPanel / ThoughtBubble）完整实现
- 会话侧边栏（SessionSidebar，新建/切换/删除）
- Zustand 状态管理（chat / session / agent 三个 store）
- SSE 流式消费（RAG 和 Agent 两种模式独立处理）
- Agent 模式手动切换（AgentModeToggle）
- docker-compose.yml（Qdrant + Redis）

### 与计划的差距
- **Agent 规划器是关键词匹配，非 LLM 驱动**：QueryAnalyzer 检查 23 个中文关键词（如"分析""辨证""调理"），>=2 个匹配才进入 Agent 模式。TaskPlanner 同样是关键词→工具映射，非 LLM 推理。
- **ReAct 思考步骤是硬编码模板**：execute 循环中的 thought 是固定字符串，不是 LLM 生成。LLM 仅在最终合成步骤调用。
- **Query 改写未实现**：RAG Pipeline 没有 Query Rewriting 步骤，用户原始 query 直接送入 Embedding。
- **会话历史加载未实现**：前端 `selectSession` 只设 activeId，不调用 `getSession(id)` 拉取消息。点击历史会话无效果。
- **自动 Agent 检测未集成**：`useAgent.checkQuery` 已实现但无组件调用，Agent 模式全靠手动切换。
- **引用未内联到 Markdown**：LLM 输出中的 `[参考N]` 标记和 CitationCard 之间没有可点击链接。
- **前端 `queryClient.invalidateQueries` 未调用**：SSE 流结束后未同步 TanStack Query 缓存。
- **无测试**：后端 0 个 `*_test.go` 文件，前端无测试框架配置。
- **无文档预处理管线**：PDF 解析、语义分块、古籍 OCR 均为规划阶段，无代码。
- **长期记忆未实现**：仅有 Redis 短期记忆（TTL 24h）。

### 已知问题
- **CRITICAL**: `.gitignore` 未包含 `.env`，`backend/.env` 含真实 API Key 和 JWT Secret，提交即泄露。
- `server.exe` (30MB) 和 `server.exe~` (30MB) 不应提交到仓库。
- 前端 `sessionStore.selectSession` 不加载历史消息（需调用 `getSession` API）。
- `rehype-raw` 依赖已安装但从未 import，属于死依赖。
- `MarkdownRenderer` 中 `border-l-3` 和 `border-primary-400` 不是合法 Tailwind class，blockquote 左边框无样式。
- Agent SSE 解析不检查 `done`/`error` 事件提前退出，会空转直到服务端关闭连接。
- Qdrant 连接使用已废弃的 `grpc.WithBlock()`。
- 无优雅关闭（SIGTERM 处理），SSE 连接会被强制断开。
- `/health` 端点不检查依赖健康状态，始终返回 ok。
- 无 ErrorBoundary，React 渲染错误会白屏。

## 功能需求
- 自然语言对话问答（多轮会话管理）
- 流式输出（SSE）
- 中医古籍/文献引用溯源（引用卡片 + 原文预览）
- 知识库检索与召回（中医术语、症状、药材、方剂、穴位等）
- Agent 多步推理（症状诊断 → 药材推荐 → 养生建议 → 整合方案）
- 混合部署，核心能力优先私有化
- MVP：问答可用、引用可追踪、响应稳定

## 后端技术栈

| 组件 | 用途 | 优先级 |
|------|------|--------|
| Go + Fiber v2.52.10 | HTTP 服务 + SSE 流式 | MVP |
| DeepSeek-V4-Pro | LLM 底座，中文古籍能力强 | MVP |
| Qdrant | 向量存储，支持标量过滤（按古籍来源/药材/症状筛选） | MVP |
| Redis | 会话缓存 + 短期上下文，设 TTL + 条数上限 | MVP |
| BGE-M3 / stella-m3 | Embedding 模型，独立部署（TEI/Infinity） | MVP |
| BGE-Reranker-v2 | 召回后重排序，提升 Top-K 精度 | MVP |
| PostgreSQL | 业务数据持久化（对话历史/用户） | 后续引入 |
| JWT + API Key | 用户鉴权 + 服务间调用 | 按需引入 |

### RAG 架构（当前实现）
```
用户Query → Embedding → Qdrant检索(topK×2) → Reranker重排序(topK) → 拼接上下文 → LLM生成 → SSE流式输出
```
注意：Query 改写步骤已规划但未实现。

### Agent 架构（当前实现）
```
用户Query → QueryAnalyzer(关键词匹配) → TaskPlanner(关键词→工具映射) → ReactExecutor(硬编码thought + 工具调用) → LLM最终合成 → SSE流式输出
```
注意：规划器和执行器均非 LLM 驱动，最终合成步骤才是 LLM 调用。这是 MVP 简化，后续应升级为 LLM 驱动的 ReAct。

### TCM 工具集
| 工具名 | 功能 |
|--------|------|
| diagnose_symptom | 症状辨证分析 |
| herb_query | 药材性味归经、功效查询 |
| prescription_advice | 经典方剂推荐 |
| wellness_suggestion | 饮食/起居/情志养生建议 |
| acupoint_info | 穴位定位与按摩指导 |

所有工具复用 RAG Pipeline 检索知识库。

### 文档预处理管线（规划中，未实现）
- PDF/扫描件解析：MinerU / Marker
- Chunking：语义分块（按章节/段落），不用固定长度
- 古籍 OCR：PaddleOCR / Tesseract（按需）
- 中医知识图谱结构化（症状-药材-方剂关联）

## 前端技术栈

| 组件 | 用途 |
|------|------|
| React 18 | UI 框架 |
| Tailwind CSS 3 | 样式 |
| Zustand 4 | 客户端状态（chat / session / agent） |
| TanStack Query 5 | 服务端状态（仅 session 列表） |
| react-markdown 9 + remark-gfm | LLM 输出 Markdown 渲染 |
| rehype-raw | 已安装但未使用（死依赖） |

## 项目结构（实际）
```
agri-qa-system/
├── backend/
│   ├── cmd/server/main.go          # 入口：初始化 stores/services，注册 Fiber 路由
│   ├── config/config.go            # 环境变量配置
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── chat.go             # /api/v1/chat, /chat/stream, /sessions
│   │   │   └── agent.go            # /api/v1/agent/analyze, /agent/chat/stream
│   │   ├── service/
│   │   │   ├── chat.go             # ChatService: RAG + LLM 编排
│   │   │   ├── session.go          # 会话 CRUD（Redis 代理）
│   │   │   └── agent/
│   │   │       └── coordinator.go  # Agent 协调器
│   │   ├── agent/
│   │   │   ├── planner/
│   │   │   │   ├── query_analyzer.go  # 关键词匹配（非 LLM）
│   │   │   │   └── task_planner.go    # 关键词→工具映射（非 LLM）
│   │   │   ├── executor/
│   │   │   │   └── react_executor.go  # ReAct 循环 + LLM 最终合成
│   │   │   ├── tools/
│   │   │   │   ├── interface.go    # Tool 接口
│   │   │   │   ├── registry.go     # 工具注册表
│   │   │   │   ├── retrieve.go     # 共享检索辅助
│   │   │   │   └── tcm/            # 5 个 TCM 工具实现
│   │   │   │       ├── diagnose.go
│   │   │   │       ├── herb.go
│   │   │   │       ├── prescription.go
│   │   │   │       ├── wellness.go
│   │   │   │       └── acupoint.go
│   │   │   └── memory/
│   │   │       └── short_term.go   # Redis 短期记忆
│   │   ├── rag/pipeline.go         # RAG 检索+重排+上下文构建
│   │   ├── llm/deepseek.go         # DeepSeek API（同步 + SSE）
│   │   ├── store/
│   │   │   ├── qdrant.go           # Qdrant gRPC 客户端
│   │   │   ├── redis.go            # Redis 会话存储
│   │   │   └── embedding.go        # Embedding + Reranker HTTP 客户端
│   │   ├── middleware/middleware.go # CORS / Logger / RateLimit / Auth
│   │   └── model/
│   │       ├── model.go            # 核心类型
│   │       └── agent.go            # Agent 类型
│   ├── scripts/                    # 部署脚本
│   └── .env                        # 环境变量（含密钥，勿提交！）
├── frontend/
│   └── src/
│       ├── main.tsx                # React 入口
│       ├── App.tsx                 # QueryClientProvider 包装
│       ├── index.css               # Tailwind + 自定义样式
│       ├── components/
│       │   ├── Chat/               # ChatContainer, MessageList, ChatInput, MessageBubble, AgentModeToggle
│       │   ├── Agent/              # ReasoningSteps, TaskProgressBar, ToolInvocationPanel, ThoughtBubble
│       │   ├── Citation/           # CitationCard, CitationPanel
│       │   ├── Markdown/           # MarkdownRenderer
│       │   ├── Session/            # SessionSidebar
│       │   └── Layout/             # AppLayout
│       ├── stores/                 # Zustand: chatStore, sessionStore, agentStore
│       ├── services/               # api.ts, sse.ts, agent.ts
│       ├── hooks/                  # useChat, useSessions, useAgent
│       └── types/                  # index.ts, agent.ts
└── docker-compose.yml              # Qdrant + Redis
```

## API 接口（当前实现）

### POST /api/v1/chat — 非流式对话
### POST /api/v1/chat/stream — SSE 流式对话（RAG 模式）
事件类型：`meta` | `token` | `citations` | `done` | `error`

### POST /api/v1/agent/analyze — 查询是否需要 Agent 模式
### POST /api/v1/agent/chat/stream — SSE 流式对话（Agent 模式）
事件类型：`meta` | `thought` | `task_progress` | `tool_call` | `tool_result` | `final_answer` | `error`

### GET /api/v1/sessions — 会话列表
### GET /api/v1/sessions/:id — 会话详情
### DELETE /api/v1/sessions/:id — 删除会话

## 快速开始

### 环境要求
- Go 1.22+
- Node.js 20+
- Qdrant（向量数据库）
- Redis（会话缓存）
- DeepSeek API Key
- Embedding 服务（BGE-M3，独立部署）
- Reranker 服务（BGE-Reranker-v2，独立部署）

### 1. 启动基础设施
```bash
docker-compose up -d
```

### 2. 配置环境变量
```bash
cp backend/.env.example backend/.env
# 编辑 backend/.env，填入 DEEPSEEK_API_KEY 等
```

### 3. 启动后端
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

### 4. 启动前端
```bash
cd frontend
npm install
npm run dev
```

## 后续优化项

### 短期（P0）
- [ ] `.env` 加入 `.gitignore`，清理已泄露的密钥
- [ ] `server.exe` / `server.exe~` 加入 `.gitignore`
- [ ] 修复会话历史加载（`selectSession` 调用 `getSession` API）
- [ ] 集成 `checkQuery` 自动 Agent 检测
- [ ] 修复 MarkdownRenderer CSS（`border-l-3` → `border-l-2`）
- [ ] 移除未使用的 `rehype-raw` 依赖或实际使用

### 中期（P1）
- [ ] Agent 规划器升级为 LLM 驱动（替换关键词匹配）
- [ ] ReAct 思考步骤改为 LLM 生成
- [ ] Query 改写
- [ ] 引用内联链接（Markdown `[参考N]` 可点击跳转到 CitationCard）
- [ ] 后端单元测试 + 集成测试
- [ ] 优雅关闭（SIGTERM 处理）
- [ ] `/health` 端点检查依赖健康状态
- [ ] 移动端响应式适配（Agent/Citation 面板在小屏不可见）

### 长期（P2）
- [ ] 中医知识库文档预处理管线（PDF 解析 + 语义分块）
- [ ] PostgreSQL 持久化对话历史
- [ ] 用户登录/注册（JWT 鉴权完善）
- [ ] 全文+向量混合检索
- [ ] LLM 私有化部署（vLLM）
- [ ] CI/CD + 自动化测试 + 监控告警
- [ ] 中医知识图谱（症状-药材-方剂关联）
- [ ] 多 Agent 协作
