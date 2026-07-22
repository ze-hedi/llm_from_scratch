import torch 
import torch.nn as nn 

from architecture.attention.GQA.GQA import GQAFlashAttention
from architecture.normalization.RMSNorm import RMSNorm 
from architecture.FFN.llama_feed_forward import SwiGLUMLP

class LlamaTransformerBlock(nn.Module) : 
    def __init__(self,cfg) :
        super().__init__()
        self.d_model = cfg["d_model"]
        self.d_ff = cfg["d_ff"]
        self.dropout = cfg.get("dropout",0)
        self.num_heads = cfg["num_heads"]
        self.num_kv_groups = cfg["num_kv_heads"]
        self.dtype = cfg.get("dtype",torch.float16)
        self.context_windows = cfg.get("context_window",2048)

        self.GQA = GQAFlashAttention(self.d_model,self.d_model,self.dropout,self.num_heads,
                        self.num_kv_groups,dtype=self.dtype)

        self.feed_forward = SwiGLUMLP(self.d_model, self.d_ff)

        self.RMSNorm_attention = RMSNorm(self.d_model)
        self.RMSNorm_FFN = RMSNorm(self.d_model)


    def forward(self,x) : 
        shortcut = x 
        x = self.RMSNorm_attention(x) 
        x = self.GQA(x)
        x = x + shortcut 

        shortcut = x 
        x = self.RMSNorm_FFN(x)
        x = self.feed_forward(x)
        return x + shortcut


if __name__ == "__main__" : 
    cfg = {
        "d_model" : 768 ,
        "d_ff" : 2048 ,
        "num_heads" : 12 ,
        "num_kv_heads" : 4 ,
        "context_window" : 1024
    }
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    transformer_block = LlamaTransformerBlock(cfg).to(device=device,dtype=torch.float16)
    print(f"instantiating llama block on {device}")
    x = torch.randn((6,50,768)).to(device=device,dtype=torch.float16)

    res = transformer_block(x)
    print("forward run succeeded")



