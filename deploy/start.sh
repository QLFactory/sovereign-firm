#!/bin/bash
echo "🚀 Building and deploying Sovereign Firm (Production Grade)..."

# Load .env variables if present
if [ -f .env ]; then
  export $(cat .env | xargs)
fi

# Ensure Ollama vars are present
if [ -z "$OLLAMA_HOST" ]; then
    echo "⚠️  OLLAMA_HOST not set. Defaulting to http://host.docker.internal:11434"
    echo "    (If using remote Ollama, export OLLAMA_HOST=http://IP:PORT before running this)"
fi

docker-compose -f docker-compose.prod.yaml up --build -d

echo "✅ Deployment Complete."
echo "   - Console: http://localhost:3000"
echo "   - Temporal UI: http://localhost:8088"
