import torch 
import torch.nn as nn 

from gpt_transformer_block import GPTTransformerBlock 
from normalization.layer_norm import LayerNorm
from tokenizers import Tokenizer


class GPTModel(nn.Module) : 
    def __init__(self,cfg) : 
        super().__init__() 
        self.tok_emb = nn.Embedding(cfg["vocab_size"],cfg["emb_dim"]) 
        self.pos_emb = nn.Embedding(cfg["context_length"],cfg["emb_dim"])
        self.drop_emb = nn.Dropout(cfg["drop_rate"]) 

        self.transformer_blocks = nn.Sequential(
            *[GPTTransformerBlock(cfg) for _ in range(cfg["n_layers"])]
        )

        self.final_norm = LayerNorm(cfg["emb_dim"]) 
        self.out_head = nn.Linear(cfg["emb_dim"],cfg["vocab_size"],bias=False) 


    def forward(self, in_idx) : 
        batch_size, seq_len = in_idx.shape 
        tok_embeddings = self.tok_emb(in_idx) 

        #since we the data loader will not be equal context length everytime, 
        #the arange should be changed so we cover all the rows during training
        pos_embeddings = self.pos_emb(torch.arange(seq_len, device=in_idx.device))
        x = tok_embeddings + pos_embeddings
        x = self.drop_emb(x)
        x = self.transformer_blocks(x)
        x = self.final_norm(x)
        logits = self.out_head(x)
        return logits   

def generate_text(model,idx, max_new_tokens, context_size) : 

    num_input_tokens = len(idx)
    assert num_input_tokens + max_new_tokens< context_size , "number of input tokens exceed context window"
    for _ in range(max_new_tokens) : 
        with torch.no_grad() : 
            logits = model(idx) 

        logits = logits[:,-1,:]
        probas = torch.softmax(logits, dim=-1)
        next_token = torch.multinomial(probas, num_samples=1)

        idx = torch.cat((idx,next_token),dim=1) 

    return idx 



if __name__ == "__main__" : 
    config = {
        "vocab_size" : 32768,
        "context_length" : 1024, 
        "emb_dim" : 768, 
        "n_heads" : 12, 
        "n_layers" : 12 ,
        "drop_rate" : 0.1, 
        "qkv_bias" : False
    }

    torch.manual_seed(456) 
    model = GPTModel(config) 
    model.eval() 

    start_context = "Hello, I am "
    tok = Tokenizer.from_file("../tokenizer/fr_bpe_32k.json")
    encoded_context = tok.encode(start_context).ids
    torch_encoded_context = torch.tensor(encoded_context).unsqueeze(0)

    print("encoded_context ", encoded_context)

    new_sentence_idx = generate_text(model,torch_encoded_context,2,1024)
    print("new_sentence_idx " , new_sentence_idx)






