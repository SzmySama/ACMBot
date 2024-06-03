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

def gen_message(data:list[RaceInfo]) -> str:
    output = ""
    for i in data:
        output += f"比赛名称：{i.title}\n"
        output += f"开始时间：{i.start_time}\n"
        output += f"Link🌈：{i.url}\n\n"
    
    return output


AtCoderHandler = on_command("近期at")
@AtCoderHandler.handle()
async def AtCoderHandleFunciton():
    await AtCoderHandler.finish(gen_message(fetchAtcoderRaces()))

CodeforcesHandler = on_command("近期cf")
@CodeforcesHandler.handle()
async def CodeforcesHandleFunction():
    await CodeforcesHandler.finish(gen_message(fetchCodeforcesRaces()))

NowcoderHandler = on_command("近期nk")
@NowcoderHandler.handle()
async def NowcoderHandleFunction():
    await NowcoderHandler.finish(gen_message(fetchNowcoderRaces()))