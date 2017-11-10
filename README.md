# ServerMonitor

### 本服务监测包含以下：
- HTTP — Nginx、Tomcat等
- TCP — MQ服务等TCP协议连接
- Redis
- MySQL


### 配置文件:
```yaml
# config
enabled: true
instances:
  http:
    - name: Nginx
      url: http://192.168.1.100:80
  mysql:
    - name: MySQL
      host: 192.168.10.100
      port: 3306
      user: root
      pass: pass
  redis:
    - name: Redis
      host: 192.168.10.100
      port: 6379
  tcp:
    - name: ActiveMQ
      host: 192.168.10.102
      port: 61616
# 钉钉机器人token      
ddRobotToken: 027956b4093ae5194ceb180ca549eaa1fec45b5b8915f0753b851a9691af3649
```


### 编译原型：
可使用gox编译
参考:
```shell
## 编译linux
gox -osarch "linux"
## 编译linux及windows
gox -osarch "linux windows" 
```


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

### 下载:
[monitor_linux_386](http://oz6t8di9l.bkt.clouddn.com/monitor_linux_386)<br>
[monitor_linux_amd64](http://oz6t8di9l.bkt.clouddn.com/monitor_linux_amd64)<br>
[monitor_windows_386](http://oz6t8di9l.bkt.clouddn.com/monitor_windows_386.exe)<br>
[monitor_windows_amd64](http://oz6t8di9l.bkt.clouddn.com/monitor_windows_amd64.exe)
