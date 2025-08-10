#!/usr/bin/env python3
"""
Create a simple initial model for federated learning demonstration.
This mimics what would typically be a PyTorch or TensorFlow model.
"""
import struct
import os
import argparse

def create_initial_model(model_path: str, num_params: int = 10):
    """Create a simple binary model file with float32 weights."""
    os.makedirs(os.path.dirname(model_path), exist_ok=True)
    
    # Simple initialization: small random-like values scaled by num_params
    weights = [0.1 * (i + 1) for i in range(num_params)]
    
    with open(model_path, 'wb') as f:
        for weight in weights:
            f.write(struct.pack('<f', weight))  # little-endian float32
    
    print(f"Created initial model: {model_path}")
    print(f"Model size: {len(weights)} parameters ({len(weights) * 4} bytes)")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Create a simple initial model for federated learning")
    parser.add_argument("--output", "-o", default="models/init_model.pt", 
                        help="Output path for the model file")
    parser.add_argument("--size", "-s", type=int, default=10,
                        help="Number of parameters in the model")
    
    args = parser.parse_args()
    create_initial_model(args.output, args.size)