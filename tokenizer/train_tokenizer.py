import os
import tempfile
from collections.abc import Iterator

from tokenizers import Tokenizer, pre_tokenizers, models, trainers, decoders
from tokenizers.pre_tokenizers import Split
from tokenizers.normalizers import NFC


def corpus_iterator(files: list[str]) -> Iterator[str]:
    for file_path in files:
        with open(file_path, "r", encoding="utf-8") as f:
            yield f.read()


class TokenizerTrainer:
    def __init__(
        self,
        training_corpus: str | Iterator[str],
        special_tokens: list[str],
        pattern: str,
        output: str,
        vocab_size: int = 32768,
        max_token_length: int = 64,
    ):
        self.training_corpus = training_corpus
        self.special_tokens = special_tokens
        self.pattern = pattern
        self.output = output
        self.vocab_size = vocab_size
        self.max_token_length = max_token_length

    def _build_tokenizer_and_trainer(self) -> tuple[Tokenizer, trainers.BpeTrainer]:
        tokenizer = Tokenizer(models.BPE())
        tokenizer.normalizer = NFC()
        tokenizer.pre_tokenizer = pre_tokenizers.Sequence([
            Split(pattern=self.pattern, behavior="isolated", invert=False),
            pre_tokenizers.ByteLevel(add_prefix_space=False, use_regex=False),
        ])
        tokenizer.decoder = decoders.ByteLevel()

        bpe_trainer = trainers.BpeTrainer(
            vocab_size=self.vocab_size,
            special_tokens=self.special_tokens,
            show_progress=True,
            max_token_length=self.max_token_length,
        )
        return tokenizer, bpe_trainer

    def _save(self, tokenizer: Tokenizer) -> None:
        tokenizer.save(self.output)
        print(f"Training complete — {self.output} written. Vocab size: {tokenizer.get_vocab_size()}")

    def train_from_txt(self) -> Tokenizer:
        tokenizer, bpe_trainer = self._build_tokenizer_and_trainer()
        print("Training tokenizer from text file...")
        tokenizer.train([self.training_corpus], trainer=bpe_trainer)
        self._save(tokenizer)
        return tokenizer

    def train_from_iterator(self) -> Tokenizer:
        tokenizer, bpe_trainer = self._build_tokenizer_and_trainer()
        print("Training tokenizer from iterator...")
        tokenizer.train_from_iterator(self.training_corpus, trainer=bpe_trainer)
        self._save(tokenizer)
        return tokenizer

    def train(self, source: str = "from_txt") -> Tokenizer:
        if source == "from_txt":
            return self.train_from_txt()
        elif source == "from_iterator":
            return self.train_from_iterator()
        else:
            raise ValueError(f"Unknown source: {source!r}. Use 'from_txt' or 'from_iterator'.")


if __name__ == "__main__":
    FRENCH_PATTERN = r"""[^\r\n\p{L}\p{N}]?[\p{L}\p{M}]+(?:['\u2019][\p{L}\p{M}]+)*|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+"""

    corpus_file = "tokenizer_training_set/tokenizer_training_corpus_small.txt"
    size_mb = os.path.getsize(corpus_file) / (1024 ** 2)
    iterator = corpus_iterator([corpus_file])

    print(f"training 32k french tokenizer on {size_mb:.0f}mb")
    trainer = TokenizerTrainer(
        training_corpus=corpus_file,
        special_tokens=["<|pad|>", "<|eos|>", "<|user|>", "<|assistant|>"],
        pattern=FRENCH_PATTERN,
        output=f"trained_tokenizers/fr_bpe_32k_{size_mb:.0f}.json",
    )
    trainer.train()
    del trainer

    corpus_files_2 = [corpus_file, "tokenizer_training_set/french-books.txt"]
    size_mb_2 = sum(os.path.getsize(f) / (1024 ** 2) for f in corpus_files_2)
    print(f"training 32k french tokenizer on {size_mb_2:.0f}mb")
    with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=True, encoding="utf-8") as tmp:
        for f in corpus_files_2:
            with open(f, "r", encoding="utf-8") as src:
                tmp.write(src.read())
        tmp.flush()
        trainer = TokenizerTrainer(
            training_corpus=tmp.name,
            special_tokens=["<|pad|>", "<|eos|>", "<|user|>", "<|assistant|>"],
            pattern=FRENCH_PATTERN,
            output=f"trained_tokenizers/fr_bpe_32k_{size_mb_2:.0f}.json",
        )
        trainer.train()
        del trainer

    corpus_files_3 = [*corpus_files_2, "tokenizer_training_set/french-classic-books.txt"]
    size_mb_3 = sum(os.path.getsize(f) / (1024 ** 2) for f in corpus_files_3)
    print(f"training 32k french tokenizer on {size_mb_3:.0f}mb")
    with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=True, encoding="utf-8") as tmp:
        for f in corpus_files_3:
            with open(f, "r", encoding="utf-8") as src:
                tmp.write(src.read())
        tmp.flush()
        trainer = TokenizerTrainer(
            training_corpus=tmp.name,
            special_tokens=["<|pad|>", "<|eos|>", "<|user|>", "<|assistant|>"],
            pattern=FRENCH_PATTERN,
            output=f"trained_tokenizers/fr_bpe_32k_{size_mb_3:.0f}.json",
        )
        trainer.train()
        del trainer

    corpus_files_4 = [*corpus_files_3, "tokenizer_training_set/wikipedia/wikipedia_0000.txt"]
    size_mb_4 = sum(os.path.getsize(f) / (1024 ** 2) for f in corpus_files_4)
    print(f"training 32k french tokenizer on {size_mb_4:.0f}mb")
    with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=True, encoding="utf-8") as tmp:
        for f in corpus_files_4:
            with open(f, "r", encoding="utf-8") as src:
                tmp.write(src.read())
        tmp.flush()
        trainer = TokenizerTrainer(
            training_corpus=tmp.name,
            special_tokens=["<|pad|>", "<|eos|>", "<|user|>", "<|assistant|>"],
            pattern=FRENCH_PATTERN,
            output=f"trained_tokenizers/fr_bpe_32k_{size_mb_4:.0f}.json",
        )
        trainer.train()
        del trainer

    corpus_files_5 = [*corpus_files_4, "tokenizer_training_set/wikipedia/wikipedia_0001.txt"]
    size_mb_5 = sum(os.path.getsize(f) / (1024 ** 2) for f in corpus_files_5)
    print(f"training 32k french tokenizer on {size_mb_5:.0f}mb")
    with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=True, encoding="utf-8") as tmp:
        for f in corpus_files_5:
            with open(f, "r", encoding="utf-8") as src:
                tmp.write(src.read())
        tmp.flush()
        trainer = TokenizerTrainer(
            training_corpus=tmp.name,
            special_tokens=["<|pad|>", "<|eos|>", "<|user|>", "<|assistant|>"],
            pattern=FRENCH_PATTERN,
            output=f"trained_tokenizers/fr_bpe_32k_{size_mb_5:.0f}.json",
        )
        trainer.train()
        del trainer
