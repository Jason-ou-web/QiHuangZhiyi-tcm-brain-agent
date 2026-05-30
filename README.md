# 中医智能助手（TCM Agent）

中医知识问答平台：智能诊断 + 养生建议 + 药材查询 + Agent 推理 + 云原生部署，传承中医智慧，赋能健康生活。

## 当前状态

MVP 核心链路已打通：RAG 单轮问答 + Agent 多步推理双模式均可运行，SSE 流式输出、引用卡片、Agent 可视化组件已完整实现。Agent 规划器当前为关键词匹配（非 LLM 驱动），会话历史加载、自动 Agent 检测等功能待完善。详见 [CLAUDE.md](CLAUDE.md)。

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                          Frontend                               │
│  React + Tailwind + Zustand + TanStack Query                    │
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
│  │  ┌──────┐  ┌──────┐  ┌───────────┐               │         │
│  │  │Qdrant│  │Redis │  │ Embedding │               │         │
│  │  │(向量)│  │(会话)│  │ (BGE-M3)  │               │         │
│  │  └──────┘  └──────┘  └───────────┘               │         │
│  └──────────────────────────────────────────────────┘         │
│                    │                                             │
│  ┌─────────────────┼──────────────────────────────────┐         │
│  │             External                               │         │
│  │  ┌────────────┐  ┌───────────────┐               │         │
│  │  │ DeepSeek   │  │  Reranker     │               │         │
│  │  │ V4-Pro     │  │(BGE-Reranker) │               │         │
│  │  └────────────┘  └───────────────┘               │         │
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
agri-qa-system/
├── backend/                        # Go + Fiber 后端
│   ├── cmd/server/main.go          # 入口
│   ├── config/config.go            # 配置管理
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
│   │   ├── store/                  # Qdrant / Redis / Embedding / Reranker
│   │   ├── middleware/middleware.go # CORS / Logger / RateLimit / Auth
│   │   └── model/                  # 数据模型
│   └── .env                        # 环境变量（不提交！）
├── frontend/                       # React 前端
│   └── src/
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
├── docker-compose.yml              # Qdrant + Redis
├── CLAUDE.md                       # 开发指南
└── README.md
```

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

### 2. 环境变量

```bash
# 必填
export DEEPSEEK_API_KEY="sk-xxx"

# 可选（有默认值）
export SERVER_PORT="8080"
export DEEPSEEK_MODEL="deepseek-chat"
export QDRANT_HOST="localhost"
export QDRANT_PORT="6334"
export QDRANT_COLLECTION="tcm_knowledge"
export REDIS_ADDR="localhost:6379"
export EMBEDDING_URL="http://localhost:8081/embed"
export RERANKER_URL="http://localhost:8082/rerank"
export JWT_SECRET="your-secret-here"
```

### 3. 启动后端

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

服务启动在 `http://localhost:8080`。

### 4. 启动前端

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

## 技术栈

| 层级 | 技术 | 用途 |
|------|------|------|
| 后端框架 | Go + Fiber v2.52.10 | HTTP 服务 + SSE 流式 |
| LLM | DeepSeek-V4-Pro | 大模型底座 |
| 向量库 | Qdrant | 知识库向量存储与检索 |
| 缓存 | Redis | 会话缓存 + 短期上下文 |
| Embedding | BGE-M3 / stella-m3 | 文本向量化（独立部署） |
| Reranker | BGE-Reranker-v2 | 召回重排序 |
| Agent 引擎 | ReAct 模式 | 多步推理 + 工具调用（规划器为关键词匹配） |
| 前端框架 | React 18 | UI |
| 样式 | Tailwind CSS 3 | 原子化 CSS |
| 状态管理 | Zustand 4 | 客户端状态（含 Agent 状态） |
| 数据管理 | TanStack Query 5 | 服务端状态缓存 |
| Markdown | react-markdown 9 | LLM 输出渲染 |

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
- [ ] 修复会话历史加载（`selectSession` 需调用 `getSession` API）
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
- [ ] Embedding/Reranker 服务容器化
- [ ] CI/CD 流水线 + 自动化测试
- [ ] 监控告警（Prometheus + Grafana）
- [ ] 中医知识图谱（症状-药材-方剂关联）
- [ ] 知识库自动更新（新文献入库 → 自动分块 → 向量化）
- [ ] 多 Agent 协作（诊断 Agent + 用药 Agent + 养生 Agent）
