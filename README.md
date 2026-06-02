# 中医智能助手（TCM Agent）

中医知识问答平台：智能诊断 + 养生建议 + 药材查询 + Agent 推理 + 云原生部署，传承中医智慧，赋能健康生活。

## 当前状态

MVP 核心链路已打通：RAG 单轮问答 + Agent 多步推理双模式均可运行，SSE 流式输出、引用卡片、Agent 可视化组件已完整实现。

**2026-06-02 更新**：已完成 14 项代码质量修复和安全加固，包括优雅关闭、健康检查、CORS 配置化、SSE 提前退出、Tailwind CSS 修复、ErrorBoundary、会话历史加载等。Embedding/Reranker 改用 Python 本地部署（`embed_server.py`），解决国内 Docker 镜像拉取受限问题。Agent 规划器当前为关键词匹配（非 LLM 驱动），自动 Agent 检测等功能待完善。详见 [CLAUDE.md](CLAUDE.md)。

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                          Frontend                               │
│  React + Tailwind + Zustand + TanStack Query + ErrorBoundary     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐         │
│  │ Chat UI  │ │Citation  │ │ Session  │ │ Markdown  │         │
│  │          │ │ Panel    │ │ Sidebar  │ │ Renderer  │         │
│  └──────────┘ └──────────┘ └──────────┘ └───────────┘         │
│  ┌─────────────────────────────────────────────────────┐       │
│  │         Agent Visualization Components              │       │
│  │  ┌──────────────┐ ┌──────────────┐ ┌─────────────┐ │       │
│  │  │ReasoningSteps│ │TaskProgress  │ │ToolPanel    │ │       │
│  │  └──────────────┘ └──────────────┘ └─────────────┘ │       │
│  └─────────────────────────────────────────────────────┘       │
│                        │ SSE / HTTP REST                        │
└────────────────────────┼────────────────────────────────────────┘
                         │
┌────────────────────────┼────────────────────────────────────────┐
│                    Backend (Go + Fiber)                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐         │
│  │ Handler  │ │ Service  │ │   RAG    │ │ Middleware │         │
│  │ (Chat)   │ │ (Agent)  │ │ Pipeline │ │ (Auth/Lim)│         │
│  └──────────┘ └──────────┘ └──────────┘ └───────────┘         │
│       │            │            │                                │
│       └────────────┼────────────┘                                │
│                    │                                             │
│  ┌─────────────────┼──────────────────────────────────┐         │
│  │             Agent Core                              │         │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │         │
│  │  │ Planner  │ │ Executor │ │  Tools   │            │         │
│  │  │(关键词)  │ │(ReAct循环)│ │(5个工具) │            │         │
│  │  └──────────┘ └──────────┘ └──────────┘            │         │
│  │  ┌──────────────────────────────────────┐          │         │
│  │  │   TCM Tools                          │          │         │
│  │  │  diagnose_symptom | herb_query       │          │         │
│  │  │  prescription_advice                 │          │         │
│  │  │  wellness_suggestion | acupoint_info │          │         │
│  │  └──────────────────────────────────────┘          │         │
│  └────────────────────────────────────────────────────┘         │
│                    │                                             │
│  ┌─────────────────┼──────────────────────────────────┐         │
│  │             Stores                                 │         │
│  │  ┌──────┐  ┌──────┐  ┌───────────────────┐        │         │
│  │  │Qdrant│  │Redis │  │ Embedding/Reranker│        │         │
│  │  │(向量)│  │(会话)│  │ (Python :8081)    │        │         │
│  │  └──────┘  └──────┘  └───────────────────┘        │         │
│  └──────────────────────────────────────────────────┘         │
│                    │                                             │
│  ┌─────────────────┼──────────────────────────────────┐         │
│  │             External                               │         │
│  │  ┌────────────┐                                    │         │
│  │  │ DeepSeek   │                                    │         │
│  │  │ V4-Pro     │                                    │         │
│  │  └────────────┘                                    │         │
│  └──────────────────────────────────────────────────┘         │
└───────────────────────────────────────────────────────────────┘
```

## 两种工作模式

### RAG 模式（默认）
适合简单问答，直接从知识库检索相关中医文献后生成回答。

```
用户Query → Embedding → Qdrant检索(召回Top-K×2)
    → Reranker重排序(精排Top-K) → 拼接上下文 → LLM生成 → SSE流式输出
```

### Agent 模式（多步推理）
适合复杂问题，自动拆解任务，调用 TCM 工具链收集信息后综合分析。

```
用户Query → QueryAnalyzer(关键词匹配) → TaskPlanner → ReAct执行引擎
    → 工具调用(诊断/药材/方剂/养生/穴位) → LLM合成 → SSE流式输出
```

> 当前 Agent 规划器使用关键词匹配（23 个中文关键词），后续版本将升级为 LLM 驱动的语义分析。

示例工作流：用户问"帮我分析这个症状并给出调理方案"

```
Step 1: 症状辨证 (diagnose_symptom)
  Thought: 需要先了解用户症状，调用 diagnose_symptom 工具进行辨证分析
  Action: diagnose_symptom
  Observation: 检索到 5 条相关中医知识

Step 2: 药材推荐 (herb_query)
  Thought: 根据诊断结果推荐对症药材
  Action: herb_query
  Observation: 检索到 4 条相关中医知识

Step 3: 养生建议 (wellness_suggestion)
  Thought: 给出日常调理建议
  Action: wellness_suggestion
  Observation: 检索到 3 条相关中医知识

Step 4: 整合答案
  Final Answer: 完整的中医调理方案
```

## 项目结构

```
QiHuangZhiyi-tcm-brain-agent/
├── backend/                        # Go + Fiber 后端
│   ├── cmd/server/main.go          # 入口（含优雅关闭）
│   ├── config/config.go            # 配置管理（含 ALLOWED_ORIGINS）
│   ├── embed_server.py             # Python Embedding + Reranker 服务（FastAPI :8081）
│   ├── requirements-embed.txt      # Python 依赖（sentence-transformers, fastapi, uvicorn）
│   ├── .env.example                # 环境变量模板（安全，可提交）
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── chat.go             # Chat + Session HTTP 处理器
│   │   │   └── agent.go            # Agent 专用接口
│   │   ├── service/
│   │   │   ├── chat.go             # ChatService（RAG + LLM 编排）
│   │   │   ├── session.go          # 会话管理
│   │   │   └── agent/coordinator.go
│   │   ├── agent/
│   │   │   ├── planner/            # QueryAnalyzer + TaskPlanner（关键词匹配）
│   │   │   ├── executor/           # ReactExecutor（ReAct 循环 + LLM 合成）
│   │   │   ├── tools/              # Tool 接口 + Registry + 5 个 TCM 工具
│   │   │   └── memory/             # ShortTermMemory（Redis）
│   │   ├── rag/pipeline.go         # RAG Pipeline
│   │   ├── llm/deepseek.go         # DeepSeek API 客户端
│   │   ├── store/                  # Qdrant / Redis / Embedding / Reranker（均含 Close/状态码检查）
│   │   ├── middleware/middleware.go # CORS / Logger / RateLimit / Auth
│   │   └── model/                  # 数据模型
│   └── .env                        # 环境变量（不提交！已被 .gitignore 排除）
├── frontend/                       # React 前端
│   └── src/
│       ├── components/
│       │   ├── Chat/               # ChatContainer, MessageList, ChatInput, MessageBubble, AgentModeToggle
│       │   ├── Agent/              # ReasoningSteps, TaskProgressBar, ToolInvocationPanel, ThoughtBubble
│       │   ├── Citation/           # CitationCard, CitationPanel
│       │   ├── Markdown/           # MarkdownRenderer
│       │   ├── Session/            # SessionSidebar
│       │   ├── Layout/             # AppLayout
│       │   └── ErrorBoundary.tsx   # React 错误边界
│       ├── stores/                 # Zustand: chatStore, sessionStore, agentStore
│       ├── services/               # api.ts, sse.ts, agent.ts
│       ├── hooks/                  # useChat, useSessions, useAgent
│       └── types/                  # index.ts, agent.ts
├── docker-compose.yml              # Qdrant + Redis（Docker Desktop 手动启动）
├── CLAUDE.md                       # 开发指南
└── README.md
```

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 20+
- Python 3.10+（Embedding 服务）
- Docker Desktop（Qdrant + Redis）
- DeepSeek API Key

### 1. 启动基础设施

在 Docker Desktop 中启动 `tcm-qdrant` 和 `tcm-redis` 容器，或通过命令行：

```bash
docker compose up -d
```

### 2. 启动 Embedding 服务

```bash
cd backend

# 首次运行：安装依赖（清华源）
python -m pip install -r requirements-embed.txt -i https://pypi.tuna.tsinghua.edu.cn/simple

# 模型下载走国内镜像
# PowerShell:
$env:HF_ENDPOINT = "https://hf-mirror.com"
# Bash:
# export HF_ENDPOINT=https://hf-mirror.com

# 启动服务（首次自动下载模型 ~400MB）
python embed_server.py
```

默认监听 `http://127.0.0.1:8081`，同时提供 `/embeddings` 和 `/rerank`。
可用 `--embed-only` 仅启动 Embedding，`--rerank-only` 仅启动 Reranker。

### 3. 环境变量

```bash
# 从模板创建配置文件
cp backend/.env.example backend/.env
# 编辑 backend/.env，填入 DEEPSEEK_API_KEY 和 JWT_SECRET
```

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `DEEPSEEK_API_KEY` | ✅ | - | DeepSeek API 密钥 |
| `JWT_SECRET` | ✅ | - | JWT 签名密钥（`openssl rand -hex 32`） |
| `SERVER_PORT` | - | `8080` | 后端服务端口 |
| `ALLOWED_ORIGINS` | - | `http://localhost:3000` | CORS 允许的前端来源 |
| `QDRANT_HOST` | - | `localhost` | Qdrant 主机地址 |
| `REDIS_ADDR` | - | `localhost:6379` | Redis 地址 |
| `EMBEDDING_URL` | - | `http://localhost:8081/embeddings` | Embedding 服务（Python embed_server.py） |
| `RERANKER_URL` | - | `http://localhost:8081/rerank` | Reranker 服务（同进程） |

### 4. 启动后端

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

服务启动在 `http://localhost:8080`。按 `Ctrl+C` 触发优雅关闭（10s 超时）。

### 5. 启动前端

```bash
cd frontend
npm install
npm run dev
```

开发服务器启动在 `http://localhost:3000`，自动代理 API 到后端。

## API 接口

### POST /api/v1/chat

非流式对话（RAG 模式）。

```json
// Request
{ "query": "肝肾阴虚如何调理？", "session_id": "optional-uuid" }

// Response
{
  "session_id": "uuid",
  "answer": "肝肾阴虚是...",
  "citations": [
    { "book_title": "中医基础理论", "chapter": "脏腑辨证", "content": "...", "score": 0.92, "category": "症状" }
  ]
}
```

### POST /api/v1/chat/stream

SSE 流式对话（RAG 模式）。事件类型：

- `meta` — 返回 session_id
- `citations` — 返回引用列表
- `token` — 流式 token
- `done` — 生成完成
- `error` — 错误信息

### POST /api/v1/agent/analyze

分析查询是否需要 Agent 模式。

```json
// Request
{ "query": "帮我分析这个症状并给出调理方案" }

// Response
{ "needs_agent": true, "reason": "需要多步推理，使用Agent模式" }
```

### POST /api/v1/agent/chat/stream

Agent 模式 SSE 流式对话。事件类型：

- `meta` — 返回 session_id（含 mode: "agent"）
- `task_progress` — 任务规划 + 任务进度更新
- `thought` — ReAct 思考步骤
- `tool_call` — 工具调用（名称 + 参数）
- `tool_result` — 工具执行结果
- `final_answer` — LLM 合成的最终答案
- `error` — 错误信息

### GET /api/v1/sessions

获取所有会话列表。

### GET /api/v1/sessions/:id

获取指定会话详情。

### DELETE /api/v1/sessions/:id

删除指定会话。

### GET /health

健康检查，返回依赖状态：

```json
{ "status": "ok", "details": { "qdrant": "ok", "redis": "ok" } }
// 或降级时：
{ "status": "degraded", "details": { "qdrant": "unavailable", "redis": "ok" } }
```

## 技术栈

| 层级 | 技术 | 用途 |
|------|------|------|
| 后端框架 | Go + Fiber v2.52.10 | HTTP 服务 + SSE 流式 |
| LLM | DeepSeek-V4-Pro | 大模型底座 |
| 向量库 | Qdrant | 知识库向量存储与检索 |
| 缓存 | Redis | 会话缓存 + 短期上下文 |
| Embedding | bge-small-zh-v1.5 | 文本向量化（Python sentence-transformers, :8081） |
| Reranker | bge-reranker-base | 召回重排序（同进程 :8081/rerank） |
| Agent 引擎 | ReAct 模式 | 多步推理 + 工具调用（规划器为关键词匹配） |
| 前端框架 | React 18 | UI |
| 样式 | Tailwind CSS 3 | 原子化 CSS |
| 状态管理 | Zustand 4 | 客户端状态（含 Agent 状态） |
| 数据管理 | TanStack Query 5 | 服务端状态缓存 |
| Markdown | react-markdown 9 + remark-gfm | LLM 输出渲染 |

## TCM 工具集

| 工具名 | 功能 | 说明 |
|--------|------|------|
| diagnose_symptom | 症状辨证 | 根据症状进行中医辨证分析，判断证型 |
| herb_query | 药材查询 | 查询中药材性味归经、功效主治、配伍禁忌 |
| prescription_advice | 方剂推荐 | 根据证型推荐经典方剂，含组成、功效、加减 |
| wellness_suggestion | 养生建议 | 提供饮食、起居、情志、运动等养生指导 |
| acupoint_info | 穴位指导 | 查询穴位定位、经络归属、按摩方法 |

所有工具均复用 RAG Pipeline 进行知识检索。

## 设计决策

### MVP 阶段取舍
- **已实现**：RAG 单轮问答 + Agent 多步推理双模式
- **已实现**：核心问答可用、引用可追踪、SSE 流式响应
- **已实现**：5 个 TCM 工具覆盖诊断/药材/方剂/养生/穴位场景
- **已实现**：Agent 可视化组件（推理步骤 + 任务进度 + 工具调用面板）
- **已实现**：优雅关闭、健康检查、ErrorBoundary 等生产基础
- **已实现**：Python 本地 Embedding/Reranker 服务（避免国内 Docker 镜像拉取问题）
- **未实现**：PostgreSQL 持久化，先用 Redis 跑通核心链路
- **未实现**：用户登录/注册，先做免登录体验
- **未实现**：全文检索（Elasticsearch），Qdrant 单用足够

### 上下文窗口管理
- 对话历史仅保留最近 10 条（Redis TTL 24 小时）
- 总 token 超过 ~6000 时自动裁剪早期历史
- 单会话最多 50 条消息

### Agent 工具设计原则
- 工具层复用 RAG Pipeline，不在工具内独立检索
- 每个工具返回检索到的知识文本 + 引用计数
- LLM 基于工具收集的知识综合生成最终答案

## 后续优化项

### 短期（完善 MVP）
- [x] `.env` + `*.ps1` 加入 `.gitignore`，创建 `.env.example` 安全模板
- [x] 修复会话历史加载（`selectSession` 调用 `getSession` API）
- [x] 修复 MarkdownRenderer Tailwind CSS 类
- [x] 移除死依赖 `rehype-raw`
- [x] Qdrant gRPC 升级 + 优雅关闭 + Health 检查 + CORS 配置化
- [x] Agent SSE done/error 提前退出 + EmbeddingClient 状态码检查
- [x] React ErrorBoundary 错误边界
- [x] ReAct `buildArgs` 使用 `prevResults` 传递工具间上下文
- [x] IME 输入法兼容 + isStreaming 并发修复 + Zustand 选择器优化
- [x] TCM 工具参数组合一致性 + nil ragPipe 防护
- [x] Agent 可视化组件 key 稳定性 + Agent 状态清理
- [x] Python `embed_server.py` 本地 Embedding + Reranker 服务
- [ ] 集成自动 Agent 检测（`checkQuery` 已实现但未接入 UI）
- [ ] 中医知识库文档预处理管线（PDF 解析 + 语义分块）
- [ ] 知识库批量导入脚本
- [ ] Query 改写（提升检索召回率）
- [ ] 引用内联链接（Markdown 中 `[参考N]` 可点击跳转）
- [ ] 移动端响应式适配

### 中期（增强体验）
- [ ] Agent 规划器升级为 LLM 驱动（替换关键词匹配）
- [ ] ReAct 思考步骤改为 LLM 生成
- [ ] 后端单元测试 + 集成测试
- [ ] 用户登录/注册（JWT 鉴权完善）
- [ ] PostgreSQL 持久化对话历史
- [ ] 全文+向量混合检索（Qdrant + ES）
- [ ] 对话导出（Markdown/PDF）
- [ ] 反馈机制（回答有用/无用）
- [ ] 中医古籍 OCR 接入（PaddleOCR）

### 长期（生产就绪）
- [ ] LLM 私有化部署（vLLM 替代 API 调用）
- [ ] Embedding/Reranker 模型升级为 bge-m3（替换 bge-small）
- [ ] CI/CD 流水线 + 自动化测试
- [ ] 监控告警（Prometheus + Grafana）
- [ ] 中医知识图谱（症状-药材-方剂关联）
- [ ] 知识库自动更新（新文献入库 → 自动分块 → 向量化）
- [ ] 多 Agent 协作（诊断 Agent + 用药 Agent + 养生 Agent）
