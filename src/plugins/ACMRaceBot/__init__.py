from nonebot import get_plugin_config, logger, on_command
from nonebot.adapters import Message
from nonebot.adapters.onebot.v11 import MessageSegment
from nonebot.params import CommandArg
from nonebot.plugin import PluginMetadata

from .API import (
    fetchAtcoderRaces,
    fetchCodeforcesRaces,
    fetchCodeforcesUserInfo,
    fetchNowcoderRaces,
    fetchTodayRaces,
    fetchAllRaces,
    genCodeforcesUserProlfile,
)
from .config import Config
from .models import RaceInfo

__plugin_meta__ = PluginMetadata(
    name="hyc_race",
    description="",
    usage="",
    config=Config,
)

config = get_plugin_config(Config)

logger.debug(config)


def gen_message(data: list[RaceInfo]) -> str:
    output = ""
    for i in data:
        minutes = i.duration_minutes % 60
        output += f"{i.title}\n"
        output += f"开始时间：{i.start_time.to(
            "Asia/Shanghai").format('YYYY-MM-DD HH:mm:ss')}\n"
        output += f"比赛时长：{int(i.duration_minutes / 60)
                          }小时{f"{minutes}分钟" if minutes > 0 else ""}\n"
        output += f"传送门->：{i.url}\n\n"

    return output if data else "没有获取到数据哦"


AtCoderRaceHandler = on_command("近期at")


@AtCoderRaceHandler.handle()
async def AtCoderRaceHandleFunciton():
    await AtCoderRaceHandler.finish(
        "近期AtCoder：\n" + gen_message(await fetchAtcoderRaces())
    )


CodeforcesRaceHandler = on_command("近期cf")


@CodeforcesRaceHandler.handle()
async def CodeforcesRaceHandleFunction():
    await CodeforcesRaceHandler.finish(
        "近期CodeForces：\n" + gen_message(await fetchCodeforcesRaces())
    )


NowcoderRaceHandler = on_command("近期nk")


@NowcoderRaceHandler.handle()
async def NowcoderRaceHandleFunction():
    await NowcoderRaceHandler.finish(
        "近期牛客：\n" + gen_message(await fetchNowcoderRaces())
    )


CodeforcesUserInfoHandler = on_command("cf")


@CodeforcesUserInfoHandler.handle()
async def CodeforcesUserInfohandleFunction(args: Message = CommandArg()):
    if username := args.extract_plain_text():
        users = await fetchCodeforcesUserInfo([username])
        if users:
            pic = await genCodeforcesUserProlfile(users[0])
            await CodeforcesUserInfoHandler.finish(MessageSegment.image(pic))
        else:
            await CodeforcesUserInfoHandler.finish("没有找到这个用户，是不是打错名字了？")
    else:
        await CodeforcesUserInfoHandler.finish("请输入用户名")


TodayRaceHandler = on_command("今日比赛")


@TodayRaceHandler.handle()
async def TodyRaceHandleFunction():
    await NowcoderRaceHandler.finish(
        "今日比赛：\n" + gen_message(await fetchTodayRaces())
    )

AllRaceHandler = on_command("近期比赛")


@AllRaceHandler.handle()
async def AllRaceHandlFunction():
    await NowcoderRaceHandler.finish(
        "近期比赛：\n" + gen_message(await fetchAllRaces())
    )
