// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlfuncs // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

type FilterArguments[K any] struct {
	Target ottl.PSliceGetter[K]
	Filter string
}

func NewFilterFactory[K any]() ottl.Factory[K] {
	return ottl.NewFactory("Filter", &FilterArguments[K]{}, createFilterFunction[K])
}

func createFilterFunction[K any](_ ottl.FunctionContext, oArgs ottl.Arguments) (ottl.ExprFunc[K], error) {
	args, ok := oArgs.(*FilterArguments[K])

	if !ok {
		return nil, fmt.Errorf("FilterFactory args must be of type *FilterArguments[K]")
	}

	return filter(args.Target, args.Filter)
}

func filter[K any](target ottl.PSliceGetter[K], filterExpr string) (ottl.ExprFunc[K], error) {
	var currentValue any
	pathGetter := func(p ottl.Path[K]) (ottl.GetSetter[K], error) {
		return ottl.StandardGetSetter[K]{Getter: func(_ context.Context, tCtx K) (any, error) {
			if p.Name() != "value" {
				return nil, fmt.Errorf("only value is supported for filter functions")
			}
			return currentValue, nil
		}, Setter: nil}, nil
	}
	subParser, _ := ottl.NewParser[K](StandardConverters[K](), pathGetter, componenttest.NewNopTelemetrySettings())
	stmt, err := subParser.ParseCondition(filterExpr)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, tCtx K) (any, error) {
		slice, err := target.Get(ctx, tCtx)
		if err != nil {
			return nil, err
		}
		buffer := make([]any, 0)

		for _, value := range slice.AsRaw() {
			currentValue = value
			shouldInclude, err := stmt.Eval(ctx, tCtx)
			if err != nil {
				return nil, err
			}
			if shouldInclude {
				buffer = append(buffer, value)
			}
		}

		output := pcommon.NewSlice()
		err = output.FromRaw(buffer)
		return output, err
	}, nil
}
