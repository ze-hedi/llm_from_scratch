from datasets import load_dataset


class FrenchDataset:
    def __init__(self, dataset_id: str, text_col: str = "text", **load_kwargs):
        self.dataset_id = dataset_id
        self.text_col = text_col
        self.ds = load_dataset(dataset_id, **load_kwargs)
        self.train = self.ds["train"]

    def info(self):
        print(f"Dataset: {self.dataset_id}")
        print(f"Structure: {self.ds}")
        print(f"Type: {type(self.ds)}")
        print(f"Columns: {self.train.column_names}")
        print(f"Number of examples: {len(self.train)}")

    def preview(self, n: int = 3, max_chars: int = 300):
        print(f"\n--- First {n} examples ---")
        for i in range(min(n, len(self.train))):
            row = self.train[i]
            print(f"\n[Example {i}]")
            for key, value in row.items():
                if key == self.text_col:
                    print(f"  {key}: {str(value)[:max_chars]}")
                else:
                    print(f"  {key}: {value}")

    def docs(self):
        for row in self.train:
            for paragraph in row[self.text_col].split("\n\n"):
                text = paragraph.strip()
                if text:
                    yield text


ALL_DATASETS = [
    FrenchDataset("nirantk/french-books", text_col="complete_text"),
    FrenchDataset("Volko76/french-classic-books", text_col="text"),
    # FrenchDataset("wikimedia/wikipedia", text_col="text", name="20231101.fr"),
]


def all_docs():
    for ds in ALL_DATASETS:
        yield from ds.docs()


if __name__ == "__main__":
    data_set = FrenchDataset("wikimedia/wikipedia", text_col="text", name="20231101.fr")
    data_set.info() 
    data_set.preview()
