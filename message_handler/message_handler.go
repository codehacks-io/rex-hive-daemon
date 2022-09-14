package message_handler

import (
	"fmt"
	"rex-daemon/swarm_message"
	"sync"
)

var (
	messages []*swarm_message.SwarmMessage
	lock     sync.Mutex
)

var didStartup = false

func Run() {
	if didStartup {
		return
	}
	didStartup = true
}

func ProcessSwarmMessage(message *swarm_message.SwarmMessage) {
	lock.Lock()
	messages = append(messages, message)
	lock.Unlock()
	fmt.Println(fmt.Sprintf("(%d) Received msg from %d: %+v", len(messages), (*message).Pid, *message))
}
