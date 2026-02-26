DEV_DB = ("192.168.66.40", 3306, "hxlc", "hxlc", "hxlc_wmps_dev")
TEST_DB = ("192.168.66.40", 3306, "hxlctest2", "hxlctest2", "hxlc_wmps_test2")


def get_db(name: str):
    if not name:
        return None
    k = str(name).lower()
    if k in ("dev", "development"):
        return DEV_DB
    if k in ("test", "testing"):
        return TEST_DB
    return None
