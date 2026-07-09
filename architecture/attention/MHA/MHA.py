import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.nn.attention import sdpa_kernel, SDPBackend



class CausalAttention(nn.Module) :
    def __init__(self,d_in:int,d_out:int, context_length:int, dropout:float, qkv_bias:bool=False, compile:bool=False) :
        super().__init__()
        self.d_out = d_out
        self.W_query = nn.Linear(d_in,d_out,bias=qkv_bias)
        self.W_key = nn.Linear(d_in,d_out,bias=qkv_bias)
        self.W_value = nn.Linear(d_in,d_out,bias=qkv_bias)
        self.dropout = nn.Dropout(dropout)
        self.register_buffer(
            'mask',
            torch.triu(torch.ones(context_length,context_length),diagonal=1)
        )
        if compile:
            self.forward = torch.compile(self.forward)

    def forward(self,x) : 
        b, num_tokens, d_in = x.shape 
        keys = self.W_key(x)
        queries = self.W_query(x) 
        values = self.W_value(x)

        attn_scores = queries @ keys.transpose(1,2) 
        attn_scores.masked_fill_(
            self.mask.bool()[:num_tokens,:num_tokens],-torch.inf
        )

        attn_weights = torch.softmax(attn_scores/keys.shape[-1]**0.5,dim=-1)
        attn_weights = self.dropout(attn_weights)

        context_vec = attn_weights @ values 
        return context_vec 

class MHA_wrapper(nn.Module) :
    def __init__(self, d_in, d_out,context_length, dropout,n_heads, qkv_bias=False, compile:bool=False)  :
        assert d_out % n_heads == 0, "d_out must be divisle by num_heads"
        super().__init__()
        hidden_head_size = d_out // n_heads
        self.heads = nn.ModuleList(
            [CausalAttention(d_in, hidden_head_size, context_length, dropout, qkv_bias) for _ in range(n_heads)]
        )
        self.out_proj = nn.Linear(d_in, d_out)
        if compile:
            self.forward = torch.compile(self.forward)

    def forward(self,x) : 
        MHA = torch.cat([head(x) for head in self.heads],dim=-1)
        context_vec = self.out_proj(MHA)
        return context_vec



class MultiHeadAttention(nn.Module) :
    def __init__(self,d_in,d_out,context_length,dropout, num_heads, qkv_bias=False, compile:bool=False) :
        super().__init__()
        assert d_out % num_heads == 0, "d_out must be a divisible by num heads "
        self.d_out = d_out
        self.num_heads = num_heads
        self.head_dim = d_out // num_heads

        self.W_query = nn.Linear(d_in, d_out, bias= qkv_bias)
        self.W_value = nn.Linear(d_in, d_out, bias= qkv_bias)
        self.W_key = nn.Linear(d_in,d_out,bias = qkv_bias)
        self.out_proj = nn.Linear(d_out,d_out,bias = qkv_bias)
        self.dropout = nn.Dropout(dropout)
        self.register_buffer("mask",torch.triu(torch.ones(context_length,context_length), diagonal=1))
        if compile:
            self.forward = torch.compile(self.forward)

    def forward(self,x) : 
        b, num_tokens, d_in = x.shape 

        keys = self.W_key(x) 
        queries = self.W_query(x) 
        values = self.W_value(x) 

        keys = keys.view(b,num_tokens,self.num_heads,self.head_dim)
        values = values.view(b,num_tokens,self.num_heads,self.head_dim)
        queries = queries.view(b,num_tokens,self.num_heads,self.head_dim)

        keys = keys.transpose(1,2)
        queries = queries.transpose(1,2)
        values = values.transpose(1,2)

        attn_scores = queries @ keys.transpose(2,3) 

        mask_bool = self.mask.bool()[:num_tokens,:num_tokens]
        
        attn_scores.masked_fill_(mask_bool,-torch.inf) 

        attn_weights = torch.softmax(attn_scores/keys.shape[-1]**0.5,dim=-1)
        attn_weights = self.dropout(attn_weights)

        context_vec = (attn_weights @ values).transpose(1,2)

        context_vec = context_vec.contiguous().view(b,num_tokens,self.d_out)
        context_vec = self.out_proj(context_vec) 

        return context_vec 

class MHAFlashAttention(nn.Module):
    def __init__(self, d_in, d_out, context_length, dropout, num_heads, qkv_bias=False, compile:bool=False):
        super().__init__()
        assert d_out % num_heads == 0, "d_out must be divisible by num heads"
        self.d_out = d_out
        self.num_heads = num_heads
        self.head_dim = d_out // num_heads

        self.W_query = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.W_key = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.W_value = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.out_proj = nn.Linear(d_out, d_out, bias=qkv_bias)
        self.dropout_p = dropout
        if compile:
            self.forward = torch.compile(self.forward)

    def forward(self, x):
        b, num_tokens, d_in = x.shape

        queries = self.W_query(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)
        keys = self.W_key(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)
        values = self.W_value(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)

        with sdpa_kernel(SDPBackend.FLASH_ATTENTION):
            context_vec = F.scaled_dot_product_attention(
                queries, keys, values,
                dropout_p=self.dropout_p if self.training else 0.0,
                is_causal=True,
            )

        context_vec = context_vec.transpose(1, 2).contiguous().view(b, num_tokens, self.d_out)
        return self.out_proj(context_vec)


class MHAEfficientAttention(nn.Module):
    def __init__(self, d_in, d_out, context_length, dropout, num_heads, qkv_bias=False, compile:bool=False):
        super().__init__()
        assert d_out % num_heads == 0, "d_out must be divisible by num heads"
        self.d_out = d_out
        self.num_heads = num_heads
        self.head_dim = d_out // num_heads

        self.W_query = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.W_key = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.W_value = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.out_proj = nn.Linear(d_out, d_out, bias=qkv_bias)
        self.dropout_p = dropout
        if compile:
            self.forward = torch.compile(self.forward)

    def forward(self, x):
        b, num_tokens, d_in = x.shape

        queries = self.W_query(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)
        keys = self.W_key(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)
        values = self.W_value(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)

        with sdpa_kernel(SDPBackend.EFFICIENT_ATTENTION):
            context_vec = F.scaled_dot_product_attention(
                queries, keys, values,
                dropout_p=self.dropout_p if self.training else 0.0,
                is_causal=True,
            )

        context_vec = context_vec.transpose(1, 2).contiguous().view(b, num_tokens, self.d_out)
        return self.out_proj(context_vec)


class MHAMathAttention(nn.Module):
    def __init__(self, d_in, d_out, context_length, dropout, num_heads, qkv_bias=False, compile:bool=False):
        super().__init__()
        assert d_out % num_heads == 0, "d_out must be divisible by num heads"
        self.d_out = d_out
        self.num_heads = num_heads
        self.head_dim = d_out // num_heads

        self.W_query = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.W_key = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.W_value = nn.Linear(d_in, d_out, bias=qkv_bias)
        self.out_proj = nn.Linear(d_out, d_out, bias=qkv_bias)
        self.dropout_p = dropout
        if compile:
            self.forward = torch.compile(self.forward)

    def forward(self, x):
        b, num_tokens, d_in = x.shape

        queries = self.W_query(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)
        keys = self.W_key(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)
        values = self.W_value(x).view(b, num_tokens, self.num_heads, self.head_dim).transpose(1, 2)

        with sdpa_kernel(SDPBackend.MATH):
            context_vec = F.scaled_dot_product_attention(
                queries, keys, values,
                dropout_p=self.dropout_p if self.training else 0.0,
                is_causal=True,
            )

        context_vec = context_vec.transpose(1, 2).contiguous().view(b, num_tokens, self.d_out)
        return self.out_proj(context_vec)


if __name__ == "__main__" :

    #mha implementation test  
    d_in = 768
    d_out = 768 
    context_length = 1028 
    dropout = 0
    num_heads = 12 

    mha = MultiHeadAttention(d_in,d_out,context_length,dropout,num_heads)

    exp = torch.randn(4,2,d_in)

    context_vec = mha.forward(exp)


    #mha wrapper test 
    device = torch.device("cuda")
    torch.manual_seed(122) 
    inputs = torch.randn(2,10,512)
    inputs.to(device)

    d_in = 512
    d_out = 512
    n_heads = 2
    context_length = 1028
    dropout = 0.1

    mha_wrapper = MHA_wrapper(d_in,d_out,context_length,dropout,n_heads) 

    context_vec = mha_wrapper.forward(inputs)


