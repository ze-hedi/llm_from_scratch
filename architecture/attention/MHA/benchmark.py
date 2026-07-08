import time
import torch
import matplotlib.pyplot as plt
from MHA import (
    MHA_wrapper,
    MultiHeadAttention,
    MHAFlashAttention,
    MHAEfficientAttention,
    MHAMathAttention,
)

d_in = 768
d_out = 768
context_length = 1024
dropout = 0.0
num_heads = 12
batch_size = 4
seq_len = 512

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

models = {
    "MHA_wrapper": MHA_wrapper(d_in, d_out, context_length, dropout, num_heads),
    "MultiHeadAttention": MultiHeadAttention(d_in, d_out, context_length, dropout, num_heads),
    "FlashAttention": MHAFlashAttention(d_in, d_out, context_length, dropout, num_heads),
    "EfficientAttention": MHAEfficientAttention(d_in, d_out, context_length, dropout, num_heads),
    "MathAttention": MHAMathAttention(d_in, d_out, context_length, dropout, num_heads),
}

dtype = torch.float16 if device.type == "cuda" else torch.float32
x = torch.randn(batch_size, seq_len, d_in, device=device, dtype=dtype)

num_runs = 15
results = {}

for name, model in models.items():
    model = model.to(device=device, dtype=dtype).eval()

    # warmup for CUDA
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
colors = ["#4c72b0", "#55a868", "#c44e52", "#8172b2", "#ccb974"]

plt.figure(figsize=(10, 6))
bars = plt.bar(names, fwd_times, color=colors)
for bar, t in zip(bars, fwd_times):
    plt.text(bar.get_x() + bar.get_width() / 2, bar.get_height(),
             f"{t:.2f} ms", ha="center", va="bottom", fontsize=10)
plt.ylabel("Time (ms)")
plt.title(f"Forward Pass (avg over {num_runs} runs)")
plt.xticks(rotation=15)
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
        x_warm = torch.randn_like(x, requires_grad=True)
        out = model(x_warm)
        out.sum().backward()
        torch.cuda.synchronize()

    times = []
    for _ in range(num_runs):
        x_run = torch.randn_like(x, requires_grad=True)

        if device.type == "cuda":
            torch.cuda.synchronize()

        start = time.perf_counter()
        out = model(x_run)
        out.sum().backward()

        if device.type == "cuda":
            torch.cuda.synchronize()

        times.append((time.perf_counter() - start) * 1000)

    avg_ms = sum(times) / num_runs
    results_fwd_bwd[name] = avg_ms
    print(f"{name}: {avg_ms:.2f} ms (avg over {num_runs} runs)")

## Forward + Backward plot
fwd_bwd_times = list(results_fwd_bwd.values())

plt.figure(figsize=(10, 6))
bars = plt.bar(names, fwd_bwd_times, color=colors)
for bar, t in zip(bars, fwd_bwd_times):
    plt.text(bar.get_x() + bar.get_width() / 2, bar.get_height(),
             f"{t:.2f} ms", ha="center", va="bottom", fontsize=10)
plt.ylabel("Time (ms)")
plt.title(f"Forward + Backward Pass (avg over {num_runs} runs)")
plt.xticks(rotation=15)
plt.tight_layout()
plt.savefig("benchmark_forward_backward.png", dpi=150)
plt.show()
