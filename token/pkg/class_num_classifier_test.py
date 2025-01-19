# 테스트 함수
from class_num_classifier import load_raw_data, train
import pandas as pd


def test_load_raw_data():
    df = load_raw_data("class_num_1")
    assert isinstance(df, pd.DataFrame)


def test_train():
    train("class_num_1")
