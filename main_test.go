package main

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/timoreimann/mdtf/pkg"
)

func TestBuildOptions(t *testing.T) {
	tests := []struct {
		name      string
		noPad     bool
		lineRange string
		want      pkg.Options
		wantErr   string
	}{
		{name: "defaults", want: pkg.Options{Pad: true}},
		{name: "no padding", noPad: true, want: pkg.Options{Pad: false}},
		{name: "range", lineRange: "5:12", want: pkg.Options{Pad: true, Start: 5, End: 12}},
		{name: "range with spaces", lineRange: " 5 : 12 ", want: pkg.Options{Pad: true, Start: 5, End: 12}},
		{name: "missing colon", lineRange: "5", wantErr: `invalid range "5": expected format start:end (e.g. 5:12)`},
		{name: "invalid start", lineRange: "a:12", wantErr: `invalid range start "a"`},
		{name: "invalid end", lineRange: "5:b", wantErr: `invalid range end "b"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildOptions(tt.noPad, tt.lineRange)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
