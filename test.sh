#!/bin/bash

BASE_URL="${BASE_URL:-http://localhost:8080}"
API="$BASE_URL/api/v1/tasks"
PASSED=0
FAILED=0

run_test() {
    if [ "$1" -eq 0 ]; then
        echo "PASS: $2"
        PASSED=$((PASSED + 1))
    else
        echo "FAIL: $2"
        FAILED=$((FAILED + 1))
    fi
}

get_id() {
    echo "$1" | grep -oE '"id":[0-9]+' | head -1 | grep -oE '[0-9]+'
}

docker compose exec postgres psql -U postgres -d taskservice -c "DELETE FROM tasks WHERE title LIKE 'TEST_%';" > /dev/null 2>&1 || true

# Тест Создание периодической задачи:
RESP=$(curl -s -X POST "$API" -H "Content-Type: application/json" -d '{"title":"TEST_Daily","due_date":"2026-05-01T10:00:00Z","is_recurrence":true,"recurrence_type":"daily","recurrence_config":{"interval_days":1}}')
echo "$RESP" | grep -qE '"title":"TEST_Daily"'
run_test $? "Title matches"
echo "$RESP" | grep -qE '"is_recurrence":true'
run_test $? "Recurrence flag set"

# Тест Материализация экземпляров:
sleep 1
RESP=$(curl -s "$API?from=2026-05-01T00:00:00Z&to=2026-05-05T23:59:59Z&limit=5")
echo "$RESP" | grep -qE '"is_recurrence":false'
run_test $? "Instances materialized"
echo "$RESP" | grep -qE '"recurrence_parent_id":[0-9]+'
run_test $? "Parent ID present in instances"

# Тест Курсорная пагинация:
RESP=$(curl -s "$API?from=2026-05-01T00:00:00Z&to=2026-05-10T23:59:59Z&limit=2")
CURSOR=$(echo "$RESP" | grep -oE '"cursor":"[^"]*"' | head -1 | sed 's/"cursor":"//;s/"//')
if [ -n "$CURSOR" ]; then
    echo "PASS: Cursor returned ($CURSOR)"
    PASSED=$((PASSED + 1))
else
    echo "FAIL: Cursor missing"
    FAILED=$((FAILED + 1))
fi
RESP2=$(curl -s "$API?from=2026-05-01T00:00:00Z&to=2026-05-10T23:59:59Z&limit=2&cursor=$CURSOR")
echo "$RESP2" | grep -qE '"tasks":\['
run_test $? "Next page fetches tasks"

# Тест Обновление задачи:
RESP_INSTANCES=$(curl -s "$API?from=2026-05-01T00:00:00Z&to=2026-05-05T23:59:59Z&limit=1")
INSTANCE_ID=$(get_id "$RESP_INSTANCES")
if [ -n "$INSTANCE_ID" ]; then
    RESP=$(curl -s -X PUT "$API/$INSTANCE_ID" -H "Content-Type: application/json" -d '{"title":"TEST_Updated","status":"in_progress","due_date":"2026-05-01T10:00:00Z","is_recurrence":false,"recurrence_type":null,"recurrence_config":null}')
    echo "$RESP" | grep -qE '"status":"in_progress"'
    run_test $? "Status updated via PUT"
else
    run_test 1 "PUT skipped (no instance ID)"
fi

# Тест Удаление задачи:
CREATE_RESP=$(curl -s -X POST "$API" -H "Content-Type: application/json" -d '{"title":"TEST_Temp","due_date":"2026-06-01T10:00:00Z"}')
TEMP_ID=$(get_id "$CREATE_RESP")
if [ -n "$TEMP_ID" ]; then
    curl -s -X DELETE "$API/$TEMP_ID" > /dev/null
    RESP=$(curl -s "$API/$TEMP_ID")
    echo "$RESP" | grep -qE '"error":"task not found"'
    run_test $? "Task deleted and returns 404"
else
    run_test 1 "Delete skipped (creation failed: $CREATE_RESP)"
fi

echo ""
echo "Results: $PASSED passed, $FAILED failed"
if [ $FAILED -gt 0 ]; then exit 1; fi
exit 0