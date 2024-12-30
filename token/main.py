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
class Books:
    id: str
    isbn: str
    title: str
    author: str
    publisher: str
    publication_year: str
    set_isbn: str
    volume: str
    image_url: str
    toc: str
    recommendation: str
    source: str
    url: str
    description: str
    vector_search: str


@dataclass
class CountResult:
    isbn: str
    count: int


# cl100k_base
# gpt-4-turbo, gpt-4, gpt-3.5-turbo,
# text-embedding-ada-002, text-embedding-3-small, text-embedding-3-large
ENC = tiktoken.get_encoding("cl100k_base")


def count_string_token(string: str) -> int:
    return ENC.encode(string).__len__()


def create_book_count():
    with open(f"{os.path.abspath(os.getcwd())}/book.csv", "r+") as f:
        flds = [fld.name for fld in fields(Books)]
        reader = csv.DictReader(f, flds)
        countResultList = []
        for i, line in enumerate(reader):
            if i == 0:
                continue
            stringRawCount = " ".join(
                [
                    line["isbn"],
                    line["title"],
                    line["author"],
                    line["toc"],
                    line["recommendation"],
                    line["description"],
                ]
            ).__len__()
            countResultList.append(CountResult(line["isbn"], stringRawCount))

        with open(f"{os.path.abspath(os.getcwd())}/book_count.csv", "w+") as f:
            flds = [fld.name for fld in fields(CountResult)]
            w = csv.DictWriter(f, flds)
            w.writeheader()
            w.writerows([asdict(prop) for prop in countResultList])


def count():
    total = 0
    maxCount = 0
    with open(f"{os.path.abspath(os.getcwd())}/book_count.csv", "r+") as f:
        flds = [fld.name for fld in fields(CountResult)]
        reader = csv.DictReader(f, flds)
        for i, line in enumerate(reader):
            if i == 0:
                continue
            x = line["count"]
            if type(x) == str:
                intx = int(x)
                total += intx
                maxCount = max(maxCount, intx)

        print("total token", f"{total:,}")
        print(
            "totak price",
            "$",
            (total / 1000000) * 0.02,
            f"(kor:{(total / 1000000) * 0.02*1450})",
        )
        print("maxCount", maxCount)
        print("average token", total / reader.line_num)


if __name__ == "__main__":
    create_book_count()
    count()

    # total token 47,113,762
    # totak price $ 0.9422752400000001 (kor:1366.2990980000002)
    # maxCount 16488
    # average token 69.58487539674687
