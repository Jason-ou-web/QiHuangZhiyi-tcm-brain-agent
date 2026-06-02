## 项目目标
构建中医智能助手（TCM Agent），支持基于中医古籍和现代文献的 Agentic RAG 问答，重点满足中医知识检索、智能诊断、养生建议、药材查询、引用溯源和流式对话体验。传承中医智慧，赋能健康生活。

## 开发原则
- 优先实现 MVP 闭环
- 先做最小可用功能，再逐步增强
- 保持代码简洁，避免过度抽象
- 优先复用现有结构，减少无关改动
- 改动后先做最小验证
- 沟通尽量简洁，减少 token 消耗

## 当前实现状态（2026-06-02）

### 已完成
- Go + Fiber 后端框架，所有路由和中间件就绪
- DeepSeek API 封装（同步 + SSE 流式）
- RAG Pipeline：Embedding → Qdrant 检索(召回 topK×2) → Reranker 重排序 → 上下文拼接 → LLM 生成
- Chat SSE 流式输出（token / citations / done / error 事件）
- Agent 模式：QueryAnalyzer + TaskPlanner + ReactExecutor + 5 个 TCM 工具 + 最终 LLM 合成
- Agent SSE 流式输出（meta / thought / task_progress / tool_call / tool_result / final_answer 事件）
- Redis 会话存储（Lua 脚本原子追加 + 自动裁剪到 50 条 + TTL 24h）
- Qdrant 向量检索（gRPC 客户端，已升级为 `grpc.NewClient` 替代废弃 API）
- Embedding / Reranker HTTP 客户端（Infinity/OpenAI 兼容格式，含 HTTP 状态码检查）
- JWT 中间件（MVP 阶段免登录放行）
- CORS / 日志 / 限流中间件（CORS 来源通过 `ALLOWED_ORIGINS` 环境变量配置）
- 优雅关闭（SIGINT/SIGTERM → 10s 超时 → 关闭 Qdrant/Redis 连接）
- `/health` 端点检查 Qdrant/Redis 依赖健康状态（返回 `ok` 或 `degraded`）
- Store 层 `Close()` 方法（QdrantStore / RedisStore）
- React 前端：ChatContainer / MessageList / ChatInput / MessageBubble / MarkdownRenderer
- ErrorBoundary 错误边界组件（防止 React 渲染错误白屏）
- 引用卡片（CitationCard + CitationPanel，可展开原文预览）
- Agent 可视化组件（ReasoningSteps / TaskProgressBar / ToolInvocationPanel / ThoughtBubble）完整实现
- 会话侧边栏（SessionSidebar，新建/切换/删除/加载历史消息）
- Zustand 状态管理（chat / session / agent 三个 store）
- SSE 流式消费（RAG 和 Agent 两种模式独立处理，Agent 模式支持 done/error 提前退出）
- Agent 模式手动切换（AgentModeToggle）
- docker-compose.yml（Qdrant + Redis，国内网络环境建议手动在 Docker Desktop 启动容器）
- Python Embedding 服务（`backend/embed_server.py`，FastAPI + sentence-transformers，单端口 8081 同时提供 `/embeddings` 和 `/rerank`，支持 hf-mirror.com 国内镜像下载模型）
- `.env.example` 环境变量模板文件（已更新 Embedding/Reranker URL 指向本地 Python 服务器）
- `.gitignore` 覆盖 `.env`、`*.ps1`、`*.exe` 等敏感/构建文件

### 与计划的差距
- **Agent 规划器是关键词匹配，非 LLM 驱动**：QueryAnalyzer 检查 23 个中文关键词（如"分析""辨证""调理"），>=2 个匹配才进入 Agent 模式。TaskPlanner 同样是关键词→工具映射，非 LLM 推理。
- **ReAct 思考步骤是硬编码模板**：execute 循环中的 thought 是固定字符串，不是 LLM 生成。LLM 仅在最终合成步骤调用。
- **Query 改写未实现**：RAG Pipeline 没有 Query Rewriting 步骤，用户原始 query 直接送入 Embedding。
- **自动 Agent 检测未集成**：`useAgent.checkQuery` 已实现但无组件调用，Agent 模式全靠手动切换。
- **引用未内联到 Markdown**：LLM 输出中的 `[参考N]` 标记和 CitationCard 之间没有可点击链接。
- **前端 `queryClient.invalidateQueries` 未调用**：SSE 流结束后未同步 TanStack Query 缓存。
- **无测试**：后端 0 个 `*_test.go` 文件，前端无测试框架配置。
- **无文档预处理管线**：PDF 解析、语义分块、古籍 OCR 均为规划阶段，无代码。
- **长期记忆未实现**：仅有 Redis 短期记忆（TTL 24h）。

### 已知问题
- **CRITICAL**: `backend/.env` 和 `backend/start.ps1` 曾含真实 API Key 和 JWT Secret。`.gitignore` 已更新排除 `.env` 和 `*.ps1`，但需确认 Git 历史已清理且密钥已轮换。
- `server.exe` (30MB) 不应提交到仓库（`.gitignore` 已含 `*.exe`）。
- 前端 `sessionStore.selectSession` 已修复：调用 `getSession(id)` 加载消息历史并填充 chatStore。
- ~~`rehype-raw` 依赖已安装但从未 import~~ → 已移除。
- ~~`MarkdownRenderer` 中 `border-l-3` 和 `border-primary-400` 不是合法 Tailwind class~~ → 已修复为 `border-l-2`，`primary-400` 已添加到 Tailwind 配置。
- ~~Agent SSE 解析不检查 `done`/`error` 事件提前退出~~ → 已修复。
- ~~Qdrant 连接使用已废弃的 `grpc.WithBlock()`~~ → 已升级为 `grpc.NewClient()`。
- ~~无优雅关闭（SIGTERM 处理）~~ → 已实现。
- ~~`/health` 端点不检查依赖健康状态~~ → 已修复。
- ~~无 ErrorBoundary~~ → 已添加。
- ~~EmbeddingClient 不检查 HTTP 状态码~~ → 已修复。
- ~~CORS 来源硬编码 `localhost:3000`~~ → 已改为 `ALLOWED_ORIGINS` 环境变量。
- ~~`sendSSEError` 缺少 `X-Accel-Buffering` 头~~ → 已修复。
- ~~ReAct `buildArgs` 忽略 `prevResults`，多步推理各工具无法共享上下文~~ → 已修复。
- ~~`emitAgentEvent` 在 channel 关闭时 panic~~ → 已添加 `recover()` 防护。
- ~~前端 IME 输入法 Enter 键误触发提交~~ → 已添加 `onCompositionStart/End` 处理。
- ~~`isStreaming` 模式门控允许 Agent+RAG 并发流~~ → 已改为 `chatStreaming || agentStreaming` 联合判断。
- ~~`ToolInvocationPanel`/`CitationPanel` 使用 `key={i}` 导致展开状态错位~~ → 已改用稳定 key。
- ~~`AgentModeToggle`/`ChatContainer` 未使用 Zustand 选择器导致过度渲染~~ → 已优化。
- ~~TCM 工具中 `if/else-if` 链导致多参数时后项被丢弃~~ → herb/prescription/acupoint 已改为独立 if 组合。
- ~~`RetrieveKnowledge` 无 nil ragPipe 防护~~ → 已添加 nil 检查。
- ~~`ThoughtBubble` 全气泡脉冲 + 空格内容不拦截~~ → 已改为指示点脉冲 + `.trim()` 检查。
- ~~`types/agent.ts` 中 `AgentEvent.type` 缺少 `'done'`~~ → 已添加。
- ~~`toggleAgentMode` 未清理旧 Agent 状态~~ → 离开 Agent 模式时自动 `resetAgent()`。
- `ReactExecutor.synthesize` 调用 `e.llm.Chat(messages)` 不传 context（待修复，需扩增 LLM 客户端接口）。

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
| bge-small-zh-v1.5 | Embedding 模型，Python `embed_server.py` 本地部署（sentence-transformers + FastAPI :8081） | MVP |
| bge-reranker-base | 召回后重排序，Python `embed_server.py` 同进程部署（:8081/rerank） | MVP |
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

## 项目结构（实际）
```
agri-qa-system/
├── backend/
│   ├── cmd/server/main.go          # 入口：初始化 stores/services，注册 Fiber 路由，优雅关闭
│   ├── config/config.go            # 环境变量配置（含 ALLOWED_ORIGINS）
│   ├── .env.example                # 环境变量模板（安全，可提交）
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
│   │   │   ├── qdrant.go           # Qdrant gRPC 客户端（grpc.NewClient + Close()）
│   │   │   ├── redis.go            # Redis 会话存储（含 Close()）
│   │   │   └── embedding.go        # Embedding + Reranker HTTP 客户端（Infinity/OpenAI 格式）
│   ├── embed_server.py             # Python Embedding + Reranker 服务（FastAPI :8081）
│   ├── requirements-embed.txt       # Python 依赖（sentence-transformers, fastapi, uvicorn）
│   │   ├── middleware/middleware.go # CORS / Logger / RateLimit / Auth
│   │   └── model/
│   │       ├── model.go            # 核心类型
│   │       └── agent.go            # Agent 类型
│   └── .env                        # 环境变量（含密钥，不提交！已被 .gitignore 排除）
├── frontend/
│   └── src/
│       ├── main.tsx                # React 入口
│       ├── App.tsx                 # QueryClientProvider + ErrorBoundary 包装
│       ├── index.css               # Tailwind + 自定义样式
│       ├── components/
│       │   ├── Chat/               # ChatContainer, MessageList, ChatInput, MessageBubble, AgentModeToggle
│       │   ├── Agent/              # ReasoningSteps, TaskProgressBar, ToolInvocationPanel, ThoughtBubble
│       │   ├── Citation/           # CitationCard, CitationPanel
│       │   ├── Markdown/           # MarkdownRenderer
│       │   ├── Session/            # SessionSidebar
│       │   ├── Layout/             # AppLayout
│       │   └── ErrorBoundary.tsx   # React 错误边界（防止白屏）
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
- Python 3.10+（Embedding 服务）
- Docker Desktop（Qdrant + Redis，手动点击 Start 启动容器）
- DeepSeek API Key

### 1. 启动基础设施（Qdrant + Redis）

在 Docker Desktop 中手动启动 `tcm-qdrant` 和 `tcm-redis` 容器。
或通过命令行：`docker compose up -d`

### 2. 启动 Embedding 服务

```bash
cd backend
# 首次运行：安装依赖（走清华源）
python -m pip install -r requirements-embed.txt -i https://pypi.tuna.tsinghua.edu.cn/simple

# 设置模型下载走国内镜像
$env:HF_ENDPOINT = "https://hf-mirror.com"

# 启动（首次会下载模型，bge-small-zh-v1.5 ~100MB + bge-reranker-base ~300MB）
python embed_server.py
```

默认监听 `http://127.0.0.1:8081`，同时提供 `/embeddings` 和 `/rerank`。
可用 `--embed-only` 仅启动 Embedding，`--rerank-only` 仅启动 Reranker。

### 3. 配置环境变量
```bash
cp backend/.env.example backend/.env
# 编辑 backend/.env，填入 DEEPSEEK_API_KEY 和 JWT_SECRET
```

环境变量说明：

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `DEEPSEEK_API_KEY` | ✅ | - | DeepSeek API 密钥 |
| `JWT_SECRET` | ✅ | - | JWT 签名密钥，用 `openssl rand -hex 32` 生成 |
| `SERVER_PORT` | - | `8080` | 后端服务端口 |
| `DEEPSEEK_MODEL` | - | `deepseek-chat` | LLM 模型名 |
| `DEEPSEEK_BASE_URL` | - | `https://api.deepseek.com` | API 地址 |
| `QDRANT_HOST` | - | `localhost` | Qdrant 主机 |
| `QDRANT_PORT` | - | `6334` | Qdrant gRPC 端口 |
| `QDRANT_COLLECTION` | - | `tcm_knowledge` | 向量集合名 |
| `REDIS_ADDR` | - | `localhost:6379` | Redis 地址 |
| `EMBEDDING_URL` | - | `http://localhost:8081/embeddings` | Embedding 服务地址（Python embed_server.py） |
| `RERANKER_URL` | - | `http://localhost:8081/rerank` | Reranker 服务地址（同进程） |
| `ALLOWED_ORIGINS` | - | `http://localhost:3000` | CORS 允许的前端来源 |

### 4. 启动后端
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```
服务启动在 `http://localhost:8080`。按 `Ctrl+C` 会触发优雅关闭。

### 5. 启动前端
```bash
cd frontend
npm install
npm run dev
```
开发服务器启动在 `http://localhost:3000`，自动代理 API 到后端。

## 后续优化项

### 短期（P0）
- [x] `.env` 加入 `.gitignore`，清理已泄露的密钥
- [x] `server.exe` / `server.exe~` 加入 `.gitignore`
- [x] 修复会话历史加载（`selectSession` 调用 `getSession` API）
- [x] 修复 MarkdownRenderer CSS（`border-l-3` → `border-l-2`，添加 `primary-400`）
- [x] 移除未使用的 `rehype-raw` 依赖
- [x] Qdrant 连接升级为 `grpc.NewClient`（替代废弃的 `grpc.WithBlock`）
- [x] 优雅关闭（SIGTERM 处理）
- [x] `/health` 端点检查依赖健康状态
- [x] CORS 来源改为环境变量 `ALLOWED_ORIGINS`
- [x] Agent SSE 流 done/error 提前退出
- [x] EmbeddingClient 添加 HTTP 状态码检查
- [x] Store 层添加 Close() 方法
- [x] 添加 React ErrorBoundary 错误边界
- [x] 创建 `.env.example` 安全模板
- [x] `.gitignore` 添加 `*.ps1` 排除
- [x] ReAct `buildArgs` 使用 `prevResults` 传递工具间上下文
- [x] `emitAgentEvent` channel 关闭 panic 防护
- [x] 前端 IME 输入法兼容 + isStreaming 并发修复 + Zustand 选择器优化
- [x] TCM 工具参数组合一致性修复 + nil ragPipe 防护
- [x] `types/agent.ts` 补充 `'done'` 类型 + `toggleAgentMode` 清理状态
- [ ] 集成 `checkQuery` 自动 Agent 检测
- [ ] 清理 `start.ps1` 和 `.env` 的 Git 历史 + 轮换密钥

### 中期（P1）
- [ ] Agent 规划器升级为 LLM 驱动（替换关键词匹配）
- [ ] ReAct 思考步骤改为 LLM 生成
- [ ] Query 改写
- [ ] 引用内联链接（Markdown `[参考N]` 可点击跳转到 CitationCard）
- [ ] 后端单元测试 + 集成测试
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
