#!/bin/bash

# =============================================================================
# Скрипт нагрузочного тестирования для Notes API
# =============================================================================

set -e

# Конфигурация
BASE_URL="${BASE_URL:-http://localhost:8080}"
RESULTS_DIR="./results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Создаём директорию для результатов
mkdir -p "$RESULTS_DIR"

echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}  Нагрузочное тестирование Notes API${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
echo "Base URL: $BASE_URL"
echo "Результаты будут сохранены в: $RESULTS_DIR"
echo ""

# Функция проверки доступности сервиса
check_service() {
    echo -e "${YELLOW}Проверка доступности сервиса...${NC}"
    if curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/notes" | grep -q "200\|401\|403"; then
        echo -e "${GREEN}Сервис доступен${NC}"
        return 0
    else
        echo -e "${RED}Сервис недоступен по адресу $BASE_URL${NC}"
        return 1
    fi
}

# Функция получения количества заметок в БД
get_notes_count() {
    # Это примерная команда, нужно адаптировать под ваше окружение
    docker exec goose_postgres psql -U postgres -d notes_db -t -c "SELECT COUNT(*) FROM note;" 2>/dev/null | tr -d ' ' || echo "N/A"
}

# Функция очистки БД
reset_db() {
    echo -e "${YELLOW}Очистка базы данных...${NC}"
    docker exec goose_postgres psql -U postgres -d notes_db -c "
        TRUNCATE note CASCADE;
        ALTER SEQUENCE note_id_seq RESTART WITH 1;
    " 2>/dev/null || echo "Не удалось очистить БД (возможно, нужно адаптировать команду)"
    echo -e "${GREEN}База данных очищена${NC}"
}

# =============================================================================
# ТЕСТ 1: Создание заметок (POST /api/notes)
# =============================================================================
run_create_test() {
    local threads="${1:-4}"
    local connections="${2:-100}"
    local duration="${3:-300s}"  # 5 минут для создания 100k заметок
    local target_count="${4:-100000}"

    echo ""
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}  ТЕСТ: Создание заметок${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo "Параметры:"
    echo "  - Потоки: $threads"
    echo "  - Соединения: $connections"
    echo "  - Длительность: $duration"
    echo "  - Цель: $target_count заметок"
    echo ""

    local result_file="$RESULTS_DIR/create_${TIMESTAMP}.txt"

    echo "Запуск теста создания..."
    wrk -t"$threads" -c"$connections" -d"$duration" \
        -s scripts/create_notes.lua \
        "$BASE_URL" 2>&1 | tee "$result_file"

    echo ""
    echo "Количество заметок в БД после теста: $(get_notes_count)"
    echo "Результаты сохранены в: $result_file"
}

# =============================================================================
# ТЕСТ 2: Чтение заметок (GET /api/notes/{id})
# =============================================================================
run_read_test() {
    local threads="${1:-4}"
    local connections="${2:-100}"
    local duration="${3:-60s}"

    echo ""
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}  ТЕСТ: Чтение заметок${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo "Параметры:"
    echo "  - Потоки: $threads"
    echo "  - Соединения: $connections"
    echo "  - Длительность: $duration"
    echo ""

    local notes_count=$(get_notes_count)
    echo "Количество заметок в БД: $notes_count"

    if [ "$notes_count" = "0" ] || [ "$notes_count" = "N/A" ]; then
        echo -e "${RED}В базе нет заметок! Сначала запустите тест создания.${NC}"
        return 1
    fi

    local result_file="$RESULTS_DIR/read_${TIMESTAMP}.txt"

    echo "Запуск теста чтения..."
    wrk -t"$threads" -c"$connections" -d"$duration" \
        -s scripts/read_notes.lua \
        "$BASE_URL" 2>&1 | tee "$result_file"

    echo ""
    echo "Результаты сохранены в: $result_file"
}

# =============================================================================
# ПОЛНЫЙ ЦИКЛ ТЕСТИРОВАНИЯ
# =============================================================================
run_full_test() {
    echo -e "${GREEN}Запуск полного цикла тестирования${NC}"

    # Проверка сервиса
    check_service || exit 1

    # Опционально: сброс БД
    read -p "Очистить базу данных перед тестом? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        reset_db
    fi

    # Тест создания
    run_create_test 4 100 300s 100000

    echo ""
    echo -e "${YELLOW}Пауза 10 секунд перед тестом чтения...${NC}"
    sleep 10

    # Тест чтения
    run_read_test 4 100 60s

    echo ""
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}  Тестирование завершено!${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo "Результаты в директории: $RESULTS_DIR"
}

# =============================================================================
# СПРАВКА
# =============================================================================
show_help() {
    echo "Использование: $0 [команда] [опции]"
    echo ""
    echo "Команды:"
    echo "  create [threads] [connections] [duration]  - Тест создания заметок"
    echo "  read [threads] [connections] [duration]    - Тест чтения заметок"
    echo "  full                                        - Полный цикл тестирования"
    echo "  reset                                       - Очистить базу данных"
    echo "  count                                       - Показать количество заметок"
    echo "  help                                        - Показать эту справку"
    echo ""
    echo "Примеры:"
    echo "  $0 create 4 100 300s     # Создание: 4 потока, 100 соединений, 5 минут"
    echo "  $0 read 8 200 60s        # Чтение: 8 потоков, 200 соединений, 1 минута"
    echo "  $0 full                  # Полный цикл тестов"
    echo ""
    echo "Переменные окружения:"
    echo "  BASE_URL  - URL сервиса (по умолчанию: http://localhost:8080)"
}

# =============================================================================
# MAIN
# =============================================================================
case "${1:-help}" in
    create)
        check_service || exit 1
        run_create_test "${2:-4}" "${3:-100}" "${4:-300s}"
        ;;
    read)
        check_service || exit 1
        run_read_test "${2:-4}" "${3:-100}" "${4:-60s}"
        ;;
    full)
        run_full_test
        ;;
    reset)
        reset_db
        ;;
    count)
        echo "Количество заметок: $(get_notes_count)"
        ;;
    help|*)
        show_help
        ;;
esac