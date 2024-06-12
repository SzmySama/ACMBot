
import asyncio
import nonebot
import hashlib
import random
from nonebot import logger
from .models import RaceInfo, UserInfo, UserProfileModel
import re
from lxml import etree
import requests
import time

# fmt: off
from nonebot import require
require("nonebot_plugin_htmlrender")
from nonebot_plugin_htmlrender import template_to_pic
# fmt: on


async def genCodeforcesUserProlfile(p: UserInfo, userNumber: int) -> bytes:
    from pathlib import Path

    template_path = str(Path(__file__).parent / "templates")
    template_name = "template.j2"

    templates: UserProfileModel = UserProfileModel(
        headLogoURL=p.headLogoURL, username=p.nickname, rank="", rating=p.rating
    )

    logger.debug(templates.model_dump())
    return await template_to_pic(
        template_path=template_path,
        template_name=template_name,
        templates=templates.serialize(),
        pages={
            "viewport": {"width": 400, "height": 225},
        },
        wait=10,
    )


async def fetchCodeforcesAPI(api_mothed: str, args: dict[str, str]) -> dict | None:
    config = nonebot.get_driver().config
    key = config.hycbot["codeforces"]["key"]
    secret = config.hycbot["codeforces"]["secret"]
    logger.info(key)
    logger.info(secret)
    api_url = "https://codeforces.com/api/"
    all_arguments = {
        "apiKey": key,
        "time": str(int(time.time())),
        # 其他参数
        **args,
    }

    logger.info(all_arguments)
    rand_str = str(random.randint(100_000, 1_000_000))
    hash_source = f"{rand_str}/{api_mothed}?"
    api_fullurl = f"{api_url}{api_mothed}?"

    sorted_items = sorted(all_arguments.items())
    for k, v in sorted_items:
        hash_source += f"{k}={v}&"
        api_fullurl += f"{k}={v}&"
    hash_source = hash_source[:-1] + "#"
    hash_source += secret

    hash_sig = hashlib.sha512(hash_source.encode("utf-8")).hexdigest()
    api_fullurl += f"apiSig={rand_str}{hash_sig}"

    response = requests.get(api_fullurl)

    return response.json() if response.status_code == 200 else None


async def fetchCodeforcesUserInfo(users: list[str]) -> list[UserInfo]:
    json_data = await fetchCodeforcesAPI(
        "user.info", {"handles": ";".join(
            users), "checkHistoricHandles": "false"}
    )
    output = []
    if json_data is None:
        logger.error("请求失败")
    else:
        for i in json_data["result"]:
            output.append(
                UserInfo(i["handle"], i["rating"], i["rank"], i["avatar"]))

    return output


async def fetchAtcoderRaces() -> list[RaceInfo]:
    output = []
    url = "https://atcoder.jp/contests/"

    xpath_fullpage2allRace = "//div[@id='contest-table-upcoming']/div/div/table/tbody/tr"
    xpath_start_time = ".//a/time/text()"
    xpath_race_URL = ".//a/@href"
    xpath_race_title = ".//a/text()"
    xpath_duration_time = ".//text()"

    response = requests.get(url).text
    tree = etree.HTML(response)

    all_race = tree.xpath(xpath_fullpage2allRace)
    for race in all_race:
        elements = race.xpath(".//td")
        element_start_time = elements[0]
        element_race = elements[1]
        element_duration_time = elements[2]

        start_time = element_start_time.xpath(xpath_start_time)[0]
        race_URL = element_race.xpath(xpath_race_URL)[0]
        race_title = element_race.xpath(xpath_race_title)[0]
        race_duration_time = element_duration_time.xpath(xpath_duration_time)[
            0]
        hours, minutes = map(int, race_duration_time.split(':'))
        output.append(RaceInfo(race_title, f"{url}{race_URL}", time.strptime(
            start_time, '%Y-%m-%d %H:%M:%S%z'), hours*60 + minutes))

    return output


async def fetchCodeforcesRaces() -> list[RaceInfo]:
    json_data = await fetchCodeforcesAPI("contest.list", {"gym": "false"})
    output = []
    if json_data is None:
        logger.error("请求失败")
        output.append(RaceInfo("None", "None", "None\nCodeForces拒绝了访问申请"))
    else:
        now_time = time.time()
        for i in json_data["result"]:
            if i["startTimeSeconds"] < now_time:
                break
            output.append(
                RaceInfo(
                    title=i["name"],
                    start_time=time.localtime(float(i["startTimeSeconds"])),
                    url=f"https://codeforces.com/contests/{i['id']}",
                    duration_minutes=float(i['durationSeconds'])/60,
                )
            )
    return output[::-1]


async def fetchNowcoderRaces() -> list[RaceInfo]:
    target_url = "https://ac.nowcoder.com/acm/contest/vip-index"
    response = requests.get(target_url)
    tree = etree.HTML(response.text)
    all_a = tree.xpath(
        "//div[@class='platform-mod js-current']/div[@class='platform-item js-item ']/div[@class='platform-item-main']/div[@class='platform-item-cont']"
    )
    base_url = "https://ac.nowcoder.com"

    output: list[RaceInfo] = []

    for i in all_a:
        url: str = i.xpath(".//h4/a/@href")[0]
        url = f"{base_url}{url}"

        title: str = i.xpath(".//h4/a/text()")[0]

        time_str: str = i.xpath(".//ul/li[@class='match-time-icon']/text()")[0]
        time_str.replace("\n", "")
        if not url.startswith("/dis"):
            start_time_str = re.findall(
                r"(\d{4}-\d{2}-\d{2} \d{2}:\d{2})", time_str)[0]
            keep_time_str = re.findall(r"\([^)]*(\d)[^(]*\)", time_str)[0]
            start_time = time.strptime(start_time_str, "%Y-%m-%d %H:%M")
            output.append(
                RaceInfo(
                    title=title,
                    start_time=start_time,
                    url=url,
                    duration_minutes=int(keep_time_str) * 60,
                )
            )

    return output


async def fetchTodayRaces() -> list[RaceInfo]:
    races: list[RaceInfo] = []
    tasks = [
        fetchAtcoderRaces(),
        fetchCodeforcesRaces(),
        fetchNowcoderRaces()
    ]
    current_time = time.localtime()
    for task in asyncio.as_completed(tasks):
        try:
            result = await task
            for i in result:
                if i.start_time.tm_year == current_time.tm_year and i.start_time.tm_mon == current_time.tm_mon and i.start_time.tm_mday == current_time.tm_mday:
                    races.append(i)
        except Exception as e:
            print(f"Error fetching races: {e}")

    return races
