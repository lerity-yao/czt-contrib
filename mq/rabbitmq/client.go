package rabbitmq

import (
	"context"
	"github.com/zeromicro/go-zero/core/logc"
)

func MustNewSenderClient(ctx context.Context, rabbitSenderConf RabbitSenderConf) (Sender, error) {
	logc.Infof(ctx, "连接mq: %v", rabbitSenderConf)
	return MustNewSender(rabbitSenderConf)
}
