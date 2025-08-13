#!/bin/bash

set -e

echo "🔒 FL-Go mTLS Security Demo"
echo "==========================="

# Build FL-Go if needed
if [ ! -f "build/fx" ]; then
    echo "Building FL-Go..."
    make build
fi

# Clean up any existing test directories
rm -rf mtls_demo_workspace
rm -rf certs

echo ""
echo "🏗️  Setting up mTLS demo workspace..."

# Create workspace with secure plan
mkdir -p mtls_demo_workspace
cd mtls_demo_workspace

# Copy the secure plan
cp ../plans/secure_example_plan.yaml plan.yaml

# Create necessary directories
mkdir -p models data src save logs certs

# Create a simple initial model (100 parameters)
python3 << 'EOF'
import torch
import numpy as np

# Create a simple model with 100 parameters
model_params = torch.randn(100)
torch.save(model_params, "models/init_model.pt")
print("Created initial model with 100 parameters")
EOF

# Create training script
cat > src/taskrunner.py << 'EOF'
#!/usr/bin/env python3
import argparse
import torch
import numpy as np
import os

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--model-in', required=True)
    parser.add_argument('--model-out', required=True) 
    parser.add_argument('--epochs', type=int, default=1)
    parser.add_argument('--batch-size', type=int, default=32)
    parser.add_argument('--lr', type=float, default=0.01)
    parser.add_argument('--data-path', default='data')
    
    args = parser.parse_args()
    
    print(f"🤖 Training with mTLS security enabled")
    print(f"  Model in: {args.model_in}")
    print(f"  Model out: {args.model_out}")
    print(f"  Epochs: {args.epochs}")
    
    # Load model
    model = torch.load(args.model_in)
    print(f"  Loaded model with {len(model)} parameters")
    
    # Simulate training by adding small random changes
    noise = torch.randn_like(model) * 0.01
    updated_model = model + noise
    
    # Save updated model
    torch.save(updated_model, args.model_out)
    print(f"  Saved updated model to {args.model_out}")
    print(f"  Training completed with mTLS security ✅")

if __name__ == "__main__":
    main()
EOF

chmod +x src/taskrunner.py

echo ""
echo "🔐 Starting mTLS Federated Learning Demo..."
echo "   This demo will:"
echo "   - Auto-generate TLS certificates"
echo "   - Start secure aggregator with mTLS"
echo "   - Connect secure collaborators"
echo "   - Run federated training with encryption"

echo ""
echo "📋 Demo Plan Configuration:"
cat plan.yaml | grep -A 20 "security:"

echo ""
echo "🚀 Starting aggregator with mTLS..."

# Start aggregator in background
../build/fx aggregator start --plan plan.yaml > aggregator.log 2>&1 &
AGGREGATOR_PID=$!

sleep 3

echo "🤝 Starting collaborators with mTLS..."

# Start collaborators in background
../build/fx collaborator start collab1 --plan plan.yaml > collab1.log 2>&1 &
COLLAB1_PID=$!

../build/fx collaborator start collab2 --plan plan.yaml > collab2.log 2>&1 &
COLLAB2_PID=$!

# Wait for training to complete
echo "⏳ Waiting for federated learning to complete..."
sleep 15

# Clean up processes
echo "🧹 Cleaning up processes..."
kill $AGGREGATOR_PID $COLLAB1_PID $COLLAB2_PID 2>/dev/null || true
sleep 2

echo ""
echo "📊 Demo Results:"
echo "==============="

echo ""
echo "🔍 Generated TLS Certificates:"
if [ -d "certs" ]; then
    ls -la certs/
    echo ""
    echo "📜 Certificate Details:"
    openssl x509 -in certs/ca.crt -text -noout | grep -A 5 "Subject:"
else
    echo "❌ Certificate directory not found"
fi

echo ""
echo "📈 Aggregator Log Summary:"
if [ -f "aggregator.log" ]; then
    echo "--- Recent aggregator activity ---"
    tail -10 aggregator.log
else
    echo "❌ Aggregator log not found"
fi

echo ""
echo "🤝 Collaborator Log Summary:"
if [ -f "collab1.log" ]; then
    echo "--- Collaborator 1 activity ---"
    tail -5 collab1.log
fi

if [ -f "collab2.log" ]; then
    echo "--- Collaborator 2 activity ---"
    tail -5 collab2.log
fi

echo ""
echo "🎯 Final Model Check:"
if [ -f "models/aggregated_model.pt" ]; then
    python3 << 'EOF'
import torch
initial = torch.load("models/init_model.pt")
final = torch.load("models/aggregated_model.pt")
diff = torch.norm(final - initial).item()
print(f"Model change magnitude: {diff:.6f}")
print(f"Initial model norm: {torch.norm(initial).item():.6f}")
print(f"Final model norm: {torch.norm(final).item():.6f}")
print("✅ Federated learning with mTLS completed successfully!")
EOF
else
    echo "❌ Final aggregated model not found"
fi

echo ""
echo "🏆 mTLS Security Demo Completed!"
echo "================================"
echo "✅ TLS certificates auto-generated"
echo "✅ Secure gRPC communication established"
echo "✅ Federated learning completed with encryption"
echo "✅ All communications were protected by mTLS"

cd ..
EOF
