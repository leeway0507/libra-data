"""
select k.title, b.isbn, b.embedding, l.class_nums from bookembedding b join  
(select l.isbn, ARRAY_AGG(l.class_num) as class_nums 
from libsbooks l    
GROUP BY l.isbn) as l 
on l.isbn = b.isbn 
join books k 
on b.isbn = k.isbn;
"""

import os
import pandas as pd
from sklearn.model_selection import train_test_split

file_path = "/Users/yangwoolee/repo/libra-data/token/data/classifier"


def split_test_file(file_name: str):
    df = load_raw_data(file_name)

    train, test = train_test_split(df, test_size=0.2)
    if isinstance(train, pd.DataFrame):
        train.to_parquet(
            os.path.join(file_path, file_name + "_train" + ".parquet.gzip"),
            compression="gzip",
            engine="fastparquet",
        )
    if isinstance(test, pd.DataFrame):
        test.to_parquet(
            os.path.join(file_path, file_name + "_test" + ".parquet.gzip"),
            compression="gzip",
            engine="fastparquet",
        )


def load_raw_data(file_name: str) -> pd.DataFrame:
    parquetFilePath = os.path.join(file_path, file_name + ".parquet.gzip")
    isParquetExist = os.path.isfile(parquetFilePath)
    if isParquetExist:
        print(f"parquet exists. Get {file_name}.parquet.gzip")
        return pd.read_parquet(parquetFilePath)

    print(f"parquet does not exist. Create {file_name}.parquet.gzip")

    df = pd.read_csv(parquetFilePath.replace(".parquet.gzip", ".csv"))
    df.to_parquet(parquetFilePath, compression="gzip", engine="fastparquet")
    return df
