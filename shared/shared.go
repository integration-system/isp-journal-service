package shared

import (
	"github.com/integration-system/isp-lib/streaming"
	"google.golang.org/grpc/metadata"
)

type LogController interface {
	Transfer(stream streaming.DuplexMessageStream, md metadata.MD) error
}
