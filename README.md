# LLM Account Pool

LLM API 调用号池管理系统。该系统作为中间层代理，统一管理多个上游 LLM API 源，对外提供兼容 OpenAI 格式的统一 API 接口，并具备额度管控与负载均衡功能。

## 功能特性

- **用户认证**：单用户管理模式，支持登录、修改账号和密码
- **对外模型管理**：创建和管理供客户端调用的模型配置
- **请求源管理**：管理上游 LLM API 请求源，支持：
  - 按次计费（限制调用次数）
  - 按 Token 计费（限制 Input+Output Token 总量）
  - 自动重置间隔配置
- **API Key 管理**：生成和管理客户端调用密钥，记录详细用量统计
- **负载均衡策略**：
  - 轮询切换（Round Robin）
  - 用完切换（Sequential）
- **自动熔断**：当请求源达到额度限制时自动切换到下一个可用源
- **用量统计**：按模型维度展示各请求源的使用情况

## 技术栈

- **后端**：Go + Gin
- **数据库**：SQLite + GORM
- **前端**：原生 HTML/CSS/JS

## 快速开始

### 编译运行

```bash
cd backend
go build -o llmaccountpool.exe
./llmaccountpool.exe
```

服务启动后访问 `http://localhost:8080`

### 默认账号

- 用户名：`admin`
- 密码：`admin`

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
- `POST /api/admin/keys` - 创建 API Key
- `DELETE /api/admin/keys/:id` - 删除 API Key
- `POST /api/admin/keys/:id/reset` - 重置 API Key 用量

### 用量统计

- `GET /api/admin/usage` - 获取用量统计
- `GET /api/admin/usage/records` - 获取用量记录

### 代理接口

- `POST /v1/chat/completions` - OpenAI 兼容的聊天完成接口

Header 中需要携带 `Authorization: Bearer <API_KEY>` 或使用查询参数 `?key=<API_KEY>`

## 项目结构

```
LLMAccountPool/
├── backend/
│   ├── config/          # 配置
│   ├── handlers/        # HTTP 处理器
│   ├── middleware/     # 中间件
│   ├── models/         # 数据模型
│   ├── services/       # 业务逻辑
│   ├── utils/          # 工具函数
│   ├── main.go         # 入口文件
│   └── go.mod          # Go 依赖
├── frontend/
│   ├── index.html      # 前端页面
│   ├── css/            # 样式文件
│   └── js/             # 前端脚本
└── data/               # 数据存储
```

## 配置

配置文件位于 `backend/config/config.go`，主要配置项：

- `ServerPort`: 服务端口（默认 8080）
- `JWT_SECRET`: JWT 密钥
- `AdminUsername`: 管理员用户名
- `AdminPassword`: 管理员密码（MD5 加密）

## License

MIT
