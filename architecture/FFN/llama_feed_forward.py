import torch.nn as nn
import torch.nn.functional as F


def ffn_dim(d_model, multiple_of=64):
    h = int(8 * d_model / 3)
    return multiple_of * ((h + multiple_of - 1) // multiple_of)


class SwiGLUMLP(nn.Module):
    def __init__(self, d_model, multiple_of=64):
        super().__init__()
        d_ff = ffn_dim(d_model, multiple_of)
        self.gate_proj = nn.Linear(d_model, d_ff, bias=False)
        self.up_proj   = nn.Linear(d_model, d_ff, bias=False)
        self.down_proj = nn.Linear(d_ff, d_model, bias=False)

    def forward(self, x):
        return self.down_proj(F.silu(self.gate_proj(x)) * self.up_proj(x))
