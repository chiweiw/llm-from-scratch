import time
import random
import logging
import json
import os
import pandas as pd
from datetime import datetime
from DrissionPage import ChromiumPage, ChromiumOptions
import config

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler("spider.log", encoding="utf-8"),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger(__name__)


class ZhihuSpider:
    def __init__(self, user_id):
        self.user_id = user_id
        self.page = None
        self.answers_data = []
        self.answer_ids = set()
        self.current_page = 1
        self.browser_data_dir = config.USER_DATA_DIR
        self.max_answer_count = config.MAX_ANSWER_COUNT

    def load_progress(self):
        logger.info("加载断点续传记录...")
        try:
            if os.path.exists(config.PROGRESS_FILE):
                with open(config.PROGRESS_FILE, "r", encoding="utf-8") as f:
                    progress = json.load(f)
                    self.answer_ids = set(progress.get("answer_ids", []))
                    self.current_page = progress.get("current_page", 1)
                    logger.info(
                        f"已加载 {len(self.answer_ids)} 条历史记录，从第 {self.current_page} 页开始"
                    )
                    return True
        except Exception as e:
            logger.error(f"加载断点记录失败: {e}")
        return False

    def save_progress(self):
        logger.info("保存断点续传记录...")
        try:
            progress = {
                "user_id": self.user_id,
                "answer_ids": list(self.answer_ids),
                "current_page": self.current_page,
                "last_update": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            }
            with open(config.PROGRESS_FILE, "w", encoding="utf-8") as f:
                json.dump(progress, f, ensure_ascii=False, indent=2)
            logger.info(f"已保存 {len(self.answer_ids)} 条记录到进度文件")
        except Exception as e:
            logger.error(f"保存断点记录失败: {e}")

    def init_browser(self):
        logger.info("正在初始化浏览器...")
        # 配置浏览器选项（DrissionPage 4.x 版本）
        co = ChromiumOptions()
        co.set_user_data_path(self.browser_data_dir)
        co.headless(False)
        co.set_user_agent(config.USER_AGENT)

        self.page = ChromiumPage(addr_or_opts=co)
        logger.info("浏览器初始化完成")

    def check_login(self):
        logger.info("检查登录状态...")
        self.page.get("https://www.zhihu.com")
        time.sleep(2)

        login_button = self.page.ele("text:登录", timeout=2)

        if login_button:
            logger.info("未登录，等待手动扫码登录...")
            logger.info(f"请在 {config.LOGIN_WAIT_TIME} 秒内完成扫码登录")

            start_time = time.time()
            while time.time() - start_time < config.LOGIN_WAIT_TIME:
                if not self.page.ele("text:登录", timeout=1):
                    logger.info("登录成功！")
                    time.sleep(2)
                    return True
                time.sleep(1)

            logger.error("登录超时，请重新运行程序")
            return False
        else:
            logger.info("已登录")
            return True

    def setup_network_listener(self):
        logger.info("设置网络监听...")
        # 临时监听所有请求以便调试
        self.page.listen.start()
        logger.info(f"✓ 已开始监听所有网络请求（调试模式）")
        logger.info(f"原配置的API模式: {config.ANSWERS_API_PATTERN}")
        logger.info("提示：滚动页面时将捕获所有API请求，帮助诊断正确的API模式")

    def navigate_to_answers_page(self):
        url = f"https://www.zhihu.com/people/{self.user_id}/answers"
        logger.info(f"正在访问: {url}")
        self.page.get(url)
        time.sleep(3)

    def random_sleep(self):
        sleep_time = random.uniform(config.RANDOM_SLEEP_MIN, config.RANDOM_SLEEP_MAX)
        logger.info(f"随机等待 {sleep_time:.2f} 秒...")
        time.sleep(sleep_time)

    def scroll_page(self):
        logger.info("滚动页面...")
        self.page.scroll.down(500)
        time.sleep(2)
        logger.info("页面滚动完成，等待API响应...")

    def get_answers_data(self):
        logger.info("等待捕获API响应...")
        res = self.page.listen.wait(timeout=10)
        if res:
            logger.info(f"✓ 捕获到响应包: {res.url}")
            try:
                data = res.response.body
                logger.info(f"响应数据类型: {type(data)}")
                if data and "data" in data:
                    answers = data.get("data", [])
                    logger.info(f"✓ 获取到 {len(answers)} 条回答")
                    return answers
                else:
                    logger.warning(
                        f"响应中没有'data'字段，响应keys: {data.keys() if isinstance(data, dict) else 'N/A'}"
                    )
            except Exception as e:
                logger.error(f"✗ 解析响应数据失败: {e}")
                import traceback

                logger.error(traceback.format_exc())
        else:
            logger.warning("✗ 未捕获到API响应（超时10秒）")
        return []

    def parse_answer(self, answer):
        try:
            question = answer.get("question", {})
            question_title = question.get("title", "")
            question_id = question.get("id", "")

            answer_id = answer.get("id", "")
            excerpt = answer.get("excerpt", "")
            content = answer.get("content", "")
            voteup_count = answer.get("voteup_count", 0)
            comment_count = answer.get("comment_count", 0)
            created_time = answer.get("created", 0)

            created_date = (
                datetime.fromtimestamp(created_time).strftime("%Y-%m-%d %H:%M:%S")
                if created_time
                else ""
            )

            answer_url = (
                f"https://www.zhihu.com/question/{question_id}/answer/{answer_id}"
            )

            return {
                "Question Title": question_title,
                "Answer ID": answer_id,
                "Excerpt": excerpt,
                "Content": content,
                "Vote Count": voteup_count,
                "Comment Count": comment_count,
                "Create Time": created_date,
                "URL": answer_url,
            }
        except Exception as e:
            logger.error(f"解析回答数据失败: {e}")
            return None

    def save_to_csv(self):
        if not self.answers_data:
            logger.warning("没有数据可保存")
            return

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"{self.user_id}_answers_{timestamp}.csv"
        filepath = f"{config.OUTPUT_DIR}/{filename}"

        df = pd.DataFrame(self.answers_data)
        df.to_csv(filepath, index=False, encoding="utf-8-sig")
        logger.info(f"已保存 {len(self.answers_data)} 条数据到: {filepath}")

    def click_next_page(self):
        try:
            next_button = self.page.ele("text:下一页", timeout=5)
            if next_button:
                self.random_sleep()
                next_button.click()
                self.current_page += 1
                logger.info(f"点击下一页，当前页码: {self.current_page}")
                time.sleep(2)
                return True
            else:
                logger.info('未找到"下一页"按钮，可能已到达最后一页')
                return False
        except Exception as e:
            logger.error(f"点击下一页失败: {e}")
            return False

    def handle_captcha(self):
        logger.warning("检测到可能的验证码，请手动处理...")
        logger.warning("处理完成后，程序将自动继续")
        input("按回车键继续...")

    def check_forbidden(self):
        if self.page.url.endswith("captcha") or "403" in self.page.url:
            logger.error("检测到403错误或验证码页面")
            logger.error("请切换IP（开关飞行模式）后按回车继续...")
            input("按回车键继续...")
            return True
        return False

    def check_limit_reached(self):
        if (
            self.max_answer_count > 0
            and len(self.answers_data) >= self.max_answer_count
        ):
            logger.info(f"已达到配置的爬取数量限制: {self.max_answer_count} 条")
            return True
        return False

    def run(self):
        try:
            self.init_browser()

            if not self.check_login():
                return

            self.load_progress()

            self.setup_network_listener()
            self.navigate_to_answers_page()

            retry_count = 0

            while retry_count < config.MAX_RETRY:
                try:
                    self.scroll_page()

                    if self.check_forbidden():
                        self.navigate_to_answers_page()
                        continue

                    answers = self.get_answers_data()

                    logger.info(
                        f"当前已收集回答数: {len(self.answers_data)}, 已去重ID数: {len(self.answer_ids)}"
                    )

                    if answers:
                        logger.info(f"开始处理 {len(answers)} 条回答...")
                        new_answers = 0
                        for answer in answers:
                            answer_id = answer.get("id", "")
                            if answer_id and answer_id not in self.answer_ids:
                                self.answer_ids.add(answer_id)
                                parsed = self.parse_answer(answer)
                                if parsed:
                                    self.answers_data.append(parsed)
                                    new_answers += 1
                                    logger.info(
                                        f"  + 新增回答 #{new_answers}: {parsed['Question Title'][:30]}..."
                                    )

                                    if self.check_limit_reached():
                                        logger.info("达到数量限制，停止采集")
                                        self.save_to_csv()
                                        self.save_progress()
                                        return
                            else:
                                if answer_id:
                                    logger.debug(f"  - 跳过重复回答 ID: {answer_id}")

                        logger.info(
                            f"本页新增 {new_answers} 条回答，总计 {len(self.answers_data)} 条"
                        )

                        if new_answers == 0:
                            logger.info("本页无新数据，可能已到达最后一页")
                            break

                        self.save_to_csv()
                        self.save_progress()

                    if not self.click_next_page():
                        logger.info("翻页结束，采集完成")
                        break

                    retry_count = 0

                except Exception as e:
                    logger.error(f"处理失败: {e}")
                    retry_count += 1
                    if retry_count < config.MAX_RETRY:
                        logger.info(
                            f"等待 {config.RETRY_INTERVAL} 秒后重试 ({retry_count}/{config.MAX_RETRY})"
                        )
                        time.sleep(config.RETRY_INTERVAL)
                    else:
                        logger.error("达到最大重试次数，程序退出")
                        break

        except KeyboardInterrupt:
            logger.info("用户中断程序")
        except Exception as e:
            logger.error(f"程序异常: {e}")
        finally:
            if self.page:
                self.save_to_csv()
                self.save_progress()
                logger.info("正在关闭浏览器...")
                self.page.quit()
                logger.info("程序结束")


if __name__ == "__main__":
    user_id = config.TARGET_USER_ID
    logger.info(f"开始采集用户: {user_id} 的所有回答")
    if config.MAX_ANSWER_COUNT > 0:
        logger.info(f"配置的爬取数量限制: {config.MAX_ANSWER_COUNT} 条")
    else:
        logger.info("未配置数量限制，将爬取所有回答")

    spider = ZhihuSpider(user_id)
    spider.run()
