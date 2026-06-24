#!/bin/bash
# Run this in a SECOND terminal while the server is running in the first.
PORT="${1:-8080}"
BASE="http://127.0.0.1:$PORT"

echo "Testing server on port $PORT..."
echo

echo "1. PUT pet?v=cat"
curl -s -o /dev/null -w "   -> HTTP %{http_code}\n" -X PUT "$BASE/pet?v=cat"

echo "2. PUT pet?v=dog"
curl -s -o /dev/null -w "   -> HTTP %{http_code}\n" -X PUT "$BASE/pet?v=dog"

echo "3. GET /pet (expect: cat)"
curl -s -w "   -> HTTP %{http_code}\n" "$BASE/pet"

echo "4. GET /pet (expect: dog)"
curl -s -w "   -> HTTP %{http_code}\n" "$BASE/pet"

echo "5. GET /pet (expect: 404)"
curl -s -o /dev/null -w "   -> HTTP %{http_code}\n" "$BASE/pet"

echo
echo "Done. If you see 200/404 above, your server works."
