package rabbitmq

import (
	"fmt"
)

// RabbitListenerConf 消费者配置
// RabbitListenerConf RabbitConf
// RabbitListenerConf ListenerQueues
// RabbitListenerConf ChannelQos
// RabbitListenerConf ContentType 如果需要重新推送消息，比如消费失败，发送报文类型
type RabbitListenerConf struct {
	RabbitConf
	ListenerQueues []ConsumerConf
	ChannelQos     ChannelQosConf
	ContentType    string `json:",default=text/plain"` // MIME content type
}

// RabbitSenderConf 客户端配置
// RabbitSenderConf RabbitConf
// RabbitSenderConf ContentType 发送报文类型
type RabbitSenderConf struct {
	RabbitConf
	ContentType string `json:",default=text/plain"` // MIME content type
}

func getRabbitURL(rabbitConf RabbitConf) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", rabbitConf.Username, rabbitConf.Password,
		rabbitConf.Host, rabbitConf.Port, rabbitConf.VHost)
}
