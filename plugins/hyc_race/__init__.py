from nonebot import get_plugin_config
from nonebot.plugin import PluginMetadata
from nonebot import on_command
from .RaceInfo import RaceInfo

weather = on_command("天气")

from .config import Config
from .API import *

__plugin_meta__ = PluginMetadata(
    name="hyc_race",
    description="",
    usage="",
    config=Config,
)

config = get_plugin_config(Config)

AtCoderHandler = on_command("近期at")
@AtCoderHandler.handle()
async def AtCoderHandleFunciton():
    data = fetchAtcoderRaces()
    target = ""
    for i in data:
        target += f"比赛名称：{i.title}\n"
        target += f"开始时间：{i.start_time}\n"
        target += f"Link🌈：{i.url}\n"
    await AtCoderHandler.finish(target)


CodeforcesHandler = on_command("近期cf")
@CodeforcesHandler.handle()
async def CodeforcesHandleFunction():
    data = fetchCodeforcesRaces()
    target = ""
    for i in data:
        target += f"比赛名称：{i.title}\n"
        target += f"开始时间：{i.start_time}\n"
        target += f"Link🌈：{i.url}\n"
    await AtCoderHandler.finish(target)


NowcoderHandler = on_command("近期nk")
@NowcoderHandler.handle()
async def NowcoderHandleFunction():
    data = fetchNowcoderRaces()
    target = ""
    for i in data:
        target += f"比赛名称：{i.title}\n"
        target += f"开始时间：{i.start_time}\n"
        target += f"Link🌈：{i.url}\n"
    await AtCoderHandler.finish(target)