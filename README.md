# LLM Account Pool

LLM API 调用号池管理系统。该系统作为中间层代理，统一管理多个上游 LLM API 源，对外提供兼容 OpenAI 格式的统一 API 接口，并具备额度管控与负载均衡功能。

## 功能特性

- **用户认证**：单用户管理模式，支持登录、修改账号和密码
- **对外模型管理**：创建和管理供客户端调用的模型配置，可在模型详情中管理对应的请求源
- **请求源管理**：管理上游 LLM API 请求源，支持：
  - 按次计费（限制调用次数）
  - 按 Token 计费（限制 Input+Output Token 总量）
  - 自动重置间隔配置
  - 定时重置（指定每天固定时间重置）
- **API Key 管理**：生成和管理客户端调用密钥，记录详细用量统计
- **负载均衡策略**：
  - 轮询切换（Round Robin）
  - 用完切换（Sequential）
- **自动熔断**：当请求源达到额度限制时自动切换到下一个可用源
- **用量统计**：按模型维度展示各请求源的使用情况

## 技术栈

- **后端**：Go + Gin
- **数据库**：SQLite / PostgreSQL / MySQL / MariaDB + GORM
- **前端**：React + TypeScript + Ant Design + Vite

## 快速开始

> **注意**：本项目使用 SQLite 驱动时需要启用 CGO。

### 前置要求

- Go 1.21+
- **CGO 启用**（SQLite 模式必需）

#### 启用 CGO

```bash
# Linux/macOS（通常默认启用）
export CGO_ENABLED=1

# Windows
# 需要安装 GCC 编译器（推荐 MSYS2 或 TDM-GCC）
# 然后设置：
set CGO_ENABLED=1
```

#### 安装 GCC（Windows）

1. 下载并安装 [MSYS2](https://www.msys2.org/)
2. 安装 GCC：`pacman -S mingw-w64-x86_64-gcc`
3. 将 `C:\msys64\mingw64\bin` 添加到系统 PATH
4. 验证：`gcc --version`

## 环境配置

启动前需要配置以下环境变量：

### 基础配置

| 环境变量 | 必填 | 默认值 | 说明 |
|---------|------|--------|------|
| `JWT_SECRET` | 是 | - | JWT 签名密钥，建议使用随机字符串 |
| `SERVER_PORT` | 否 | 8080 | 服务监听端口 |
| `SERVER_HOST` | 否 | http://localhost:8080 | 服务器地址，用于生成代理地址 |
| `ALLOWED_ORIGINS` | 否 | - | 允许的跨域来源，多个用逗号分隔 |
| `MAX_LOGIN_ATTEMPTS` | 否 | 5 | 最大登录失败次数 |
| `LOCKOUT_DURATION` | 否 | 15 | 登录锁定时长（分钟） |

### 数据库配置

| 环境变量 | 必填 | 默认值 | 说明 |
|---------|------|--------|------|
| `DB_TYPE` | 否 | sqlite | 数据库类型：`sqlite` / `postgres` / `mysql` |
| `DATABASE_URL` | 否 | ../data/llmaccountpool.db | SQLite 数据库文件路径（仅 SQLite 模式） |
| `DB_HOST` | 否 | localhost | PostgreSQL/MySQL 主机地址 |
| `DB_PORT` | 否 | 5432 | PostgreSQL/MySQL 端口（PostgreSQL 默认 5432，MySQL 默认 3306） |
| `DB_USER` | 否 | postgres | 数据库用户名 |
| `DB_PASSWORD` | 否 | - | 数据库密码 |
| `DB_NAME` | 否 | llmaccountpool | 数据库名称 |
| `DB_SSLMODE` | 否 | disable | PostgreSQL SSL 模式 |

### 连接池配置（仅 PostgreSQL/MySQL）

| 环境变量 | 必填 | 默认值 | 说明 |
|---------|------|--------|------|
| `DB_MAX_OPEN_CONNS` | 否 | 100 | 最大打开连接数 |
| `DB_MAX_IDLE_CONNS` | 否 | 20 | 最大空闲连接数 |
| `DB_CONN_MAX_LIFETIME` | 否 | 300 | 连接最大生命周期（秒） |
| `DB_CONN_MAX_IDLE_TIME` | 否 | 60 | 连接最大空闲时间（秒） |

### SQLite 优化配置（仅 SQLite 模式）

| 环境变量 | 必填 | 默认值 | 说明 |
|---------|------|--------|------|
| `DB_ENABLE_WAL_MODE` | 否 | false | 启用 WAL 日志模式 |
| `DB_BUSY_TIMEOUT` | 否 | 5000 | 忙超时时间（毫秒） |

### 启动命令

#### 使用 SQLite（默认）

```bash
# Linux/macOS
export CGO_ENABLED=1
export JWT_SECRET="your-secret-key"
cd backend
go run main.go

# Windows (PowerShell)
$env:CGO_ENABLED="1"
$env:JWT_SECRET="your-secret-key"
cd backend
go run main.go
```

> **注意**：如果编译时遇到 SQLite 相关错误，请确保已安装 GCC 并启用 CGO。

#### 无 CGO 编译（仅使用 PostgreSQL/MySQL）

如果无法启用 CGO，可以使用纯 Go 的 SQLite 驱动或仅使用 PostgreSQL/MySQL：

```bash
# 使用 PostgreSQL 或 MySQL 时可以不启用 CGO
export CGO_ENABLED=0
export DB_TYPE=postgres
# ... 其他配置
go build -o main .
```

#### 使用 PostgreSQL

```bash
export JWT_SECRET="your-secret-key"
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your-password
export DB_NAME=llmaccountpool
cd backend
go run main.go
```

#### 使用 MySQL / MariaDB

```bash
export JWT_SECRET="your-secret-key"
export DB_TYPE=mysql
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your-password
export DB_NAME=llmaccountpool
cd backend
go run main.go
```

或者使用 .env 文件：

```bash
cp .env.example .env
# 编辑 .env 文件配置环境变量
go run main.go
```

## API 接口

### 认证

- `POST /api/login` - 登录

### 管理接口（需认证）

- `GET /api/admin/profile` - 获取管理员信息
- `POST /api/admin/refresh-token` - 刷新 Token
- `POST /api/admin/change-password` - 修改密码
- `POST /api/admin/change-username` - 修改用户名

### 模型管理

- `GET /api/admin/models` - 获取模型列表
- `GET /api/admin/models/:id` - 获取模型详情
- `POST /api/admin/models` - 创建模型
- `PUT /api/admin/models/:id` - 更新模型
- `DELETE /api/admin/models/:id` - 删除模型
- `GET /api/admin/models/template` - 下载Excel导入模板
- `POST /api/admin/models/import` - 从Excel批量导入模型

#### Excel 批量导入

系统支持通过 Excel 文件批量导入对外模型和上游模型：

1. **下载模板**：`GET /api/admin/models/template`
2. **导入模型**：`POST /api/admin/models/import`，通过 form-data 上传 xlsx 文件

**模板说明**：

Excel 文件包含两个工作表：

| 对外模型 sheet | | |
|---------------|---|---|
| 字段 | 必填 | 说明 |
| Name | 是 | 对外模型名称 |
| Model | 是 | 对外模型标识 |
| Strategy | 否 | 策略 (round_robin/sequential)，默认为 round_robin |

| 上游模型 sheet | | |
|---------------|---|---|
| 字段 | 必填 | 说明 |
| ExternalModelName | 是 | 对应的对外模型名称 |
| Name | 是 | 上游模型名称 |
| APIURL | 是 | 上游 API 地址 |
| APIKey | 是 | 上游 API Key |
| ModelName | 是 | 上游模型名称 |
| BillingMode | 否 | 计费模式 (count/tokens)，默认为 count |
| LimitCount | 否 | 调用次数限制 |
| LimitTokens | 否 | Token 限制 |
| LimitResetInterval | 否 | 重置间隔（秒），与 LimitResetTime 二选一 |
| LimitResetTime | 否 | 定时重置时间（HH:MM 格式，如 00:00），与 LimitResetInterval 二选一 |
| IsActive | 否 | 是否启用 (true/false)，默认为 true |

**导入规则**：
- 对外模型：名称相同时更新，不存在则创建
- 上游模型：根据对外模型名称+上游模型名称匹配，存在则更新，不存在则创建

### 请求源管理

- `GET /api/admin/sources` - 获取请求源列表
- `GET /api/admin/sources/:id` - 获取请求源详情
- `POST /api/admin/sources` - 创建请求源
- `PUT /api/admin/sources/:id` - 更新请求源
- `DELETE /api/admin/sources/:id` - 删除请求源
- `POST /api/admin/sources/:id/reset` - 重置请求源用量
- `PATCH /api/admin/sources/:id/name` - 修改请求源名称

### API Key 管理

- `GET /api/admin/keys` - 获取 API Key 列表
- `POST /api/admin/keys` - 创建 API Key（`external_model_id` 为空或 0 时表示可访问全部模型）
- `DELETE /api/admin/keys/:id` - 删除 API Key
- `POST /api/admin/keys/:id/reset` - 重置 API Key 用量

> **说明**：创建 API Key 时，可以选择绑定特定模型或留空（访问全部模型）。绑定特定模型时，该 key 只能用于访问对应模型；留空时，可以根据请求中的 model 名称自动匹配可用模型。

### 用量统计

- `GET /api/admin/usage` - 获取用量统计
- `GET /api/admin/usage/records` - 获取用量记录
- `GET /api/admin/server-info` - 获取服务器信息（包括代理地址）

### 代理接口

- `POST /v1/chat/completions` - OpenAI 兼容的聊天完成接口

Header 中需要携带 `Authorization: Bearer <API_KEY>` 或使用查询参数 `?key=<API_KEY>`

## 项目结构

```
LLMAccountPool/
├── backend/
│   ├── config/          # 配置
│   ├── handlers/        # HTTP 处理器
│   ├── middleware/      # 中间件
│   ├── models/          # 数据模型
│   ├── services/        # 业务逻辑
│   ├── utils/           # 工具函数
│   ├── migrations/      # 数据库迁移脚本
│   ├── main.go          # 入口文件
│   ├── go.mod           # Go 依赖
│   └── .env.example     # 环境变量示例
├── frontend/
│   ├── src/
│   │   ├── components/  # React 组件
│   │   ├── pages/       # 页面组件
│   │   ├── services/    # API 服务
│   │   ├── hooks/       # React Hooks
│   │   ├── types/       # TypeScript 类型
│   │   └── utils/       # 工具函数
│   ├── dist/            # 构建输出
│   └── vite.config.ts   # Vite 配置
├── data/                # 数据存储（SQLite 文件）
├── README.md            # 项目说明
└── DATABASE_MIGRATION.md # 数据库迁移指南
```

## 前端开发

```bash
# 安装依赖
cd frontend
npm install

# 开发模式
npm run dev

# 生产构建
npm run build
```

## 配置

详细配置说明请参考 `backend/.env.example` 文件。

更多数据库配置和迁移指南请查看 [DATABASE_MIGRATION.md](DATABASE_MIGRATION.md)。

## License

MIT
