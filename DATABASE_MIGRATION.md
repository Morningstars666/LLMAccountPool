# 数据库迁移指南

## 概述

本项目支持三种数据库：
- **SQLite** (默认) - 适合开发和小型部署
- **PostgreSQL** - 推荐用于生产环境
- **MySQL/MariaDB** - 替代生产环境选项

## 快速开始

### 使用 SQLite (默认)

无需额外配置，直接运行：

```bash
cd backend
go run .
```

### 使用 PostgreSQL

1. 安装 PostgreSQL 16+

2. 创建数据库：
```sql
CREATE DATABASE llmaccountpool;
CREATE USER llmuser WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE llmaccountpool TO llmuser;
```

3. 配置环境变量：
```bash
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=llmuser
export DB_PASSWORD=your_password
export DB_NAME=llmaccountpool
export DB_SSLMODE=disable
export JWT_SECRET=your-secret-key
```

4. 运行应用：
```bash
go run .
```

### 使用 MySQL

1. 安装 MySQL 8.0+ 或 MariaDB 10.5+

2. 创建数据库：
```sql
CREATE DATABASE llmaccountpool CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'llmuser'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON llmaccountpool.* TO 'llmuser'@'localhost';
FLUSH PRIVILEGES;
```

3. 配置环境变量：
```bash
export DB_TYPE=mysql
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=llmuser
export DB_PASSWORD=your_password
export DB_NAME=llmaccountpool
export JWT_SECRET=your-secret-key
```

4. 运行应用：
```bash
go run .
```

## 连接池配置

以下配置仅对 PostgreSQL 和 MySQL 生效：

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `DB_MAX_OPEN_CONNS` | 100 | 最大打开连接数 |
| `DB_MAX_IDLE_CONNS` | 20 | 最大空闲连接数 |
| `DB_CONN_MAX_LIFETIME` | 300 | 连接最大生命周期 (秒) |
| `DB_CONN_MAX_IDLE_TIME` | 60 | 连接最大空闲时间 (秒) |

## SQLite 优化配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `DB_ENABLE_WAL_MODE` | false | 启用 WAL 日志模式 |
| `DB_BUSY_TIMEOUT` | 5000 | 忙超时时间 (毫秒) |

## 从 SQLite 迁移到 PostgreSQL

### 步骤 1: 导出 SQLite 数据

```bash
cd backend
sqlite3 ../data/llmaccountpool.db ".dump" > sqlite_dump.sql
```

### 步骤 2: 创建 PostgreSQL 数据库

```bash
createdb -U postgres llmaccountpool
```

### 步骤 3: 运行迁移脚本

```bash
psql -U postgres -d llmaccountpool -f migrations/001_postgres.sql
```

### 步骤 4: 数据转换 (可选)

使用第三方工具或手动脚本将 SQLite 数据导入 PostgreSQL。

### 步骤 5: 切换数据库

修改环境变量 `DB_TYPE=postgres` 并重启应用。

## 生产环境建议

### PostgreSQL

```bash
DB_TYPE=postgres
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=prod_user
DB_PASSWORD=strong_password
DB_NAME=llmaccountpool
DB_SSLMODE=require

DB_MAX_OPEN_CONNS=200
DB_MAX_IDLE_CONNS=50
DB_CONN_MAX_LIFETIME=600
DB_CONN_MAX_IDLE_TIME=120
```

### MySQL

```bash
DB_TYPE=mysql
DB_HOST=your-db-host
DB_PORT=3306
DB_USER=prod_user
DB_PASSWORD=strong_password
DB_NAME=llmaccountpool

DB_MAX_OPEN_CONNS=200
DB_MAX_IDLE_CONNS=50
DB_CONN_MAX_LIFETIME=600
DB_CONN_MAX_IDLE_TIME=120
```

## 验证数据库连接

启动应用后，检查日志输出：

```
Using database type: postgres
Connection pool configured: max_open=100, max_idle=20, lifetime=300s, idle_time=60s
PostgreSQL transaction isolation set to READ COMMITTED
Database initialized successfully
```

## 故障排除

### 连接被拒绝

检查防火墙设置和数据库监听地址。

### 认证失败

确认用户名密码正确，检查 `pg_hba.conf` (PostgreSQL) 或用户权限 (MySQL)。

### 超时错误

增加 `DB_BUSY_TIMEOUT` (SQLite) 或调整连接池参数。
