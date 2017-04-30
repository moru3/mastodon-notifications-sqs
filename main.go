package main

import (
	"context"
	"io/ioutil"
	"log"
	"regexp"
	"sync"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mattn/go-mastodon"
)

const settingFile = "./setting.yml"

var queueURL string

// Settings from yaml
type Settings struct {
	AwsRegion   string       `yaml:"awsRegion"`
	QueueURL    string       `yaml:"queueURL"`
	ServerConfs []ServerConf `yaml:"serverConfs"`
}

// ServerConf mastodon server's setting
type ServerConf struct {
	ServerName   string `yaml:"serverName"`
	ServerURL    string `yaml:"serverURL"`
	ClientID     string `yaml:"clientID"`
	ClientSecret string `yaml:"clientSecret"`
	Account      string `yaml:"account"`
	Password     string `yaml:"password"`
}

// notification struct
type notification struct {
	nType       string
	displayName string
	content     string
	serverName  string
}

// SQS connection
var svc *sqs.SQS

func main() {
	// load setting file
	buf, err := ioutil.ReadFile(settingFile)
	if err != nil {
		return
	}
	var s Settings
	err = yaml.Unmarshal(buf, &s)
	if err != nil {
		panic(err)
	}
	queueURL = s.QueueURL

	// SQS set up
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.AwsRegion)},
	)
	if err != nil {
		log.Fatal(err)
	}

	svc = sqs.New(sess)

	wg := &sync.WaitGroup{}
	for i := range s.ServerConfs {
		log.Println("loop", i)
		wg.Add(1)
		go connect(s.ServerConfs[i])
	}

	// no reachable
	wg.Wait()
}

func connect(conf ServerConf) {
	c := mastodon.NewClient(&mastodon.Config{
		Server:       conf.ServerURL,
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := c.Authenticate(ctx, conf.Account, conf.Password)
	if err != nil {
		log.Fatal(err)
	}

	wsc := c.NewWSClient()

	q, err := wsc.StreamingWSUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("start - ", conf.ServerName)

	cnl := make(chan bool)
	go watchStream(q, conf.ServerName, cnl)

	select {
	case <-cnl:
		log.Println("channel down and restart")
	}
	connect(conf)
}

func watchStream(q chan mastodon.Event, serverName string, c chan bool) {
	defer func() { c <- true }()
	// get event stream
	for e := range q {
		if t, ok := e.(*mastodon.NotificationEvent); ok {
			log.Println(t.Notification.Type)
			log.Println(t.Notification.Account.DisplayName)
			mentionBody := ""
			if t.Notification.Type == "mention" {
				mentionBody = removeTag(t.Notification.Status.Content)
			}
			log.Println(mentionBody)
			pushMessage(notification{
				nType:       t.Notification.Type,
				displayName: t.Notification.Account.DisplayName,
				content:     mentionBody,
				serverName:  serverName,
			})
		}
	}
}

// push a message to AWS SQS
func pushMessage(n notification) {
	pushContent := "[" + n.serverName + "]"
	switch n.nType {
	case "follow":
		pushContent += n.displayName + "さんにフォローされた"
	case "favourite":
		pushContent += n.displayName + "さんにお気に入りされた"
	case "reblog":
		pushContent += n.displayName + "さんにブーストされた"
	case "mention":
		pushContent += n.displayName + "さんから：" + n.content
	}

	params := &sqs.SendMessageInput{
		MessageBody:  aws.String(pushContent),
		QueueUrl:     aws.String(queueURL),
		DelaySeconds: aws.Int64(1),
	}

	// send
	sqsRes, err := svc.SendMessage(params)
	if err != nil {
		log.Fatal("sqs send error : ", err.Error())
	}

	log.Println("send : ", *sqsRes.MessageId)
}

// remove xml tags
// mention event contains HTML tags
func removeTag(str string) string {
	rep := regexp.MustCompile(`<("[^"]*"|'[^']*'|[^'">])*>`)
	str = rep.ReplaceAllString(str, "")
	return str
}
