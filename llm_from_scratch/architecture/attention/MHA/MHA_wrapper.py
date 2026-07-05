import torch 
import torch.nn as nn 

#causal attention : one head 
# inputs : 
# d_in : input dimension for the head 
# d_out : output dimension of the head 
# context_length : max_token of input prompt 
# dropout : dropout rate after computing attention weights 
# qkv_bias : bias vector on projection matrices 

class CausalAttention(nn.Module) : 
    def __init__(self,d_in:int,d_out:int, context_length:int, dropout:float, qkv_bias:bool=False) : 
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
    def __init__(self, d_in, d_out,context_length, dropout,n_heads, qkv_bias=False)  :
        assert d_out % n_heads == 0, "d_out must be divisle by num_heads"
        super().__init__()
        hidden_head_size = d_out // n_heads
        self.heads = nn.ModuleList(
            [CausalAttention(d_in, hidden_head_size, context_length, dropout, qkv_bias) for _ in range(n_heads)]
        )
        self.out_proj = nn.Linear(d_in, d_out)

    def forward(self,x) : 
        MHA = torch.cat([head(x) for head in self.heads],dim=-1)
        context_vec = self.out_proj(MHA)
        return context_vec



if __name__ == '__main__' : 
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



    





