// mysql
package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/cihub/seelog"
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
		Mysql []struct {
			Name string `yaml:"name"`
			Host string `yaml:"host"`
			User string `yaml:"user"`
			Pass string `yaml:"pass"`
			Port string `yaml:"port"`
		} `yaml:"mysql"`
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
	checkMySqlServer()
	sendMsgToDingDing()
}

//检查Mysql
func checkMySqlServer() {
	if len(conf.Instances.Mysql) != 0 {
		for _, mysql := range conf.Instances.Mysql {
			dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8", mysql.User, mysql.Pass, mysql.Host, mysql.Port)
			db, err := sql.Open("mysql", dataSource)
			if err != nil {
				log.Errorf("DB connect error", err)
				appendToMsg("MySQL -> "+mysql.Name+"【"+mysql.Host+":"+mysql.Port+"】", err.Error())
				continue
			}
			rows, err := db.Query(validation_sql_mysql)
			defer rows.Close()
			if err != nil {
				log.Errorf("DB validate error", err)
				appendToMsg("MySQL -> "+mysql.Name+"【"+mysql.Host+":"+mysql.Port+"】", err.Error())
				continue
			}
			log.Info("MySQL -> "+mysql.Name+"【"+mysql.Host+":"+mysql.Port+"】", "is running")
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
	yamlFile, err := ioutil.ReadFile("mysql-config.yml")
	if err != nil {
		log.Errorf("yamlFile.Get err   #%v ", err)
		if len(yamlFile) == 0 {
			yamlFile, err = ioutil.ReadFile("config.yml")
			if err != nil {
				log.Errorf("yamlFile.Get err   #%v ", err)
			}
		}
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
		        <buffered size="1" flushperiod="1">
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
