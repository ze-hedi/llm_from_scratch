from tokenizers import Tokenizer 

class DataLoader : 
    def __init__(self,tokenizer_file,training_corpus_file,context_size) :
        self.tokenizer = Tokenizer.from_file(tokenizer_file) 
        self.token_ids = None
        self.training_tokens = []
        self.target_tokens = [] 

        print("reading training corpus .... ")
        with open(training_corpus_file,"r") as file :
            training_corpus = file.read() 
            self.token_ids = self.tokenizer.encode(training_corpus) 
        
        print("building training strides .....")
        num_tokens = self.token_ids 
        for i in range(0,num_tokens - context_size, context_size) : 
            self.training_tokens.append(self.token_ids[i:i+context_size]) 
            self.target_tokens.append(self.token_ids[i+1:i+context_size+1])

    def get_data(self) : 
        return self.training_tokens, self.target_tokens 

        
         


            

        
