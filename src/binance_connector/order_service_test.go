package binance_connector

import (
	"context"
	"reflect"
	"testing"
)

func TestGetAllOpenOrderService_Do(t *testing.T) {
	type fields struct {
		c      *Client
		symbol *string
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes []*NewCurrentOpenOrderResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				c: NewClient("AiXIEYKnXUBe34bXqC6yG2DuPVgg2KtVyi2EhO5Nrk9SBMiy6au7xc32YuN14UKm", "1ALhqbFilvZbn1amG4zbh6sxbn9aO26svFoPFG3qgS4zHES8MsRT8t8A0GgA8Zry"),
				symbol: func() *string {
					s := "BTCUSDT"
					return &s
				}(),
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: []*NewCurrentOpenOrderResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GetAllOpenOrderService{
				c:      tt.fields.c,
				symbol: tt.fields.symbol,
			}
			gotRes, err := s.Do(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllOpenOrderService.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetAllOpenOrderService.Do() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
