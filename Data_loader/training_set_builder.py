from datasets import load_dataset
import sys


class TrainingSetBuilder:
    def __init__(self, dataset_name, text_column="text", subset_name=None):
        self.dataset_name = dataset_name
        self.text_column = text_column
        self.subset_name = subset_name
        self.output_filename = dataset_name.split("/")[-1] + ".txt"

    def build(self):
        if self.subset_name:
            dataset = load_dataset(self.dataset_name, self.subset_name)
        else:
            dataset = load_dataset(self.dataset_name)

        rows = []
        for split in dataset:
            for row in dataset[split]:
                rows.append(row[self.text_column])

        content = "<|eos|>".join(rows)

        with open(self.output_filename, "w") as f:
            f.write(content)

        print(f"Wrote {len(rows)} rows to {self.output_filename}")

    def build_big(self, chunk_size=50_000):
        import gc
        import os

        if self.subset_name:
            dataset = load_dataset(self.dataset_name, self.subset_name, streaming=True)
        else:
            dataset = load_dataset(self.dataset_name, streaming=True)

        output_dir = self.dataset_name.split("/")[-1]
        os.makedirs(output_dir, exist_ok=True)

        file_index = 0
        row_count = 0
        total_rows = 0
        buffer = []

        for split in dataset:
            for row in dataset[split]:
                buffer.append(row[self.text_column])
                row_count += 1

                if row_count >= chunk_size:
                    filename = os.path.join(output_dir, f"{output_dir}_{file_index:04d}.txt")
                    with open(filename, "w") as f:
                        f.write("<|eos|>".join(buffer))
                    print(f"Wrote {row_count} rows to {filename}")
                    total_rows += row_count
                    file_index += 1
                    row_count = 0
                    buffer = []
                    gc.collect()

        if buffer:
            filename = os.path.join(output_dir, f"{output_dir}_{file_index:04d}.txt")
            with open(filename, "w") as f:
                f.write("<|eos|>".join(buffer))
            print(f"Wrote {row_count} rows to {filename}")
            total_rows += row_count

        print(f"Done. Wrote {total_rows} rows across {file_index + 1} files in {output_dir}/")


if __name__ == "__main__":
    dataset_name = sys.argv[1]
    text_column = sys.argv[2] if len(sys.argv) > 2 else "text"
    subset_name = sys.argv[3] if len(sys.argv) > 3 else None
    mode = sys.argv[4] if len(sys.argv) > 4 else None
    builder = TrainingSetBuilder(dataset_name, text_column, subset_name)
    if mode == "big":
        builder.build_big()
    else:
        builder.build()
