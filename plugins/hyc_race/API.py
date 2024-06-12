
import asyncio
from datetime import datetime
import nonebot
import hashlib
import random
from nonebot import logger
from .models import RaceInfo, UserInfo, UserProfileModel
import re
import pprint
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
    url = "https://atcoder.jp/"
    output = []
    response = requests.get(url)
    response.raise_for_status()
    response.encoding = response.apparent_encoding
    html_content = response.text

    xpath_contests = "//div[@class='col-sm-8']/div[@class='panel panel-default']"
    xpath_title = ".//h3[@class='panel-title']/a/text()"
    xpath_message = ".//div[@class='panel-body blog-post']/text()"
    regex_race_time = r"iso=(\d{4}\d{2}\d{2}T\d{4})"
    regex_race_title = r"<a[^>]*>([^<]*)</a>"
    regex_race_url = r'href="([^"]*)"'
    regex_keep_time = r'- Duration: *(\d+).*\n'

    tree = etree.HTML(html_content)
    contests = tree.xpath(xpath_contests)
    for contest in contests[::-1]:
        title = contest.xpath(xpath_title)[0]
        all_data = contest.xpath(xpath_message)
        for data in all_data:
            str_data = "".join(data)
            race_time_origin = re.findall(regex_race_time, str_data)[0]
            race_time_obj = time.strptime(race_time_origin, "%Y%m%dT%H%M")
            race_timestamp = time.mktime(race_time_obj)
            race_timestamp += 3600  # 原来是+7时区的，现在转换成+8时区
            race_time = time.localtime(race_timestamp)
            race_title: str = re.findall(regex_race_title, str_data)[0]
            race_url: str = re.findall(regex_race_url, str_data)[0]
            race_keep_time: int = int(re.findall(regex_keep_time, str_data)[0])
            output.append(RaceInfo(title=race_title, url=race_url,
                          start_time=race_time, duration_hours=race_keep_time/60))
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
                    duration_hours=float(i['durationSeconds'])/3600,
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
                    duration_hours=float(keep_time_str),
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
