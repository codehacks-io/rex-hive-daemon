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

func ProcessSwarmMessage(message *swarm_message.SwarmMessage) {
	lock.Lock()
	messages = append(messages, message)
	lock.Unlock()
	fmt.Println(fmt.Sprintf("(%d) Received msg from %d", len(messages), (*message).Pid))
}
