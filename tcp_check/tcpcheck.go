// activemq
package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
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
	checkTCPServer()
	sendMsgToDingDing()
}

//检查TCP
func checkTCPServer() {
	if len(conf.Instances.TCP) != 0 {
		for _, tcp := range conf.Instances.TCP {
			_, err := net.Dial("tcp", net.JoinHostPort(tcp.Host, tcp.Port))
			if err != nil {
				//tcp test
				log.Errorf("Connect error", err)
				appendToMsg("TCP -> "+tcp.Name+"【"+tcp.Host+":"+tcp.Port+"】", err.Error())
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
	yamlFile, err := ioutil.ReadFile("tcpcheck-config.yml")
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
