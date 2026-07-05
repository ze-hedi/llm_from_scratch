import sentencepiece as spm

sp = spm.SentencePieceProcessor(model_file="fr_bpe_32k.model")

print(f"Vocab size: {sp.get_piece_size()}")
print(f"BOS id: {sp.bos_id()} → {sp.id_to_piece(sp.bos_id())}")
print(f"EOS id: {sp.eos_id()} → {sp.id_to_piece(sp.eos_id())}")
print(f"UNK id: {sp.unk_id()} → {sp.id_to_piece(sp.unk_id())}")
print(f"EOF id: {sp.piece_to_id('<|eof|>')}")

text = "L'homme est allé au marché pour acheter du pain."
print(f"\nInput: {text}")

tokens = sp.encode(text, out_type=str)
print(f"Tokens: {tokens}")

ids = sp.encode(text, out_type=int)
print(f"IDs: {ids}")

decoded = sp.decode(ids)
print(f"Decoded: {decoded}")
print(f"Round-trip OK: {decoded == text}")
