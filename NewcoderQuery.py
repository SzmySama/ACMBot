import requests
import time
import re


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

    # 计算出今天0的时间戳，中国位于东八区因此需要调整8小时的偏移量
    now = time.time() + 8 * 3600
    start_time = now - now % 86400

    # 根据用户名获取uid
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0",
    }
    uid_url = f"https://ac.nowcoder.com/acm/contest/rating-index?searchUserName={username}"
    uid_resp = requests.get(uid_url, headers=headers).text
    uid = re.compile('tr data-isFollowedByHost=" 0 " data-uid="(.*)">').findall(uid_resp)[0]

    # 练习页url，一页两百个，一天的做题数量理应不超过两百
    practice_url = f"https://ac.nowcoder.com/acm/contest/profile/{uid}/practice-coding?&pageSize=200"
    rating_url = f"https://ac.nowcoder.com/acm/contest/rating-history?token=&uid={uid}"  # 比赛记录页面url

    # rating统计
    rating_resp = requests.get(rating_url, headers=headers).json()
    user_data['RatingNow'] = int(rating_resp['data'][-1]['rating'])
    try:  # 没参加过比赛会获取不到数据，用异常退出
        for i in rating_resp['data']:
            user_data['RatingMax'] = int(max(user_data['RatingMax'], i['rating']))
            user_data['RatingTimes'] += 1
    except:
        pass

    # 今日做题统计
    practice_resp = requests.get(practice_url, headers=headers, verify=False).text
    practice_data = re.compile('<div class="state-num">(.*?)</div>', re.S).findall(practice_resp)
    user_data["AllSubmit"] = practice_data[2]
    user_data['AllAccept'] = practice_data[1]
    practice_resp = re.compile(
        '<td><a href="/acm/problem/(.*?)" target="_blank">.*?<td><span class="match-score.">(.*?)</span></td>(.*?)</tr>',
        re.S).findall(practice_resp)
    for i in practice_resp:
        score = i[1]
        submit_time = " ".join(re.compile('<td>(.*)</td>').findall(i[2])[-1].split())
        if time.mktime(time.strptime(submit_time, "%Y-%m-%d %H:%M:%S")) >= start_time:  # 今天的提交
            user_data['TodaySubmit'] += 1
            if score == "100":
                user_data['TodayAccept'] += 1

    # 最新的题目
    problem_id = practice_resp[0][0]
    user_data['LatestSubmitLink'] = f"https://ac.nowcoder.com/acm/problem/{problem_id}"
    problem_url = f"https://ac.nowcoder.com/acm/contest/profile/{uid}/practice-coding?pageSize=200&search={problem_id}"
    problem_resp = requests.get(problem_url, headers=headers).text
    problem_resp = re.compile('<span class="match-score.">(.*?)</span></td>(.*?)</tr>', re.S).findall(problem_resp)
    for i in problem_resp:
        score = i[0]
        user_data['LatestSubmitTimes'] += 1
        if score == '100':
            user_data['LatestSubmitPassed'] += 1
    user_data['LatestSubmitStatus'] = int(practice_resp[0][1] == '100')
    latest_time = re.compile('<td>(.*)</td>').findall(problem_resp[0][1])[-1]
    user_data['LatestSubmitTime'] = int(time.time() - time.mktime(time.strptime(latest_time, "%Y-%m-%d %H:%M:%S")))

    # 将所有信息返回
    return user_data
