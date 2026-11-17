#!/usr/bin/env bash
# ==============================================================================
# stop_offline.sh — Dừng toàn bộ hệ thống RAG và giải phóng các cổng 6666-6672
# ==============================================================================

set -e
cd "$(dirname "$0")"

echo "================================================================================"
echo "    ĐANG DỪNG HỆ THỐNG VÀ GIẢI PHÓNG TOÀN BỘ CỔNG (6666 - 6672)...            "
echo "================================================================================"

# 1. Dừng cụm Docker Compose
docker compose -f docker-compose.offline-24gb.yml down --remove-orphans || true

# 2. Cưỡng chế dừng sạch mọi container nếu còn sót ngầm
docker stop WeKnora-frontend WeKnora-app WeKnora-postgres WeKnora-redis WeKnora-vllm-qwen25-vl WeKnora-infinity-models WeKnora-docreader 2>/dev/null || true
docker rm -f WeKnora-frontend WeKnora-app WeKnora-postgres WeKnora-redis WeKnora-vllm-qwen25-vl WeKnora-infinity-models WeKnora-docreader 2>/dev/null || true

# 3. Quét giải phóng cổng 7666-7672 và 6666-6672 nếu có tiến trình chiếm giữ
for PORT in 7666 7667 7668 7669 7670 7671 7672 6666 6667 6668 6669 6670 6671 6672; do
  fuser -k "${PORT}/tcp" 2>/dev/null || true
done

echo "================================================================================"
echo "    [✓] ĐÃ DỪNG TOÀN BỘ DỊCH VỤ VÀ ĐÓNG HẾT CÁC CỔNG THÀNH CÔNG!               "
echo "================================================================================"
