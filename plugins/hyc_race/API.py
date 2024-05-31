import time
import requests
from lxml import etree
import pprint
import re
from .RaceInfo import RaceInfo
from nonebot import logger
import random
import hashlib
import nonebot
from datetime import datetime

def fetchAtcoderRaces() -> list[RaceInfo]:
    """
    Fetches the list of AtCoder contests from the AtCoder website.

    This function sends an HTTP GET request to the AtCoder website and then
    parses the HTML response to extract the contests. It returns a dictionary
    keep the contest infomations

    Return:
    list[dict]
    the dict_keys is['Title', 'Contest URL', 'Start Time', 'Duration', 'Number of Tasks', 'Writer', 'Tester', 'Rated range']
    """
    url = 'https://atcoder.jp/'
    output = []
    try:
        response = requests.get(url)
        response.raise_for_status()
        response.encoding = response.apparent_encoding
        html_content = response.text

        tree = etree.HTML(html_content)
        xpath_contests = "//div[@class='col-sm-8']/div[@class='panel panel-default']"
        contests = tree.xpath(xpath_contests)
        for contest in contests:
            now_race_map = {}
            xpath_title = ".//h3[@class='panel-title']/a/text()"
            xpath_message = ".//div[@class='panel-body blog-post']/text()"
            title = contest.xpath(xpath_title)[0]
            now_race_map['Title'] = title
            all_data = contest.xpath(xpath_message)
            regex_race_time = r"iso=(\d{4}\d{2}\d{2}T\d{4})"
            regex_race_title = r'<a[^>]*>([^<]*)</a>'
            regex_race_url = r'href="([^"]*)"'
            for data in all_data:
                str_data = ''.join(data)
                race_time = re.findall(regex_race_time,str_data)[0]
                race_title = re.findall(regex_race_title,str_data)[0]
                race_url = re.findall(regex_race_url,str_data)[0]
                output.append(RaceInfo(race_title,race_time,race_url))

    except requests.RequestException as e:
        print(f"Error fetching data: {e}")

    return output

def fetchCodeforcesRaces() -> list[RaceInfo]:
    config = nonebot.get_driver().config
    key= config.hycbot['codeforces']['key']
    secret= config.hycbot['codeforces']['secret']
    logger.info(key)
    logger.info(secret)
    api_url = "https://codeforces.com/api/"
    api_mothed = "contest.list"
    all_arguments = {
        "apiKey": key,
        "time": str(int(time.time())),
        # 其他参数
        "gym": "false",
    }

    rand_str = str(random.randint(100_000,1_000_999))
    hash_source = f"{rand_str}/{api_mothed}?"
    api_fullurl = f"{api_url}{api_mothed}?"

    sorted_items = sorted(all_arguments.items())
    for k,v in sorted_items:
        hash_source+=f"{k}={v}&"
        api_fullurl+=f"{k}={v}&"
    hash_source = hash_source[:-1] + "#"
    hash_source += secret

    hash_sig = hashlib.sha512(hash_source.encode('utf-8')).hexdigest()
    api_fullurl+=f"apiSig={rand_str}{hash_sig}"

    response = requests.get(api_fullurl)
    output = []
    if (response.status_code != 200):
        logger.error("请求失败")
        output.append(RaceInfo("None","None","None\nCodeForces拒绝了访问申请"))
    else:
        logger.info(response)
        json_data = response.json()
        for i in json_data['result'][:5]:
            output.append(RaceInfo(i['name'],datetime.fromtimestamp(int(i['startTimeSeconds'])),f"https://codeforces.com/contests/{i['id']}"))
    return output