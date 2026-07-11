import time
import torch
import matplotlib.pyplot as plt
from architecture.attention.MHA.MHA import (
    MHA_wrapper,
    MultiHeadAttention,
    MHAFlashAttention,
    MHAEfficientAttention,
)

batch_size    = 8
seq_len = context_length    = 2048
d_in = d_out  = 1024
num_heads     = 16
num_kv_groups = 4
dtype         = torch.float16
dropout       = 0.0
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

model_factories = {
    "MHA_wrapper": lambda: MHA_wrapper(d_in, d_out, context_length, dropout, num_heads),
    "MHA_wrapper (compiled)": lambda: MHA_wrapper(d_in, d_out, context_length, dropout, num_heads, compile=True),
    "MultiHeadAttention": lambda: MultiHeadAttention(d_in, d_out, context_length, dropout, num_heads),
    "MultiHeadAttention (compiled)": lambda: MultiHeadAttention(d_in, d_out, context_length, dropout, num_heads, compile=True),
    "FlashAttention": lambda: MHAFlashAttention(d_in, d_out, context_length, dropout, num_heads),
    "FlashAttention (compiled)": lambda: MHAFlashAttention(d_in, d_out, context_length, dropout, num_heads, compile=True),
    "EfficientAttention": lambda: MHAEfficientAttention(d_in, d_out, context_length, dropout, num_heads),
    "EfficientAttention (compiled)": lambda: MHAEfficientAttention(d_in, d_out, context_length, dropout, num_heads, compile=True),
}

x = torch.randn(batch_size, seq_len, d_in, device=device, dtype=dtype)

num_runs = 100
results = {}

print("\n--- Forward  ---")


## Forward benchmark
for name, factory in model_factories.items():
    model = factory().to(device=device, dtype=dtype).eval()
    if device.type == "cuda":
        free, total = torch.cuda.mem_get_info()
        print(f"[{name}] after init — Free: {free / 1e9:.2f} GB / Total: {total / 1e9:.2f} GB")

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
    del model
    if device.type == "cuda":
        torch.cuda.empty_cache()
        free, total = torch.cuda.mem_get_info()
        print(f"[{name}] after delete — Free: {free / 1e9:.2f} GB / Total: {total / 1e9:.2f} GB")

## Forward plot
names = list(results.keys())
fwd_times = list(results.values())
colors = [
    "#4c72b0", "#2e4a7a",
    "#55a868", "#367a45",
    "#c44e52", "#8b2e31",
    "#8172b2", "#5a4e8a",
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

for name, factory in model_factories.items():
    model = factory().to(device=device, dtype=dtype).train()
    if device.type == "cuda":
        free, total = torch.cuda.mem_get_info()
        print(f"[{name}] after init — Free: {free / 1e9:.2f} GB / Total: {total / 1e9:.2f} GB")

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
    del model
    if device.type == "cuda":
        torch.cuda.empty_cache()
        free, total = torch.cuda.mem_get_info()
        print(f"[{name}] after delete — Free: {free / 1e9:.2f} GB / Total: {total / 1e9:.2f} GB")

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
