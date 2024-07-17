from pydantic import BaseModel


class Config(BaseModel):
    acm_cf_key: str = ''
    acm_cf_secret: str = ''
    acm_db_host: str = "127.0.0.1"
    acm_db_port: int = 3306
    acm_db_username: str = "root"
    acm_db_passwd: str = ""
    acm_db_dbname: str = "ACMRaceBot"
