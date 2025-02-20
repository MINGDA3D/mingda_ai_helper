#!/bin/bash

# 确保脚本以root权限运行
if [ "$EUID" -ne 0 ]; then 
    echo "请使用root权限运行此脚本"
    exit 1
fi

# 设置变量
APP_NAME="mingda_ai_helper"
INSTALL_DIR="/home/mingda/mingda_ai_helper"
SERVICE_NAME="${APP_NAME}.service"
CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "开始部署 ${APP_NAME}..."

# 停止现有服务（如果存在）
if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo "停止现有服务..."
    systemctl stop ${SERVICE_NAME}
fi

# 编译程序
echo "编译程序..."
cd ${INSTALL_DIR}
# 查找main.go文件
MAIN_GO=$(find . -name "main.go" -type f)
if [ -z "$MAIN_GO" ]; then
    echo "错误：未找到main.go文件"
    exit 1
fi

MAIN_DIR=$(dirname "$MAIN_GO")
echo "找到main.go文件：$MAIN_GO"
echo "切换到目录：$MAIN_DIR"
cd "$MAIN_DIR"

# 设置GOPATH和其他必要的环境变量
export GOPATH="/home/mingda/go"
export PATH=$PATH:/usr/local/go/bin

# 获取依赖
echo "获取依赖..."
go mod tidy

# 编译
echo "开始编译..."
go build -o ${INSTALL_DIR}/${APP_NAME}

if [ ! -f "${INSTALL_DIR}/${APP_NAME}" ]; then
    echo "错误：编译失败，可执行文件未生成"
    exit 1
fi

# 设置权限
echo "设置权限..."
chmod +x ${INSTALL_DIR}/${APP_NAME}
chown -R mingda:mingda ${INSTALL_DIR}

# 复制并安装systemd服务文件
echo "安装systemd服务..."
cp ${CURRENT_DIR}/${SERVICE_NAME} /etc/systemd/system/
chmod 644 /etc/systemd/system/${SERVICE_NAME}

# 重新加载systemd配置
echo "重新加载systemd配置..."
systemctl daemon-reload

# 启用并启动服务
echo "启用并启动服务..."
systemctl enable ${SERVICE_NAME}
systemctl start ${SERVICE_NAME}

# 检查服务状态
echo "检查服务状态..."
systemctl status ${SERVICE_NAME}

echo "部署完成！"
echo "可以使用以下命令管理服务："
echo "  启动服务: systemctl start ${SERVICE_NAME}"
echo "  停止服务: systemctl stop ${SERVICE_NAME}"
echo "  重启服务: systemctl restart ${SERVICE_NAME}"
echo "  查看状态: systemctl status ${SERVICE_NAME}"
echo "  查看日志: journalctl -u ${SERVICE_NAME} -f"