package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	LogLevel       string  `yaml:"logLevel"`
	LoopDurationSecond   int64     `yaml:"loopDurationSecond"`
	LoseRatioLimit float64 `yaml:"loseRatioLimit"`
	Servers        []string
}

var config Config
var count chan int
var expectConnCount int

func init() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Config:\n%v\n\n", config)

	count = make(chan int, 100000000)
	expectConnCount = 0
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(level)
}

func newConn(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logrus.Error(addr, "conn ", err)
		return
	}
	count <- 0
	sig := make(chan bool)
	go func() {
		for {
			select {
			case <-sig:
				<-count
				return
			default:
				fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: "+addr+"\r\nConnection: keep-alive\r\n\r\n")
				time.Sleep(10 * time.Second)
			}
		}
	}()
	go func() {
		for {
			line, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				logrus.Error(addr, "read over ", err)
				sig <- true
				break
			}
			logrus.Debug(addr, "response line:", line)
		}
	}()

}

func main() {
	// start tcp connection loop
	for _, v := range config.Servers {
		go func(host string) {
			for {

				go newConn(host + ":80")
				time.Sleep(time.Duration(config.LoopDurationSecond * int64(time.Second)))
			}
		}(v)
	}

	// estimate the limit
	go func() {
		for {
			time.Sleep(time.Duration(config.LoopDurationSecond * int64(time.Second)))
			expectConnCount = expectConnCount + len(config.Servers)
			actualConnCount := len(count)
			loseRatio := float64(expectConnCount-actualConnCount) / float64(expectConnCount)
			if loseRatio >= config.LoseRatioLimit {
				panic(fmt.Sprintf("\nestimate tcp connection limit is %d\n", actualConnCount))
			}
			logrus.Info("Actual Conn: ", actualConnCount, "\tExpect Conn: ", expectConnCount, "\tLose ratio: ", loseRatio)
		}
	}()

	// wait tcp connection loop
	finished := make(chan bool)
	<-finished
}
