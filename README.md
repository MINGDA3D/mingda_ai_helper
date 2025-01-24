# MINGDA AI助手

MINGDA AI助手是一个专为3D打印机设计的AI推理服务助手，通过与Klipper固件配合使用，能够实时监控和预测3D打印过程中可能出现的错误。

## 功能特性

- 🤖 实时AI预测：监控打印过程中的潜在问题
- 🔄 灵活部署：支持本地和云端AI服务混合调用
- 🛑 智能暂停：根据预测结果自动暂停打印
- 🔐 安全认证：采用AES-256加密保护数据传输
- 💾 数据存储：使用SQLite3数据库保存配置和预测结果
- 🔌 Moonraker集成：无缝对接Klipper生态系统

## 系统要求

- Go 1.16或更高版本
- SQLite3
- 支持Moonraker API的Klipper固件

## 安装说明

1. 克隆仓库：
```bash
git clone [repository_url]
cd mingda_ai_helper
```

2. 安装依赖：
```bash
go mod download
```

3. 配置文件：
- 复制`config/config.yaml.example`到`config/config.yaml`
- 根据实际环境修改配置文件

4. 编译运行：
```bash
go build -o mingda_ai_helper cmd/main.go
./mingda_ai_helper
```

## API接口

### 1. 健康检查
```
GET /api/v1/ai/health
```

### 2. 设备注册
```
POST /api/v1/machine/register
Content-Type: application/json

{
  "machine_model": "string",
  "machine_sn": "string"
}
```

### 3. Token刷新
```
POST /api/v1/token/refresh
Content-Type: application/json

{
  "machine_sn": "string",
  "old_token": "string"
}
```

### 4. 用户设置同步
```
POST /api/v1/settings/sync
Content-Type: application/json

{
  "enable_ai": true,
  "enable_cloud_ai": true,
  "confidence_threshold": 80,
  "pause_on_threshold": true
}
```

### 5. 预测请求
```
POST /api/v1/predict
Content-Type: application/json

{
  "image_url": "http://example.com/images/xxx.jpg",
  "task_id": "PT202403120001",
  "callback_url": "http://cloud-service/api/v1/ai/callback"
}
```

## 目录结构

```
mingda_ai_help/
├── cmd/
│   └── main.go          # 应用程序入口点
├── config/
│   ├── config.yaml      # 配置文件
│   └── config.go        # 配置加载逻辑
├── handlers/
│   ├── api.go          # API路由定义
│   └── handler.go      # API处理函数
├── models/
│   ├── machine.go      # 机型信息模型
│   ├── setting.go      # 用户设置模型
│   └── prediction.go   # 预测结果模型
├── services/
│   ├── ai_service.go   # AI服务调用逻辑
│   ├── db_service.go   # 数据库操作逻辑
│   └── log_service.go  # 日志记录逻辑
└── utils/
    ├── jwt.go         # JWT相关工具函数
    └── utils.go       # 其他辅助函数
```

## 安全说明

- 所有API请求都需要携带认证Token
- 使用AES-256加密算法保护数据传输
- 完整的日志审计系统

