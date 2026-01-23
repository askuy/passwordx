# PasswordX - 密码管理工具

一个类似1Password的密码管理系统，支持多租户、AES-256加密、OAuth登录，包含Web端和Chrome扩展。

## 技术栈

- **后端**: Go + EGO框架 + MySQL
- **前端**: React + TypeScript + Vite + TailwindCSS
- **Chrome扩展**: React + TypeScript + Manifest V3
- **加密**: AES-256-GCM + PBKDF2密钥派生

## 项目结构

```
passwordx/
├── backend/                # Go后端
│   ├── cmd/server/        # 入口文件
│   ├── config/            # 配置文件
│   └── internal/          # 内部模块
├── frontend/              # React Web应用
│   └── src/
│       ├── pages/         # 页面组件
│       ├── components/    # 通用组件
│       ├── services/      # API服务
│       ├── stores/        # 状态管理
│       └── utils/         # 工具函数
└── chrome-extension/      # Chrome扩展
    └── src/
        ├── popup/         # 扩展弹窗
        ├── background/    # 后台脚本
        └── content/       # 内容脚本
```

## 快速开始

### 一键启动（开发环境）

```bash
# 1. 创建数据库
mysql -u root -p -e "CREATE DATABASE passwordx CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. 启动后端（终端1）
cd backend && go mod tidy && go run cmd/server/main.go

# 3. 启动前端（终端2）
cd frontend && npm install && npm run dev

# 4. 构建Chrome扩展（终端3）
cd chrome-extension && npm install && npm run build
```

**访问地址：**
- 前端: http://localhost:3000
- 后端API: http://localhost:8080
- Chrome扩展: 在 `chrome://extensions/` 加载 `chrome-extension/dist` 目录

---

### 1. 数据库准备

创建MySQL数据库:

```sql
CREATE DATABASE passwordx CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 启动后端

```bash
cd backend

# 安装依赖
go mod tidy

# 修改配置文件 config/config.toml 中的数据库连接信息

# 启动服务
go run cmd/server/main.go
```

后端默认运行在 http://localhost:8080

### 3. 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端默认运行在 http://localhost:3000

### 4. 构建Chrome扩展

```bash
cd chrome-extension

# 安装依赖
npm install

# 构建
npm run build
```

然后在Chrome中:
1. 打开 `chrome://extensions/`
2. 启用"开发者模式"
3. 点击"加载已解压的扩展程序"
4. 选择 `chrome-extension/dist` 目录

## 功能特性

### 后端
- 用户注册/登录 (邮箱密码)
- OAuth登录 (Google/GitHub)
- 多租户支持
- 保险库管理
- 密码凭证CRUD
- JWT认证
- AES-256加密存储

### 前端Web
- 响应式设计
- 保险库管理
- 密码生成器
- 密码强度检测
- 端到端加密

### Chrome扩展
- 快速登录
- 密码列表
- 自动检测登录表单
- 一键填充
- 密码生成器

## API 端点

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录 |
| GET | /api/auth/oauth/:provider | OAuth登录 |
| GET | /api/tenants | 获取租户列表 |
| POST | /api/vaults | 创建保险库 |
| GET | /api/vaults | 获取保险库列表 |
| POST | /api/vaults/:id/credentials | 创建凭证 |
| GET | /api/vaults/:id/credentials | 获取凭证列表 |
| GET | /api/credentials/search | 搜索凭证 |

## 安全说明

1. **密码加密**: 所有密码使用AES-256-GCM加密后存储
2. **密钥派生**: 使用PBKDF2从用户密码派生主密钥
3. **主密钥**: 主密钥仅存储在客户端内存中，不会传输到服务器
4. **传输安全**: 生产环境应使用HTTPS

## 配置OAuth

在 `backend/config/config.toml` 中配置OAuth:

```toml
[oauth.google]
clientId = "your-google-client-id"
clientSecret = "your-google-client-secret"
redirectUrl = "http://localhost:8080/api/auth/oauth/google/callback"

[oauth.github]
clientId = "your-github-client-id"
clientSecret = "your-github-client-secret"
redirectUrl = "http://localhost:8080/api/auth/oauth/github/callback"
```

## 许可证

MIT


使用方案2

好的，我来帮你配置 CRXJS Vite Plugin。