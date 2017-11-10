// server_monitor
package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
)

var (
	conf                 Conf
	msgs                 []Message
	validation_sql_mysql = "select 1"
	dingdingBaseServer   = "https://oapi.dingtalk.com/robot/send?access_token="
	dingdingMsgTemplet   = "{\"msgtype\":\"text\",\"text\":{\"content\":\"%s\"}}"
)

//配置
type Conf struct {
	Enabled   bool `yaml:"enabled"`
	Instances struct {
		Http []struct {
			Name         string `yaml:"name"`
			Url          string `yaml:"url"`
			Username     string `yaml:"username"`
			Password     string `yaml:"password"`
			ContentMatch string `yaml:"content_match"`
			StatusCode   int    `yaml:"status_code"`
		} `yaml:"http"`
		Mysql []struct {
			Name string `yaml:"name"`
			Host string `yaml:"host"`
			User string `yaml:"user"`
			Pass string `yaml:"pass"`
			Port string `yaml:"port"`
		} `yaml:"mysql"`
		Redis []struct {
			Name string `yaml:"name"`
			Host string `yaml:"host"`
			Pass string `yaml:"pass"`
			Port string `yaml:"port"`
		} `yaml:"redis"`
		TCP []struct {
			Name string `yaml:"name"`
			Host string `yaml:"host"`
			Port string `yaml:"port"`
		} `yaml:"tcp"`
	} `yaml:"instances"`
	DdRobotToken string `yaml:"ddRobotToken"`
}

//消息
type Message struct {
	title   string
	content string
}

func main() {
	initLogFileWriter()
	conf.initConf()
	if !conf.Enabled {
		return
	}
	checkHttpServer()
	checkMySqlServer()
	checkRedisServer()
	checkTCPServer()
	sendMsgToDingDing()
}

//检查Http
func checkHttpServer() {
	if len(conf.Instances.Http) != 0 {
		for _, httpc := range conf.Instances.Http {
			resp, err := http.Get(httpc.Url)
			if err != nil {
				log.Errorf("Get data error", err)
				appendToMsg("HTTP -> "+httpc.Name+"【"+httpc.Url+"】", "请求异常")
				continue
			}
			defer resp.Body.Close()
			if httpc.StatusCode == 0 {
				httpc.StatusCode = 200
			}
			if httpc.StatusCode != resp.StatusCode {
				log.Errorf("HTTP StatusCode error", err)
				appendToMsg("HTTP -> "+httpc.Name+"【"+httpc.Url+"】", err.Error())
				continue
			}
			if httpc.ContentMatch != "" {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Error("Read response error ", err)
					appendToMsg("HTTP -> "+httpc.Name+"【"+httpc.Url+"】", err.Error())
				}
				log.Info("HTTP -> ", resp)
				match, err := regexp.MatchString(httpc.ContentMatch, string(body))
				if err != nil {
					log.Errorf("HTTP content_match error", err)
					appendToMsg("HTTP -> "+httpc.Name+"【"+httpc.Url+"】", err.Error())
					continue
				} else if !match {
					log.Errorf("HTTP response check mismatching")
					appendToMsg("HTTP -> "+httpc.Name+"【"+httpc.Url+"】", "Check response mismatching")
					continue
				}
			}
			log.Info("HTTP -> "+httpc.Name+"【"+httpc.Url+"】", "test success")
		}
	}
}

//检查Mysql
func checkMySqlServer() {
	if len(conf.Instances.Mysql) != 0 {
		for _, mysql := range conf.Instances.Mysql {
			dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8", mysql.User, mysql.Pass, mysql.Host, mysql.Port)
			db, err := sql.Open("mysql", dataSource)
			if err != nil {
				log.Errorf("DB connect error", err)
				appendToMsg("MySQL -> "+mysql.Name+"【"+mysql.Host+":"+mysql.Port+"】", "连接异常")
				continue
			}
			rows, err := db.Query(validation_sql_mysql)
			if err != nil {
				log.Errorf("DB validate error", err)
				appendToMsg("MySQL -> "+mysql.Name+"【"+mysql.Host+":"+mysql.Port+"】", "查询测试失败，请检查服务")
				continue
			}
			defer rows.Close()
			log.Info("MySQL -> "+mysql.Name+"【"+mysql.Host+":"+mysql.Port+"】", "is running")
		}
	}
}

//检查Redis
func checkRedisServer() {
	if len(conf.Instances.Redis) != 0 {
		for _, redisdb := range conf.Instances.Redis {
			conn, err := redis.Dial("tcp", redisdb.Host+":"+redisdb.Port)
			if err != nil {
				log.Errorf("connect redis error", err)
				appendToMsg("Redis -> "+redisdb.Name+"【"+redisdb.Host+":"+redisdb.Port+"】", "连接异常")
				continue
			}
			if redisdb.Pass != "" {
				if _, err = conn.Do("AUTH", redisdb.Pass); err != nil {
					log.Errorf("Redis AUTH error", err)
					appendToMsg("Redis -> "+redisdb.Name+"【"+redisdb.Host+":"+redisdb.Port+"】", err.Error())
					continue
				}
			}
			if _, err = conn.Do("SET", "GO_TEST_KEY", 123456); err != nil {
				log.Errorf("Test Redis GET error", err)
				appendToMsg("Redis -> "+redisdb.Name+"【"+redisdb.Host+":"+redisdb.Port+"】", err.Error())
				continue
			}
			log.Info("Redis -> "+redisdb.Name+"【"+redisdb.Host+":"+redisdb.Port+"】", "is running")
			defer conn.Close()
		}
	}
}

//检查TCP
func checkTCPServer() {
	if len(conf.Instances.TCP) != 0 {
		for _, tcp := range conf.Instances.TCP {
			_, err := net.Dial("tcp", net.JoinHostPort(tcp.Host, tcp.Port))
			if err != nil {
				//tcp test
				log.Errorf("Connect error", err)
				appendToMsg("TCP -> "+tcp.Name+"【"+tcp.Host+":"+tcp.Port+"】", "连接异常")
				continue
			}
			log.Info("TCP -> "+tcp.Name+"【"+tcp.Host+":"+tcp.Port+"】", "connect success")

		}
	}
}

//发送消息到钉钉
func sendMsgToDingDing() {
	var content = ""
	for _, msg := range msgs {
		content += msg.title + "\n" + msg.content + "\n"
	}
	httpPost(dingdingBaseServer+conf.DdRobotToken, fmt.Sprintf(dingdingMsgTemplet, content))
	msgs = []Message{}
}

//POST及处理响应
func httpPost(url string, msg string) {
	resp, err := http.Post(url, "application/json", strings.NewReader(msg))
	if err != nil {
		log.Error("Post data error ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Read response error ", err)
	}
	log.Info("POST -> ", resp)
	result := string(body)
	log.Info("Response data ", result)
}

//消息
func appendToMsg(title string, content string) {
	var m Message
	m.title = title
	m.content = content
	msgs = append(msgs, m)
	log.Info(msgs)
}

//初始化配置
func (conf *Conf) initConf() *Conf {
	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Errorf("yamlFile.Get err #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		log.Errorf("Unmarshal: %v", err)
	}
	return conf
}

//初始化日志配置
func initLogFileWriter() {
	logConfig := `
		<seelog>
		    <outputs formatid="main">   
				<console/>
		        <buffered size="10" flushperiod="10">
					<file path="./out.log" />
				</buffered>
		    </outputs>
		    <formats>
		        <format id="main" format="%Date %Time [%LEV] %Msg%n"/>
		    </formats>
		</seelog>
	`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(logConfig))
	log.ReplaceLogger(logger)
}
