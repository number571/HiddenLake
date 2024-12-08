package network

import (
	"context"
	"errors"
	"sync"

	"github.com/number571/go-peer/pkg/anonymity"
	"github.com/number571/go-peer/pkg/anonymity/queue"
	"github.com/number571/go-peer/pkg/client"
	"github.com/number571/go-peer/pkg/crypto/asymmetric"
	"github.com/number571/go-peer/pkg/encoding"
	net_message "github.com/number571/go-peer/pkg/network/message"
	"github.com/number571/go-peer/pkg/payload"
	"github.com/number571/go-peer/pkg/storage/database"
	"github.com/number571/hidden-lake/build"
	"github.com/number571/hidden-lake/pkg/adapters"
	"github.com/number571/hidden-lake/pkg/handler"
	"github.com/number571/hidden-lake/pkg/request"
	"github.com/number571/hidden-lake/pkg/response"
)

var (
	_ IHiddenLakeNode = &sHiddenLakeNode{}
)

type sHiddenLakeNode struct {
	fAnonymityNode anonymity.INode
}

func NewRawHiddenLakeNode(pAnonymityNode anonymity.INode) IHiddenLakeNode {
	return &sHiddenLakeNode{fAnonymityNode: pAnonymityNode}
}

func NewHiddenLakeNode(
	pSettings ISettings,
	pPrivKey asymmetric.IPrivKey,
	pKVDatabase database.IKVDatabase,
	pRunnerAdapter adapters.IRunnerAdapter,
	pHandlerF handler.IHandlerF,
) IHiddenLakeNode {
	return &sHiddenLakeNode{
		anonymity.NewNode(
			anonymity.NewSettings(&anonymity.SSettings{
				FServiceName:  pSettings.GetServiceName(),
				FFetchTimeout: pSettings.GetFetchTimeout(),
			}),
			pSettings.GetLogger(),
			pRunnerAdapter,
			pKVDatabase,
			queue.NewQBProblemProcessor(
				queue.NewSettings(&queue.SSettings{
					FMessageConstructSettings: net_message.NewConstructSettings(&net_message.SConstructSettings{
						FSettings: net_message.NewSettings(&net_message.SSettings{
							FWorkSizeBits: pSettings.GetWorkSizeBits(),
							FNetworkKey:   pSettings.GetNetworkKey(),
						}),
						FParallel: pSettings.GetParallel(),
					}),
					FQueuePeriod:  pSettings.GetQueuePeriod(),
					FNetworkMask:  build.GSettings.FProtoMask.FNetwork,
					FConsumersCap: build.GSettings.FQueueProblem.FConsumersCap,
					FQueuePoolCap: [2]uint64{
						build.GSettings.FQueueProblem.FMainPoolCap,
						build.GSettings.FQueueProblem.FRandPoolCap,
					},
				}),
				func() client.IClient {
					client := client.NewClient(pPrivKey, pSettings.GetMessageSizeBytes())
					if client.GetPayloadLimit() <= encoding.CSizeUint64 {
						panic(`client.GetPayloadLimit() <= encoding.CSizeUint64`)
					}
					return client
				}(),
			),
		).HandleFunc(
			build.GSettings.FProtoMask.FService,
			handler.RequestHandler(pHandlerF),
		),
	}
}

func (p *sHiddenLakeNode) GetAnonymityNode() anonymity.INode {
	return p.fAnonymityNode
}

func (p *sHiddenLakeNode) Run(pCtx context.Context) error {
	chCtx, cancel := context.WithCancel(pCtx)
	defer cancel()

	const N = 2

	errs := make([]error, N)
	wg := &sync.WaitGroup{}
	wg.Add(N)

	go func() {
		defer func() { wg.Done(); cancel() }()
		runnerAdapter := p.fAnonymityNode.GetAdapter().(adapters.IRunnerAdapter)
		errs[0] = runnerAdapter.Run(chCtx)
	}()

	go func() {
		defer func() { wg.Done(); cancel() }()
		errs[1] = p.fAnonymityNode.Run(chCtx)
	}()

	wg.Wait()

	select {
	case <-pCtx.Done():
		return pCtx.Err()
	default:
		return errors.Join(errs...)
	}
}

func (p *sHiddenLakeNode) SendRequest(
	pCtx context.Context,
	pPubKey asymmetric.IPubKey,
	pRequest request.IRequest,
) error {
	return p.fAnonymityNode.SendPayload(
		pCtx,
		pPubKey,
		payload.NewPayload64(
			uint64(build.GSettings.FProtoMask.FService),
			pRequest.ToBytes(),
		),
	)
}

func (p *sHiddenLakeNode) FetchRequest(
	pCtx context.Context,
	pPubKey asymmetric.IPubKey,
	pRequest request.IRequest,
) (response.IResponse, error) {
	rspBytes, err := p.fAnonymityNode.FetchPayload(
		pCtx,
		pPubKey,
		payload.NewPayload32(
			build.GSettings.FProtoMask.FService,
			pRequest.ToBytes(),
		),
	)
	if err != nil {
		return nil, err
	}
	return response.LoadResponse(rspBytes)
}
