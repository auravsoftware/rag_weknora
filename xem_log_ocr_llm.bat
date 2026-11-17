@echo off
chcp 65001 >nul
echo ================================================================================
echo    THEO DÕI TIẾN TRÌNH & LOG CỦA MÔ HÌNH VLLM (QWEN2.5-VL OCR + LLM)
echo ================================================================================
echo Đang lấy log thời gian thực từ container WeKnora-vllm-qwen25-vl...
echo (Nhấn Ctrl + C để thoát khỏi màn hình log)
echo ================================================================================
docker logs -f WeKnora-vllm-qwen25-vl
pause
