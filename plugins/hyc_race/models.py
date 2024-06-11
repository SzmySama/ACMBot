from time import struct_time
from pydantic import BaseModel


class RaceInfo:
    def __init__(self, title: str, url: str, start_time: struct_time, duration_hours: float):
        self.title = title
        self.url = url
        self.start_time = start_time
        self.duration_hours = duration_hours


class UserInfo:
    nickname: str
    rating: int
    rank: int
    headLogoURL: str

    def __init__(self, nickname: str, rating: int, rank: int, headLogoURL: str) -> None:
        self.nickname = nickname
        self.rating = rating
        self.rank = rank
        self.headLogoURL = headLogoURL


class UserProfileModel(BaseModel):
    headLogoURL: str
    username: str
    rank: str
    rating: int

    @property
    def level(self):
        rating_levels = {
            "newbie": range(0, 1200),
            "pupil": range(1200, 1400),
            "specialist": range(1400, 1600),
            "expert": range(1600, 1900),
            "candidate-master": range(1900, 2100),
            "master": range(2100, 2300),
            "international-master": range(2300, 2400),
            "grandmaster": range(2400, 2600),
            "international-grandmaster": range(2600, 3000),
            "legendary-grandmaster": range(3000, 9999),
        }
        for level, rating_range in rating_levels.items():
            if self.rating in rating_range:
                return level

    def serialize(self):
        model_dict = self.model_dump()
        model_dict["level"] = self.level
        return model_dict
