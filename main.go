package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dustin/go-broadcast"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

var b broadcast.Broadcaster
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//	fmt.Printf("TOPIC: %s\n", msg.Topic())
	//	fmt.Printf("MSG: %s\n", msg.Payload())
	b.Submit(msg.Payload())
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	b = broadcast.NewBroadcaster(100)
        url := os.Getenv("MQTT_URL")
        deviceId := os.Getenv("MQTT_DEVICE_ID")
        password := os.Getenv("MQTT_PASSWORD")

        log.Println(url)
        opts := mqtt.NewClientOptions().AddBroker(url).SetClientID(deviceId)
	//opts := mqtt.NewClientOptions().AddBroker("ssl://mqtt.openprio.nl:8883").SetClientID("mqtt-to-websocket")
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetUsername(deviceId)
	//opts.SetPassword("!*zcqXCsD/fn:24)")
        opts.SetPassword(password)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := c.Subscribe("#", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	http.HandleFunc("/positions", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
		if err != nil {
			log.Print(err)
		}
		ch := make(chan interface{})
		b.Register(ch)
                defer b.Unregister(ch)
                defer log.Println("Broker is done")

		for {
			data := <-ch
			log.Println("test")
			// Write message back to browser
			if err := conn.WriteMessage(2, data.([]byte)); err != nil {
				return
			}
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
