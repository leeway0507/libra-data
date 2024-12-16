import os
import json
from os import listdir
from dataclasses import dataclass, fields, asdict
import tiktoken
import csv
from functools import reduce


@dataclass
class ScrapResult:
    Description: str
    Recommendation: str
    Source: str
    Toc: str
    Url: str
    Isbn: str


@dataclass
class CountResult:
    isbn: str
    count: int


# cl100k_base
# gpt-4-turbo, gpt-4, gpt-3.5-turbo,
# text-embedding-ada-002, text-embedding-3-small, text-embedding-3-large
ENC = tiktoken.get_encoding("cl100k_base")


def load_json_data(json_path: str) -> str:
    with open(json_path, "r") as file:
        try:
            data: ScrapResult = ScrapResult(**json.load(file))
            return data.Toc + data.Description + data.Recommendation
        except:
            raise ValueError("error : ", json_path)


def count_string_token(string: str) -> int:
    return ENC.encode(string).__len__()


def func(isbn_path: str, file_name: str) -> CountResult | None:
    file_path = os.path.join(isbn_path, file_name)
    if file_path.split(".")[-1] != "json":
        return

    string = load_json_data(file_path)
    count = count_string_token(string)
    return CountResult(file_name.rstrip(".json"), count)


if __name__ == "__main__":

    isbn_path = "/Users/yangwoolee/repo/libra-data/data/isbn"
    fileArr = listdir(isbn_path)

    countResultList = list(
        filter(None, map(lambda file_name: func(isbn_path, file_name), fileArr))
    )
    max_num = max(map(lambda x: x.count, countResultList))
    min_num = min(map(lambda x: x.count, countResultList))
    totalToken = reduce(lambda acc, cur: acc + cur.count, countResultList, 0)
    totalTokenWithMaxToken = reduce(
        lambda acc, cur: acc + min(5000, cur.count), countResultList, 0
    )
    print(
        "max",
        max_num,
        [(isbn) for isbn in countResultList if isbn.count == max_num],
    )
    print("min", min_num)
    print(
        "avergage",
        totalToken / countResultList.__len__(),
    )
    print("total token", f"{totalToken:,}")
    print("totalTokenWithMaxToken", f"{totalTokenWithMaxToken:,}")

    # with open(f"{os.path.abspath(os.getcwd())}/token/count.csv", "w+") as f:
    #     flds = [fld.name for fld in fields(CountResult)]
    #     w = csv.DictWriter(f, flds)
    #     w.writeheader()
    #     w.writerows([asdict(prop) for prop in countResultList])
