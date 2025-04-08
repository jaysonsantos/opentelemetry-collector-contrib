// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlfuncs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

func Test_filter(t *testing.T) {
	tests := []struct {
		name       string
		target     []any
		expression string
		expected   []any
	}{
		{
			name:       "simple",
			target:     []any{1, 2, 3, 4},
			expression: "value >= 2",
			expected:   []any{2, 3, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := pcommon.NewSlice()
			err := m.FromRaw(tt.target)
			assert.NoError(t, err)
			target := ottl.StandardPSliceGetter[any]{
				Getter: func(_ context.Context, _ any) (any, error) {
					return m, nil
				},
			}

			exprFunc, err := filter[any](target, tt.expression)
			assert.NoError(t, err)
			rv, err := exprFunc(nil, nil)
			assert.NoError(t, err)
			slice, ok := rv.(pcommon.Slice)
			assert.True(t, ok)

			raw := slice.AsRaw()
			assert.True(t, compareSlices(tt.expected, raw))
		})
	}
}
