from nonebot import get_plugin_config
from nonebot.plugin import PluginMetadata

from .config import Config
from .API import *

__plugin_meta__ = PluginMetadata(
    name="hyc_race",
    description="",
    usage="",
    config=Config,
)

config = get_plugin_config(Config)

