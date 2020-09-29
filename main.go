package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dustin/go-broadcast"
	"github.com/zeromq/goczmq"
	"github.com/eclipse/paho.mqtt.golang"
	"C"
)

var b broadcast.Broadcaster
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	b.Submit(msg.Payload())
}

func connectionLostHandler(client mqtt.Client, reason error) {
	log.Println("Connection lost:", reason.Error())
        time.Sleep(10 * time.Second)
        os.Exit(1)
}

func main() {
	b = broadcast.NewBroadcaster(100)
	url := os.Getenv("MQTT_URL")
	deviceId := os.Getenv("MQTT_DEVICE_ID")
	password := os.Getenv("MQTT_PASSWORD")

	log.Println(url)
	opts := mqtt.NewClientOptions().AddBroker(url).SetClientID(deviceId)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(2 * time.Second)
	opts.SetUsername(deviceId)
	opts.SetPassword(password)
	opts.SetConnectionLostHandler(connectionLostHandler)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := c.Subscribe("/prod/pt/position/#", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	startZmq(b)
}

func startZmq(b broadcast.Broadcaster) {
	ch := make(chan interface{})
	b.Register(ch)
	defer b.Unregister(ch)

	zmqServer := os.Getenv("ZMQ_IP")
	dealer := goczmq.NewDealerChanneler(zmqServer)
	defer dealer.Destroy()

	log.Println("dealer created and connected")
	log.Println("Start sending data to ZeroMQ")
	lastLogTime := time.Now()
	counter := 0
	for {
		data := <-ch
		dealer.SendChan <- [][]byte{[]byte("/openprio"), data.([]byte)}
		counter = counter + 1
		if time.Now().Sub(lastLogTime) > time.Minute * 1 {
			lastLogTime = time.Now()
			log.Printf("Sent %d messages since previous log", counter)
			counter = 0
		}
	}
}
