import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.nn.attention import sdpa_kernel, SDPBackend

class GroupedQueryAttention(nn.Module) : 
    def __init__(self,d_in,d_out, dropout,num_heads,num_kv_groups, dtype=torch.float16, qkv_bias=False): 

        super().__init__() 

        assert d_out % num_heads == 0 , "d_out should be divisible by num_heads " 
        assert num_heads % num_kv_groups==0 , "num_heads should be divisible by num_kv_groups"

        self.d_out = d_out 
        self.num_heads = num_heads
        self.head_dim = d_out // num_heads
        self.num_kv_groups = num_kv_groups
        self.group_size = num_heads // num_kv_groups
        self.max_tokens = 10000
        self.register_buffer(
            'mask',
            torch.triu(torch.ones(self.max_tokens,self.max_tokens),diagonal=1)
        )

        self.W_key = nn.Linear(d_in,num_kv_groups*self.head_dim, bias=qkv_bias, dtype=dtype)
        self.W_value = nn.Linear(d_in, num_kv_groups*self.head_dim, bias=qkv_bias, dtype=dtype) 
        self.W_query = nn.Linear(d_in,d_out,bias=qkv_bias,dtype=dtype)
        self.W_out_proj = nn.Linear(d_out,d_out,bias=qkv_bias,dtype=dtype) 
        self.dropout = nn.Dropout(dropout)

        self.forward = torch.compile(self.forward)


    def forward(self,x) : 

        b, num_tokens, d_in = x.shape
        
        queries = self.W_query(x)
        values = self.W_value(x) 
        keys = self.W_key(x) 

        queries = queries.view(b,num_tokens,self.num_heads,self.head_dim).transpose(1,2)
        keys_base = keys.view(b,num_tokens, self.num_kv_groups, self.head_dim).transpose(1,2)
        keys = keys_base.repeat_interleave(self.group_size,dim=1)
        values_base = values.view(b,num_tokens,self.num_kv_groups,self.head_dim).transpose(1,2)
        values = values_base.repeat_interleave(self.group_size,dim=1)

        attn_scores = queries @ keys.transpose(2,3)
        attn_scores.masked_fill_(self.mask.bool()[:num_tokens,:num_tokens], -torch.inf)

        attn_weights = torch.softmax(attn_scores / self.head_dim**0.5, dim=-1)
        attn_weights = self.dropout(attn_weights)

        context_vec = (attn_weights @ values).transpose(1,2)
        context_vec = context_vec.contiguous().view(b,num_tokens,self.d_out)
        context_vec = self.W_out_proj(context_vec)

        return context_vec


    

class GQAFlashAttention(nn.Module) :
    def __init__(self,d_in,d_out, dropout,num_heads,num_kv_groups, dtype=torch.float16, qkv_bias=False, compile:bool=False):

        super().__init__()

        assert d_out % num_heads == 0 , "d_out should be divisible by num_heads "
        assert num_heads % num_kv_groups==0 , "num_heads should be divisible by num_kv_groups"

        self.d_out = d_out
        self.num_heads = num_heads
        self.head_dim = d_out // num_heads
        self.num_kv_groups = num_kv_groups
        self.dropout_p = dropout

        self.W_key = nn.Linear(d_in,num_kv_groups*self.head_dim, bias=qkv_bias, dtype=dtype)
        self.W_value = nn.Linear(d_in, num_kv_groups*self.head_dim, bias=qkv_bias, dtype=dtype)
        self.W_query = nn.Linear(d_in,d_out,bias=qkv_bias,dtype=dtype)
        self.W_out_proj = nn.Linear(d_out,d_out,bias=qkv_bias,dtype=dtype)

        if compile:
            self.forward = torch.compile(self.forward)

    def forward(self, x):
        b, num_tokens, _ = x.shape

        q = self.W_query(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)
        k = self.W_key(x).view(b, num_tokens, self.num_kv_groups, self.head_dim).transpose(1, 2)
        v = self.W_value(x).view(b, num_tokens, self.num_kv_groups, self.head_dim).transpose(1, 2)

        with sdpa_kernel(SDPBackend.FLASH_ATTENTION):
            context = F.scaled_dot_product_attention(
                q, k, v,
                is_causal=True,
                dropout_p=self.dropout_p if self.training else 0.0,
                enable_gqa=True,
            )

        context = context.transpose(1, 2).contiguous().view(b, num_tokens, self.d_out)
        return self.W_out_proj(context)


if __name__ == "__main__" :
    d_in = 512
    d_out = 512
    dropout = 0 
    num_heads = 4
    num_kv_groups = 2
    dtype = torch.float16
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    print(f"used device : {device}")
    torch.manual_seed(123) 
    x = torch.randn((5,50,512),device=device,dtype=dtype)
    GQA = GroupedQueryAttention(d_in,d_out,dropout,num_heads,num_kv_groups).to(device)
    context_vec = GQA.forward(x)



    










