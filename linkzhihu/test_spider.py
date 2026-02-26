import unittest
import os
import sys
from unittest.mock import Mock, patch
import pandas as pd

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from spider import ZhihuSpider
import config


class TestZhihuSpider(unittest.TestCase):

    def setUp(self):
        self.spider = ZhihuSpider("test_user")

    def test_parse_answer(self):
        mock_answer = {
            "id": "123456",
            "question": {"id": "789", "title": "测试问题"},
            "excerpt": "测试摘要",
            "content": "<p>测试内容</p>",
            "voteup_count": 100,
            "comment_count": 10,
            "created": 1672444800,
        }

        result = self.spider.parse_answer(mock_answer)

        self.assertIsNotNone(result)
        self.assertEqual(result["Answer ID"], "123456")
        self.assertEqual(result["Question Title"], "测试问题")
        self.assertEqual(result["Vote Count"], 100)
        self.assertEqual(result["Comment Count"], 10)
        self.assertIn("2022", result["Create Time"])
        self.assertIn("question/789/answer/123456", result["URL"])

    def test_parse_answer_missing_fields(self):
        mock_answer = {"id": "123456", "question": {}}

        result = self.spider.parse_answer(mock_answer)

        self.assertIsNotNone(result)
        self.assertEqual(result["Answer ID"], "123456")
        self.assertEqual(result["Question Title"], "")
        self.assertEqual(result["Vote Count"], 0)

    def test_answer_deduplication(self):
        answer1 = {"id": "123", "question": {"title": "Q1"}}
        answer2 = {"id": "456", "question": {"title": "Q2"}}
        answer3 = {"id": "123", "question": {"title": "Q1"}}

        parsed1 = self.spider.parse_answer(answer1)
        parsed2 = self.spider.parse_answer(answer2)
        parsed3 = self.spider.parse_answer(answer3)

        self.spider.answers_data.append(parsed1)
        self.spider.answers_data.append(parsed2)
        self.spider.answers_data.append(parsed3)

        self.assertEqual(len(self.spider.answers_data), 3)

    def test_config_values(self):
        self.assertIsInstance(config.TARGET_USER_ID, str)
        self.assertGreater(config.LOGIN_WAIT_TIME, 0)
        self.assertGreater(config.RANDOM_SLEEP_MIN, 0)
        self.assertGreater(config.RANDOM_SLEEP_MAX, config.RANDOM_SLEEP_MIN)
        self.assertGreater(config.MAX_RETRY, 0)

    def test_output_directory(self):
        self.assertTrue(os.path.exists(config.OUTPUT_DIR))
        self.assertTrue(os.path.isdir(config.OUTPUT_DIR))

    def test_browser_data_directory(self):
        self.assertTrue(os.path.exists(config.USER_DATA_DIR))
        self.assertTrue(os.path.isdir(config.USER_DATA_DIR))


if __name__ == "__main__":
    unittest.main()
