// http_check
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	log "github.com/cihub/seelog"
	"gopkg.in/yaml.v2"
)

var (
	conf               Conf
	msgs               []Message
	dingdingBaseServer = "https://oapi.dingtalk.com/robot/send?access_token="
	dingdingMsgTemplet = "{\"msgtype\":\"text\",\"text\":{\"content\":\"%s\"}}"
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
	sendMsgToDingDing()
}

//检查Http
func checkHttpServer() {
	if len(conf.Instances.Http) != 0 {
		for _, httpc := range conf.Instances.Http {
			resp, err := http.Get(httpc.Url)
			if err != nil {
				log.Error("Get data error ", err)
				appendToMsg("HTTP -> "+httpc.Name+"【"+httpc.Url+"】", err.Error())
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
	yamlFile, err := ioutil.ReadFile("httpcheck-config.yml")
	if err != nil {
		log.Errorf("yamlFile.Get err   #%v ", err)
	} else if len(yamlFile) == 0 {
		yamlFile, err = ioutil.ReadFile("config.yml")
		if err != nil {
			log.Errorf("yamlFile.Get err   #%v ", err)
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
