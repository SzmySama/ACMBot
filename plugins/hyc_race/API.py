import requests
from lxml import etree
import pprint

def fetch_atcoder_contests() -> list[dict]:
    """
    Fetches the list of AtCoder contests from the AtCoder website.

    This function sends an HTTP GET request to the AtCoder website and then
    parses the HTML response to extract the contests. It returns a dictionary
    where the keys are the contest infomations (as strings) and the values are the
    corresponding URLs (also as strings).

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
            for data in all_data:
                lines = data.split('\n')
                # pprint.pprint(lines)
                for line in lines:
                    words = line.split(':')
                    if len(words) > 1:
                        now_race_map[words[0].replace("- ","")] = ":".join(words[1:])
            pprint.pprint(now_race_map)
            print(now_race_map.keys())
            output.append(now_race_map)

        # xpath = "//div[@class='col-sm-8']/div[@class='panel panel-default']"
        # links: list[str] = tree.xpath(xpath)
        

    except requests.RequestException as e:
        print(f"Error fetching data: {e}")
    return output

if __name__ == "__main__":
    fetch_atcoder_contests()
