#!/usr/bin/env bash
# Demo LOST UPDATE — không dùng SELECT FOR UPDATE.
# Chạy 5 request đồng thời vào cùng 1 todo_id.
# Kết quả mong đợi: completed_count < 5 (lost update xảy ra).

TODO_ID=${1:?Usage: $0 <todo_id>}
ENDPOINT=${ENDPOINT:-http://localhost:8081}
N=5

echo "=== RACE DEMO (NO LOCK) ==="
echo "todo_id=$TODO_ID  endpoint=$ENDPOINT  concurrent=$N"
echo "Bắn $N request đồng thời... (mỗi request sleep 5s trong tx)"
echo ""

for i in $(seq 1 $N); do
  curl -s --max-time 30 -X POST "$ENDPOINT/debug/race/$TODO_ID" | \
    python3 -c "import sys,json; d=json.load(sys.stdin); print(f'[{d[\"pod\"]}] read={d[\"read_count\"]} write={d[\"written_count\"]}')" &
done
wait

echo ""
echo "=== Kiểm tra DB ==="
echo "Chạy lệnh sau trong postgres:"
echo "  SELECT id, completed_count FROM todos WHERE id = $TODO_ID;"
echo ""
echo "Expected: completed_count < $N  (lost update!)"
