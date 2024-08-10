import requests
import re
import urllib
import json
import time


def get_data(username: str) -> dict:
    """
    根据传入的ID查询该用户在网站Codeforces的做题情况

    参数：
    username（str）: 用户的ID

    返回：
    res（str）:按照格式返回这个用户在网站的做题记录
    """
    # 用户的所有信息
    user_data = {
        "Username": username,  # 用户名
        "RatingTimes": 0,  # 参赛次数
        "RatingNow": 0,  # 当前分
        "RatingMax": 0,  # 最高分
        "AllSubmit": 0,  # 总提交数
        "AllAccept": 0,  # 总通过数
        "TodaySubmit": 0,  # 今日提交数
        "TodayAccept": 0,  # 今日通过数
        "LatestSubmitTime": 0,  # 最新的提交时间
        "LatestSubmitStatus": 0,  # 最新的提交状态
        "LatestSubmitPassed": 0,  # 最新提交的题目以前是否通过
        "LatestSubmitTimes": 0,  # 最新题目的提交次数
        "LatestSubmitLink": ""  # 最新提交题目链接
    }
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
    }
    cookies = {
        "__client_id": "1c1cc0970b0e45efb99dc74150148a1f325e0b8b",
        "_uid": "1054142"
    }

    # 计算出今天0的时间戳，中国位于东八区因此需要调整8小时的偏移量
    now = time.time() + 8 * 3600
    start_time = now - now % 86400

    # 根据用户名获取uid
    uid_url = f"https://www.luogu.com.cn/api/user/search?keyword={username}"
    uid = requests.get(uid_url, headers=headers).json()['users'][0]['uid']

    # rating统计
    rating_url = f"https://www.luogu.com.cn/user/{uid}#practice"
    rating_resp = urllib.parse.unquote(requests.get(rating_url, headers=headers).text)
    rating_resp = rating_resp.encode('utf8').decode('unicode_escape')
    pattern = re.compile(r'JSON.parse\(decodeURIComponent\("(.*)"\)\)', re.S)
    rating_result = pattern.findall(rating_resp)[0].replace("\\", "/").split(',"passedProblems"')[0] + "}}"
    rating_data = json.loads(rating_result, strict=False)
    try:
        if rating_data['currentData']['user']['eloValue']:
            user_data['RatingNow'] = rating_data['currentData']['user']['eloValue']
            user_data['RatingMax'] = rating_data['currentData']['eloMax']['rating']
    except:
        pass

    # 对所有ac的统计
    pattern = re.compile(r'JSON.parse\(decodeURIComponent\("(.*)"\)\)', re.S)
    ac_url = f"https://www.luogu.com.cn/record/list?user={uid}&status=12&page=1"  # 只看ac
    ac_resp = requests.get(ac_url, headers=headers, cookies=cookies).text
    ac_resp = urllib.parse.unquote(pattern.findall(ac_resp)[0]).encode("utf8").decode("raw_unicode_escape")  # 编码解析
    ac_data = json.loads(ac_resp)
    user_data['AllAccept'] = ac_data["currentData"]["records"]["count"]

    # 今日所有状态的做题统计
    page = todayflag = 0
    while 1:
        page += 1
        all_url = f"https://www.luogu.com.cn/record/list?user={uid}&status=&page={page}"
        pattern = re.compile(r'JSON.parse\(decodeURIComponent\("(.*)"\)\)', re.S)
        all_resp = requests.get(all_url, headers=headers, cookies=cookies).text
        all_resp = urllib.parse.unquote(pattern.findall(all_resp)[0]).encode("utf8").decode(
            "raw_unicode_escape")  # 编码解析
        all_data = json.loads(all_resp)
        user_data['AllSubmit'] = all_data["currentData"]["records"]["count"]  # 所有做题的统计
        for i in all_data["currentData"]["records"]["result"]:
            if not user_data['LatestSubmitLink']:  # 最新的题目
                pid = i['problem']['pid']
                user_data['LatestSubmitLink'] = f"https://www.luogu.com.cn/problem/{pid}"
                user_data['LatestSubmitTime'] = int(time.time() - i["submitTime"])
                user_data['LatestSubmitStatus'] = int(i["status"] == 12)
            if i["submitTime"] >= start_time:  # 今日做题
                user_data['TodaySubmit'] += 1
                if i["status"] == 12:
                    user_data['TodayAccept'] += 1
            else:
                todayflag = 1
        if todayflag:
            break

    # 最新提交的做题次数
    lastest_try_url = f"https://www.luogu.com.cn/record/list?user={uid}&pid={pid}&page=1"
    lastest_try_resp = requests.get(lastest_try_url, headers=headers, cookies=cookies).text
    lastest_try_resp = urllib.parse.unquote(pattern.findall(lastest_try_resp)[0]).encode("utf8").decode(
        "raw_unicode_escape")  # 编码解析
    lastest_try_data = json.loads(lastest_try_resp)
    user_data["LatestSubmitTimes"] = lastest_try_data["currentData"]["records"]["count"]

    # 最新提交的ac次数
    lastest_ac_url = f"https://www.luogu.com.cn/record/list?user={uid}&pid={pid}&status=12&page=1"
    lastest_ac_resp = requests.get(lastest_ac_url, headers=headers, cookies=cookies).text
    lastest_ac_resp = urllib.parse.unquote(pattern.findall(lastest_ac_resp)[0]).encode("utf8").decode(
        "raw_unicode_escape")
    lastest_ac_data = json.loads(lastest_ac_resp)
    user_data["LatestSubmitPassed"] = lastest_ac_data["currentData"]["records"]["count"]

    # 将所有信息返回
    return user_data
