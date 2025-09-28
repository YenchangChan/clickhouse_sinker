# ClickHouse Sinker Web 管理界面

这是一个基于Vue 3和Element Plus构建的ClickHouse Sinker管理界面。

## 功能特性

### 1. 概览页面
- 显示系统基本信息：版本、Goroutines数量、内存使用等
- 实时更新系统状态
- 显示任务总数和服务状态

### 2. 任务管理
- 以表格形式展示所有任务
- 支持按任务名搜索和排序
- 显示任务详细信息：名称、Topic、消费者组、表名、任务类型等
- 提供查看详情和任务指标功能

### 3. 实时日志
- 显示最新500条日志
- 支持日志过滤和搜索
- 可暂停/继续日志刷新
- 日志按级别着色显示（ERROR、WARN、INFO、DEBUG）

### 4. 系统诊断
- 获取Heap内存信息
- 获取Goroutine信息
- 支持pprof相关诊断功能

### 5. 配置管理
- 查看命令行参数
- 查看配置文件内容
- JSON格式化显示

## API接口

界面调用以下后端API：

- `GET /api/v1/metrics/procinfo` - 获取系统进程信息
- `GET /api/v1/tasks` - 获取所有任务列表
- `GET /api/v1/tasks/:taskname` - 获取指定任务详情
- `GET /api/v1/config` - 获取配置信息
- `GET /api/v1/cmdline` - 获取命令行参数
- `GET /debug/pprof/heap` - 获取Heap信息
- `GET /debug/pprof/goroutine?debug=1` - 获取Goroutine信息

## 技术栈

- **前端框架**: Vue 3 (CDN)
- **HTTP客户端**: Fetch API
- **样式**: CSS3 + Flexbox + Grid (内联)
- **架构**: 单文件应用 (SPA)

## 文件结构

```
mvc/static/
├── index.html      # 主页面（包含所有HTML、CSS、JavaScript）
└── README.md       # 说明文档
```

## 使用方法

### 方式一：集成模式（推荐）
1. 使用 `make ui` 编译前端资源到二进制文件
2. 启动ClickHouse Sinker服务
3. 访问 `http://localhost:端口/` 

### 方式二：开发模式
1. 启动ClickHouse Sinker服务
2. 访问 `http://localhost:端口/static/` 或 `http://localhost:端口/`
3. 通过导航栏切换不同功能页面

### 方式三：单独启动前端项目

#### 使用Python简单服务器
```bash
cd mvc/static
python -m http.server 8080
# 访问 http://localhost:8080
```

#### 使用Node.js serve
```bash
# 安装serve
npm install -g serve

# 启动服务
cd mvc/static
serve -s . -p 8080
# 访问 http://localhost:8080
```

#### 代理到其他后端服务

如果需要代理到其他后端服务，可以修改 `app.js` 中的API基础路径：

```javascript
// 在app.js中修改apiCall方法
async apiCall(url, options = {}) {
    try {
        // 代理到其他服务器
        const baseURL = 'http://your-backend-server:port';
        const response = await axios.get(`${baseURL}/api/v1${url}`, options);
        // ... 其他代码
    } catch (error) {
        // ... 错误处理
    }
}
```

或者使用nginx反向代理：

```nginx
server {
    listen 80;
    server_name localhost;
    
    # 前端静态文件
    location / {
        root /path/to/mvc/static;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
    
    # API代理到后端
    location /api/ {
        proxy_pass http://backend-server:port;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # pprof代理
    location /debug/ {
        proxy_pass http://backend-server:port;
    }
}

## 响应式设计

界面支持移动端访问，在小屏幕设备上会自动调整布局。

## 待实现功能

- [ ] SSE实时日志流
- [ ] 任务指标详细展示
- [ ] 更多系统监控指标
- [ ] 配置文件编辑功能
- [ ] 任务管理操作（启动/停止/重启）