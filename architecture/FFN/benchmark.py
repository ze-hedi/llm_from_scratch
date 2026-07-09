import time 
import torch 
import matplotlib.pyplot as plt 
from architecture.FFN.feed_forward import (
    FeedForward ,
    FeedForward_torch
)

d_in = 768
d_ff = 4*d_in 
d_out = 768 

context_length = 1024 
dropout = 0.0
batch_size = 5


cfg = {
    "emb_dim" : d_in, 
}

device = torch.device("cuda" if torch.cuda.is_available() else "cpu") 
models = {
    "FeedForward" : FeedForward(cfg,d_ff) , 
    "FeedForward (compiled)" : FeedForward(cfg,d_ff,True) , 
    "FeedForward_torch" : FeedForward_torch(cfg,d_ff),
    "FeedForward_torch (compiled)" : FeedForward_torch(cfg,d_ff,True)
}

dtype = torch.float16 if device.type == "cuda" else torch.float32
x = torch.randn(batch_size, context_length, d_in, device=device, dtype=dtype)

num_runs = 30
results = {}

for name, model in models.items():
    model = model.to(device=device, dtype=dtype).eval()

    if device.type == "cuda":
        with torch.no_grad():
            model(x)
        torch.cuda.synchronize()

    times = []
    for _ in range(num_runs):
        if device.type == "cuda":
            torch.cuda.synchronize()

        start = time.perf_counter()
        with torch.no_grad():
            model(x)

        if device.type == "cuda":
            torch.cuda.synchronize()

        times.append((time.perf_counter() - start) * 1000)

    avg_ms = sum(times) / num_runs
    results[name] = avg_ms
    print(f"{name}: {avg_ms:.2f} ms (avg over {num_runs} runs)")

## Forward plot
names = list(results.keys())
fwd_times = list(results.values())

colors = [
    "#4c72b0", "#2e4a7a",
    "#55a868", "#367a45",
]

plt.figure(figsize=(14, 6))
bars = plt.bar(names, fwd_times, color=colors)
for bar, t in zip(bars, fwd_times):
    plt.text(bar.get_x() + bar.get_width() / 2, bar.get_height(),
             f"{t:.2f} ms", ha="center", va="bottom", fontsize=9)
plt.ylabel("Time (ms)")
plt.title(f"Forward Pass (avg over {num_runs} runs)")
plt.xticks(rotation=30, ha="right")
plt.tight_layout()
plt.savefig("benchmark_forward.png", dpi=150)
plt.show()

## Forward + Backward benchmark
print("\n--- Forward + Backward ---")
results_fwd_bwd = {}

for name, model in models.items():
    model = model.to(device=device, dtype=dtype).train()

    # warmup
    if device.type == "cuda":
        out = model(x)
        out.sum().backward()
        torch.cuda.synchronize()

    times = []
    for _ in range(num_runs):
        if device.type == "cuda":
            torch.cuda.synchronize()

        start = time.perf_counter()
        out = model(x)
        out.sum().backward()

        if device.type == "cuda":
            torch.cuda.synchronize()

        times.append((time.perf_counter() - start) * 1000)

    avg_ms = sum(times) / num_runs
    results_fwd_bwd[name] = avg_ms
    print(f"{name}: {avg_ms:.2f} ms (avg over {num_runs} runs)")

## Forward + Backward plot
fwd_bwd_times = list(results_fwd_bwd.values())

plt.figure(figsize=(14, 6))
bars = plt.bar(names, fwd_bwd_times, color=colors)
for bar, t in zip(bars, fwd_bwd_times):
    plt.text(bar.get_x() + bar.get_width() / 2, bar.get_height(),
             f"{t:.2f} ms", ha="center", va="bottom", fontsize=9)
plt.ylabel("Time (ms)")
plt.title(f"Forward + Backward Pass (avg over {num_runs} runs)")
plt.xticks(rotation=30, ha="right")
plt.tight_layout()
plt.savefig("benchmark_forward_backward.png", dpi=150)
plt.show()

