package http

import (
	"io"
	"net/http"

	net_message "github.com/number571/go-peer/pkg/network/message"
	"github.com/number571/go-peer/pkg/storage/cache"
	"github.com/number571/hidden-lake/build"
	http_logger "github.com/number571/hidden-lake/internal/utils/logger/http"
)

func (p *sHTTPAdapter) produceHandler() func(http.ResponseWriter, *http.Request) {
	adapterSettings := p.fSettings.GetAdapterSettings()
	cache := cache.NewLRUCache(build.GSettings.FNetworkManager.FCacheHashesCap)

	return func(w http.ResponseWriter, r *http.Request) {
		logBuilder := http_logger.NewLogBuilder(p.fName.Short(), r)

		if r.Method != http.MethodPost {
			p.fLogger.PushWarn(logBuilder.WithMessage(http_logger.CLogMethod))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		msgSize := adapterSettings.GetMessageSizeBytes()
		msgLen := uint64(msgSize+net_message.CMessageHeadSize) << 1 // nolint: unconvert
		msgStr := make([]byte, msgLen)
		n, err := io.ReadFull(r.Body, msgStr)
		if err != nil || uint64(n) != msgLen {
			p.fLogger.PushWarn(logBuilder.WithMessage("read_message"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		msg, err := net_message.LoadMessage(adapterSettings, string(msgStr))
		if err != nil {
			p.fLogger.PushWarn(logBuilder.WithMessage("load_message"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hash := msg.GetHash()
		if ok := cache.Set(hash, []byte{}); !ok {
			p.fLogger.PushWarn(logBuilder.WithMessage("message_exist"))
			w.WriteHeader(http.StatusAlreadyReported)
			return
		}

		p.fLogger.PushInfo(logBuilder.WithMessage(http_logger.CLogSuccess))
		p.fNetMsgChan <- msg
	}
}
