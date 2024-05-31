class RaceInfo:
    def __init__(self, title: str, start_time: str, url: str) -> None:
        """
        比赛信息类，用于存储比赛的标题、开始时间和相关链接。

        :param title: 比赛标题
        :param start_time: 比赛开始时间
        :param url: 比赛信息的网址
        """
        self.title = title
        self.start_time = start_time
        self.url = url
