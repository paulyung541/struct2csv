package struct2csv

import (
	"reflect"
	"testing"
)

type testStruct struct {
	A1 int
	A2 bool
	B1 []int
	B2 []struct {
		B21 int
		B22 string
	}
	B3 []*struct {
		B31 []*int
	}
}

func TestStructConverter_Convert(t *testing.T) {
	conv, _ := NewStructConverter(NewHeaderAutoIncrementConv())

	type args struct {
		data interface{}
	}
	type wants struct {
		paths []string // un encode key
	}
	tests := []struct {
		name    string
		args    args
		want    wants
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				data: []testStruct{{
					A1: 1,
					A2: true,
					B1: []int{2, 3},
					B2: []struct {
						B21 int
						B22 string
					}{
						{
							B21: 4,
							B22: "a",
						},
					},
					B3: []*struct {
						B31 []*int
					}{
						{
							B31: []*int{newInt(5)},
						},
					},
				},
					{
						A1: 1,
						A2: true,
						B1: []int{2, 3},
						B2: []struct {
							B21 int
							B22 string
						}{
							{
								B21: 4,
								B22: "a",
							},
						},
						B3: []*struct {
							B31 []*int
						}{
							{
								B31: []*int{newInt(5), newInt(6)},
							},
						},
					},
				},
			},
			want: wants{
				paths: []string{
					"/A1",
					"/A2",
					"/B1/0",
					"/B1/1",
					"/B2/0/B21",
					"/B2/0/B22",
					"/B3/0/B31/0",
					"/B3/0/B31/1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.Convert(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			paths := got.GetUnEncodedSortHeader()
			if !reflect.DeepEqual(paths, tt.want.paths) {
				t.Errorf("Convert() got UnEncodedSortHeader = %v, want %v", paths, tt.want.paths)
			}
		})
	}
}

func newInt(i int) *int {
	return &i
}
