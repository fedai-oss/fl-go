#!/usr/bin/env python3
"""
Create a simple initial model for federated learning demonstration.
This mimics what would typically be a PyTorch or TensorFlow model.
"""
import struct
import os

def create_initial_model(model_path: str, num_params: int = 10):
    """Create a simple binary model file with float32 weights."""
    os.makedirs(os.path.dirname(model_path), exist_ok=True)
    
    # Simple initialization: small random-like values
    weights = [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0]
    
    with open(model_path, 'wb') as f:
        for weight in weights:
            f.write(struct.pack('<f', weight))  # little-endian float32
    
    print(f"Created initial model: {model_path}")
    print(f"Model size: {len(weights)} parameters ({len(weights) * 4} bytes)")

if __name__ == "__main__":
    create_initial_model("models/init_model.pt")