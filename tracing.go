package tracing

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ouihealth/gqlgen-apollo-federated-traces/reports"
)

type (
	ApolloFederatedTracing struct{}
)

type ApolloFederatedTracingExtension struct {
	lock      sync.Mutex
	nodes     map[*graphql.FieldContext]*reports.Trace_Node
	root      *reports.Trace_Node
	startTime time.Time
	endTime   time.Time
}

func (e *ApolloFederatedTracingExtension) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&reports.Trace{
		Root:       e.root,
		StartTime:  timestamppb.New(e.startTime),
		EndTime:    timestamppb.New(e.endTime),
		DurationNs: uint64(e.endTime.Sub(e.startTime).Nanoseconds()),
	})
}
func (e *ApolloFederatedTracingExtension) MarshalText() ([]byte, error) {
	src, err := e.MarshalBinary()
	if err != nil {
		return nil, err
	}

	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	return dst, nil
}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = ApolloFederatedTracing{}

func (ApolloFederatedTracing) ExtensionName() string {
	return "ApolloFederatedTracing"
}

func (ApolloFederatedTracing) Validate(graphql.ExecutableSchema) error {
	return nil
}

func (ApolloFederatedTracing) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	start := graphql.Now()
	federatedTracingExtension, ok := graphql.GetExtension(ctx, "ftv1").(*ApolloFederatedTracingExtension)
	if !ok {
		return next(ctx)
	}

	fc := graphql.GetFieldContext(ctx)
	traceNode := &reports.Trace_Node{
		Id: &reports.Trace_Node_ResponseName{
			ResponseName: fc.Field.Name,
		},
		Type:       fc.Field.Definition.Type.String(),
		ParentType: fc.Object,
		StartTime:  uint64(start.Sub(federatedTracingExtension.startTime).Nanoseconds()),
	}

	defer func() {
		end := graphql.Now()
		traceNode.EndTime = uint64(end.Sub(federatedTracingExtension.startTime).Nanoseconds())
	}()

	federatedTracingExtension.lock.Lock()
	federatedTracingExtension.nodes[fc] = traceNode

	parentTraceNode, ok := federatedTracingExtension.nodes[fc.Parent]
	if !ok {
		switch {

		// Index node
		case fc.Parent.Index != nil:
			parentTraceNode = &reports.Trace_Node{
				Id: &reports.Trace_Node_Index{
					Index: uint32(*fc.Parent.Index),
				},
			}

			federatedTracingExtension.nodes[fc.Parent] = parentTraceNode
			grandParentTraceNode := federatedTracingExtension.nodes[fc.Parent.Parent]
			grandParentTraceNode.Child = append(grandParentTraceNode.Child, parentTraceNode)

		// Root node
		default:
			parentTraceNode = federatedTracingExtension.root
		}
	}

	parentTraceNode.Child = append(parentTraceNode.Child, traceNode)
	federatedTracingExtension.lock.Unlock()

	return next(ctx)
}

func (ApolloFederatedTracing) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	rc := graphql.GetOperationContext(ctx)
	federatedTracingExtension := &ApolloFederatedTracingExtension{
		nodes:     map[*graphql.FieldContext]*reports.Trace_Node{},
		root:      &reports.Trace_Node{},
		startTime: rc.Stats.OperationStart,
	}

	defer func() {
		federatedTracingExtension.endTime = graphql.Now()
	}()

	graphql.RegisterExtension(ctx, "ftv1", federatedTracingExtension)
	return next(ctx)
}
