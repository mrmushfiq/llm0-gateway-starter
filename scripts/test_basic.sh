#!/bin/bash

# Simple test script for the LLM Gateway

API_KEY="gw_test_abc123"
BASE_URL="http://localhost:8080"

echo "🧪 Testing LLM Gateway Starter"
echo "================================"
echo ""

# Test 1: Health Check
echo "1️⃣  Testing health endpoint..."
curl -s "${BASE_URL}/health"
echo ""
echo ""

# Test 2: Basic Chat Completion (GPT-4o-mini)
echo "2️⃣  Testing basic chat completion (gpt-4o-mini)..."
curl -s -X POST "${BASE_URL}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Say hello in one word"}
    ]
  }' | jq '.'
echo ""

# Test 3: Streaming Chat Completion
echo "3️⃣  Testing streaming chat completion..."
curl -N -X POST "${BASE_URL}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Count to 3"}
    ],
    "stream": true
  }'
echo ""
echo ""

# Test 4: Cache Test (duplicate request)
echo "4️⃣  Testing cache (duplicate request)..."
echo "First request (should be a cache miss):"
curl -s -X POST "${BASE_URL}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "What is 2+2?"}
    ]
  }' | jq '.cost_usd, .latency_ms' -r | head -1
echo ""

echo "Second request (should be a cache hit):"
curl -s -X POST "${BASE_URL}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "What is 2+2?"}
    ]
  }' | jq '.cost_usd, .latency_ms' -r | head -1
echo ""

# Test 5: Check Response Headers
echo "5️⃣  Testing response headers..."
curl -I -X POST "${BASE_URL}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Hi"}
    ]
  }' 2>&1 | grep "X-"
echo ""

echo "✅ Tests complete!"
