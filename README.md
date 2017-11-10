# ServerMonitor

### 本服务监测包含以下：

- 远程桌面服务 — TCP
- Nginx服务 — HTTP
- Web服务 — HTTP
- Redis服务 — Redis
- MySQL服务 — MySQL
- MQ服务 — TCP

可使用gox交叉编译

### 参考项目：
```xml
github.com/mitchellh/gox
github.com/cihub/seelog
gopkg.in/yaml.v2
github.com/go-sql-driver/mysql
github.com/garyburd/redigo/redis
```

### 备注:
配置文件使用yaml，支持多服务监听<br>
各模块使用单独配置文件x-config.yml，也可以统一使用config.yml配置<br>
信息推送使用钉钉自定义机器人
