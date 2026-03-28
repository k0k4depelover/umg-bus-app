#!/bin/bash

BASE="http://localhost:8084"
PASS=0
FAIL=0

check() {
    local name=$1
    local expected=$2
    local actual=$3

    if [ "$actual" = "$expected" ]; then
        echo "  PASS  $name"
        ((PASS++))
    else
        echo "  FAIL  $name (esperado: $expected, recibido: $actual)"
        ((FAIL++))
    fi
}

echo "── Smoke tests UMG Bus Tracker ──"

# Health
STATUS=$(curl -s -o /dev/null -w "%{http_code}" $BASE/health-check)
check "GET /health-check" "200" "$STATUS"

# Campus list
STATUS=$(curl -s -o /dev/null -w "%{http_code}" $BASE/campus)
check "GET /campus" "200" "$STATUS"

# Login con credenciales inválidas debe dar 401
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"noexiste","password":"mal","role":"pilot"}')
check "POST /auth/login (inválido → 401)" "401" "$STATUS"

# GraphQL sin token debe dar 401
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ campuses { name } }"}')
check "POST /graphql (sin token → 401)" "401" "$STATUS"

# MCP campus list
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE/mcp \
  -H "Content-Type: application/json" \
  -d '{"tool":"get_campus_list","input":{}}')
check "POST /mcp get_campus_list" "200" "$STATUS"

# Tool desconocido en MCP debe dar 400
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE/mcp \
  -H "Content-Type: application/json" \
  -d '{"tool":"no_existe","input":{}}')
check "POST /mcp (tool inválido → 400)" "400" "$STATUS"

echo ""
echo "Resultado: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
