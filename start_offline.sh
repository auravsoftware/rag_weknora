#!/usr/bin/env bash
# ==============================================================================
# start_offline.sh — Khởi chạy Hệ thống RAG Tiếng Việt 100% Offline (24GB VRAM)
# Bung Docker trọn gói 1 lệnh duy nhất trên máy chủ
# ==============================================================================

set -euo pipefail

cd "$(dirname "$0")"

echo "================================================================================"
echo "    BUNG DOCKER HỆ THỐNG RAG TIẾNG VIỆT 100% OFFLINE (1x 24GB VRAM GPU)        "
echo "================================================================================"

# 1. Tự động khởi tạo tệp cấu hình .env nếu chưa có
if [ ! -f .env ]; then
    echo "[+] Chưa tìm thấy tệp .env, đang khởi tạo tự động từ .env.offline-24gb.example..."
    cp .env.offline-24gb.example .env
else
    echo "[✓] Tệp cấu hình .env đã sẵn sàng."
fi

# 2. Tạo các thư mục lưu trữ dữ liệu và cache model
mkdir -p ./models_cache/huggingface ./data/files
echo "[✓] Đã đảm bảo các thư mục dữ liệu ./models_cache/huggingface và ./data/files tồn tại."

# 3. Kiểm tra công cụ Docker Compose
COMPOSE_CMD=""
if docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
    COMPOSE_CMD="docker-compose"
else
    echo "[!] Lỗi: Không tìm thấy lệnh 'docker compose' hoặc 'docker-compose'. Vui lòng cài đặt Docker Compose!" >&2
    exit 1
fi
echo "[✓] Sử dụng lệnh Compose: $COMPOSE_CMD"

# 4. Kiểm tra GPU NVIDIA (nếu có nvidia-smi)
if command -v nvidia-smi >/dev/null 2>&1; then
    echo "[✓] Đã phát hiện GPU NVIDIA trên máy chủ:"
    nvidia-smi --query-gpu=name,memory.total --format=csv,noheader || true
else
    echo "[!] Cảnh báo: Lệnh nvidia-smi chưa sẵn sàng. Hãy đảm bảo máy chủ đã cài đặt NVIDIA Driver và NVIDIA Container Toolkit!"
fi

# 5. Khởi chạy toàn bộ hệ thống bằng Docker Compose với dải cổng 6666 - 6672
echo "[+] Đang khởi chạy các container dịch vụ..."
$COMPOSE_CMD -f docker-compose.offline-24gb.yml up -d

echo "================================================================================"
echo "    KHỞI CHẠY HỆ THỐNG THÀNH CÔNG!                                              "
echo "================================================================================"
echo "  • Giao diện Web RAG Tiếng Việt : http://localhost:6666 (hoặc http://<IP>:6666)"
echo "  • Backend API                  : http://localhost:6667 (Cổng container 8080)   "
echo "  • PostgreSQL (ParadeDB HNSW)   : Cổng 6668 (Postgres 5432)                     "
echo "  • Redis Cache / Task Queue     : Cổng 6669 (Redis 6379)                        "
echo "  • vLLM (Qwen2.5-VL-7B OCR+LLM) : http://localhost:6670/v1                      "
echo "  • Infinity (BGE-M3 + Reranker) : http://localhost:6671/v1                      "
echo "  • DocReader Parser             : Cổng 6672 (gRPC 50051)                        "
echo "================================================================================"
echo "  Để theo dõi nhật ký hoạt động (logs):"
echo "    $COMPOSE_CMD -f docker-compose.offline-24gb.yml logs -f"
echo "================================================================================"
