#!/usr/bin/env python3
import argparse, torch, torch.nn as nn, torch.optim as optim
from torchvision import datasets, transforms

class Net(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc = nn.Linear(28*28, 10)
    def forward(self, x):
        return self.fc(x.view(-1, 28*28))

if __name__=="__main__":
    p = argparse.ArgumentParser()
    p.add_argument("--model-in", required=True)
    p.add_argument("--model-out", required=True)
    p.add_argument("--epochs", type=int, default=1)
    p.add_argument("--batch_size", type=int, default=32)
    args = p.parse_args()

    ds = datasets.MNIST("./data", train=True, download=True, transform=transforms.ToTensor())
    loader = torch.utils.data.DataLoader(ds, batch_size=args.batch_size, shuffle=True)
    m = Net()
    m.load_state_dict(torch.load(args.model_in))
    opt = optim.SGD(m.parameters(), lr=0.01)
    crit = nn.CrossEntropyLoss()

    for _ in range(args.epochs):
        for x,y in loader:
            opt.zero_grad()
            loss = crit(m(x), y)
            loss.backward()
            opt.step()
    torch.save(m.state_dict(), args.model_out)
