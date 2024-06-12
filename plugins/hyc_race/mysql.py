import pymysql

class MysqlTool:
    def __init__(self):
        """mysql 连接初始化"""
        self.host = '192.168.100.28'
        self.port = 3306
        self.user = 'root'
        self.password = 'hycakworldfinal'
        self.db = 'QQbot'
        self.charset = 'utf8'
        self.mysql_conn = None

    def __enter__(self):
        """打开数据库连接"""
        self.mysql_conn = pymysql.connect(
            host=self.host,
            port=self.port,
            user=self.user,
            passwd=self.password,
            db=self.db,
            charset=self.charset
        )
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """关闭数据库连接"""
        if self.mysql_conn:
            self.mysql_conn.close()
            self.mysql_conn = None

    def execute(self, sql: str, args: tuple = None, commit: bool = False) -> any:
        """执行 SQL 语句"""
        try:
            with self.mysql_conn.cursor() as cursor:
                cursor.execute(sql, args)
                if commit:
                    self.mysql_conn.commit()
                    print(f"执行 SQL 语句：{sql}，参数：{args}，数据提交成功")
                else:
                    result = cursor.fetchall()
                    print(f"执行 SQL 语句：{sql}，参数：{args}，查询到的数据为：{result}")
                    return result
        except Exception as e:
            print(f"执行 SQL 语句出错：{e}")
            self.mysql_conn.rollback()
            raise e

    def insert(self, qq_id, cf_id, headurl, cfrank):
        """增加数据为qq号，cf号，头像，cf分"""
        sql = "INSERT INTO QQbot (qq_id, cf_id, headurl, cfrank) VALUES (%s, %s, %s, %s)"
        return self.execute(sql, (qq_id, cf_id, headurl, cfrank), commit=True)

    def qqread(self, id):
        """以qq号查询数据"""
        sql = "select * from QQbot where qq_id=%s"
        return self.execute(sql, (id,), commit=False)
    
    def cfread(self, id):
        """以cf号查询数据"""
        sql = "select * from QQbot where cf_id = %s"
        return self.execute(sql, (id,), commit=False)
    
    def readall(self):
        """查询所有数据以cf号排序"""
        sql = "select * from QQbot order by cfrank"
        return self.execute(sql, commit=False)

    def cfupdate(self, cf, qq_id, cf_id, headurl, cfrank):
        """更新cf号为cf的数据为qq_id,cf_id,headurl,cfrank"""
        sql = "UPDATE QQbot SET qq_id = %s, cf_id=%s, headurl=%s, cfrank=%s WHERE cf_id = %s"
        return self.execute(sql, (qq_id, cf_id, headurl, cfrank, cf), commit=True)
    
    def qqupdate(self, qq, qq_id, cf_id, headurl, cfrank):
        """更新qq号为qq的数据为qq_id,cf_id,headurl,cfrank"""
        sql = "UPDATE QQbot SET qq_id = %s, cf_id=%s, headurl=%s, cfrank=%s WHERE qq_id = %s"
        return self.execute(sql, (qq_id, cf_id, headurl, cfrank, qq), commit=True)

    def qqdelete(self, qq):
        """删除qq号为qq的数据"""
        sql = "DELETE FROM QQbot WHERE qq_id = %s"
        return self.execute(sql, (qq,), commit=True)
    
    def cfdelete(self, cf):
        """删除cf号为cf的数据"""
        sql = "DELETE FROM QQbot WHERE cf_id = %s"
        return self.execute(sql, (cf,), commit=True)
    
    def create_table(self):
        """创建表"""
        sql = "CREATE TABLE IF NOT EXISTS QQbot (QQ_ID INT PRIMARY KEY, CF_ID VARCHAR(50), HEADURL VARCHAR(50),CFRANK INT)"
        return self.execute(sql, commit=True)