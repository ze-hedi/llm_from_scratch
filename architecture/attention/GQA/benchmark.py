######## benchmarking gqa against MHA (personal implementation)

import time 
import torch 
import matplotlib.pyplot as plt 
from architecture.attention.GQA.GQA import GroupedQueryAttention, GQAFlashAttention
from architecture.attention.MHA.MHA import MultiHeadAttention, MHAFlashAttention

batch_size    = 8
num_tokens    = 2048       
d_in = d_out  = 1024       
num_heads     = 16         
num_kv_groups = 4          
dtype         = torch.float16   
dropout       = 0.0        
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

num_runs = 100
result = {}

torch.manual_seed(123)
x = torch.randn((batch_size,num_tokens,d_in)).to(device=device,dtype=torch.float16)


print("start running GQA .....")
GQA = GroupedQueryAttention(d_in,d_out,dropout,num_heads,num_kv_groups,dtype).to(device=device).eval()

## forward run benchmark 


##warmup 
if device.type == "cuda" : 
    with torch.no_grad() : 
        GQA(x)
    torch.cuda.synchronize()

GQA_times = [] 

for _ in range(num_runs) :
    start = time.perf_counter()
    with torch.no_grad() : 
        GQA(x) 

    if device.type == "cuda" : 
        torch.cuda.synchronize() 
    GQA_times.append((time.perf_counter() - start)*1000)

avg_ms_GQA = sum(GQA_times) / num_runs 
del GQA


MHA_times = []
print("start running MHA .....")
MHA = MultiHeadAttention(d_in,d_out,num_tokens,dropout,num_heads,compile=True).to(device=device,dtype=dtype)

##warmup 
with torch.no_grad() : 
    MHA(x)

for _ in range(num_runs) : 
    start = time.perf_counter() 
    with torch.no_grad() : 
        MHA(x) 
    
    if device.type == "cuda" : 
        torch.cuda.synchronize() 

    MHA_times.append((time.perf_counter() - start)*1000)

avg_ms_MHA = sum(MHA_times) / num_runs
del MHA

MHA_Flash_times = []
print("start running MHAFlashAttention .....")
MHA_Flash = MHAFlashAttention(d_in,d_out,num_tokens,dropout,num_heads,compile=True).to(device=device,dtype=dtype)

##warmup
with torch.no_grad() :
    MHA_Flash(x)

if device.type == "cuda" :
    torch.cuda.synchronize()

for _ in range(num_runs) :
    start = time.perf_counter()
    with torch.no_grad() :
        MHA_Flash(x)

    if device.type == "cuda" :
        torch.cuda.synchronize()

    MHA_Flash_times.append((time.perf_counter() - start)*1000)

avg_ms_MHA_Flash = sum(MHA_Flash_times) / num_runs
del MHA_Flash

GQA_Flash_times = []
print("start running GQAFlashAttention .....")
GQA_Flash = GQAFlashAttention(d_in,d_out,dropout,num_heads,num_kv_groups,dtype,compile=True).to(device=device).eval()

##warmup
if device.type == "cuda" :
    with torch.no_grad() :
        GQA_Flash(x)
    torch.cuda.synchronize()

for _ in range(num_runs) :
    start = time.perf_counter()
    with torch.no_grad() :
        GQA_Flash(x)

    if device.type == "cuda" :
        torch.cuda.synchronize()

    GQA_Flash_times.append((time.perf_counter() - start)*1000)

avg_ms_GQA_Flash = sum(GQA_Flash_times) / num_runs
del GQA_Flash

names = ["GQA", "GQA Flash", "MHA", "MHA Flash"]
fwd_times = [avg_ms_GQA, avg_ms_GQA_Flash, avg_ms_MHA, avg_ms_MHA_Flash]

colors = ["#4c72b0", "#7bafd4", "#2e4a7a", "#55a868"]

plt.figure(figsize = (14,6))
bars = plt.bar(names,fwd_times,color=colors) 

for bar, t in zip(bars, fwd_times):
    plt.text(bar.get_x() + bar.get_width() / 2, bar.get_height(),
             f"{t:.2f} ms", ha="center", va="bottom", fontsize=9)
plt.ylabel("Time (ms)")
plt.title(f"Forward Pass (avg over {num_runs} runs)")
plt.xticks(rotation=30, ha="right")
plt.tight_layout()
plt.savefig("benchmark_forward.png", dpi=150)
plt.show()


######## Forward + Backward pass benchmark

torch.manual_seed(123)
x_grad = torch.randn((batch_size,num_tokens,d_in), device=device, dtype=dtype, requires_grad=True)

print("start running GQA fwd+bwd .....")
GQA = GroupedQueryAttention(d_in,d_out,dropout,num_heads,num_kv_groups,dtype).to(device=device).train()

##warmup
if device.type == "cuda" :
    out = GQA(x_grad)
    out.sum().backward()
    torch.cuda.synchronize()

GQA_fwdbwd_times = []

for _ in range(num_runs) :
    x_grad.grad = None
    start = time.perf_counter()
    out = GQA(x_grad)
    out.sum().backward()

    if device.type == "cuda" :
        torch.cuda.synchronize()
    GQA_fwdbwd_times.append((time.perf_counter() - start)*1000)

avg_ms_GQA_fwdbwd = sum(GQA_fwdbwd_times) / num_runs
del GQA

print("start running MHA fwd+bwd .....")
MHA = MultiHeadAttention(d_in,d_out,num_tokens,dropout,num_heads,compile=True).to(device=device,dtype=dtype).train()

##warmup
out = MHA(x_grad)
out.sum().backward()
if device.type == "cuda" :
    torch.cuda.synchronize()

MHA_fwdbwd_times = []

for _ in range(num_runs) :
    x_grad.grad = None
    start = time.perf_counter()
    out = MHA(x_grad)
    out.sum().backward()

    if device.type == "cuda" :
        torch.cuda.synchronize()

    MHA_fwdbwd_times.append((time.perf_counter() - start)*1000)

avg_ms_MHA_fwdbwd = sum(MHA_fwdbwd_times) / num_runs
del MHA

print("start running MHAFlashAttention fwd+bwd .....")
MHA_Flash = MHAFlashAttention(d_in,d_out,num_tokens,dropout,num_heads,compile=True).to(device=device,dtype=dtype).train()

##warmup
out = MHA_Flash(x_grad)
out.sum().backward()
if device.type == "cuda" :
    torch.cuda.synchronize()

MHA_Flash_fwdbwd_times = []

for _ in range(num_runs) :
    x_grad.grad = None
    start = time.perf_counter()
    out = MHA_Flash(x_grad)
    out.sum().backward()

    if device.type == "cuda" :
        torch.cuda.synchronize()

    MHA_Flash_fwdbwd_times.append((time.perf_counter() - start)*1000)

avg_ms_MHA_Flash_fwdbwd = sum(MHA_Flash_fwdbwd_times) / num_runs
del MHA_Flash

print("start running GQAFlashAttention fwd+bwd .....")
GQA_Flash = GQAFlashAttention(d_in,d_out,dropout,num_heads,num_kv_groups,dtype,compile=True).to(device=device).train()

##warmup
out = GQA_Flash(x_grad)
out.sum().backward()
if device.type == "cuda" :
    torch.cuda.synchronize()

GQA_Flash_fwdbwd_times = []

for _ in range(num_runs) :
    x_grad.grad = None
    start = time.perf_counter()
    out = GQA_Flash(x_grad)
    out.sum().backward()

    if device.type == "cuda" :
        torch.cuda.synchronize()

    GQA_Flash_fwdbwd_times.append((time.perf_counter() - start)*1000)

avg_ms_GQA_Flash_fwdbwd = sum(GQA_Flash_fwdbwd_times) / num_runs
del GQA_Flash

names_fwdbwd = ["GQA", "GQA Flash", "MHA", "MHA Flash"]
fwdbwd_times = [avg_ms_GQA_fwdbwd, avg_ms_GQA_Flash_fwdbwd, avg_ms_MHA_fwdbwd, avg_ms_MHA_Flash_fwdbwd]

colors = ["#4c72b0", "#7bafd4", "#2e4a7a", "#55a868"]

plt.figure(figsize = (14,6))
bars = plt.bar(names_fwdbwd,fwdbwd_times,color=colors)

for bar, t in zip(bars, fwdbwd_times):
    plt.text(bar.get_x() + bar.get_width() / 2, bar.get_height(),
             f"{t:.2f} ms", ha="center", va="bottom", fontsize=9)
plt.ylabel("Time (ms)")
plt.title(f"Forward + Backward Pass (avg over {num_runs} runs)")
plt.xticks(rotation=30, ha="right")
plt.tight_layout()
plt.savefig("benchmark_forward_backward.png", dpi=150)
plt.show()

