import torch 
from tokenizers import Tokenizer
import sys 


class NaiveDataLoader : 
    def __init__(self,tokens, seq_length) : 
        self.input_ids = [] 
        self.target_ids = [] 

        for i in range(0,len(tokens) - seq_length, seq_length)  : 
            self.input_ids.append(torch.tensor(tokens[i:i+seq_length])) 
            self.target_ids.append(torch.tensor(tokens[i+1:i+seq_length+1]))

    def __len__(self) : 
        return len(self.input_ids) 

    def __getitem__(self,idx) : 
        return self.input_ids[idx] ; self.target_ids[idx]

if __name__ == "__main__" : 
    tok = Tokenizer.from_file(sys.argv[1]) 

    with open(sys.argv[2],"r")  as training_set : 
        training =  training_set.read()
        encoding_training = tok.encode(training).ids 
        print(type(encoding_training))
        print(f"number of tokens {len(encoding_training)}")



