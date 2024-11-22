package binance_connector

import (
	"context"
	"reflect"
	"testing"
)

const (
	apikey_TEST    = "bchvEMgFA3BYLfrC7WrDI4KW8PbdqfPgW9PzT8FrzjUZsegv1lpC43KEMB8jNaAi"
	secretKey_TEST = "gRENt6IKY0jclGs8tM6QLeWOWPgBD8wDFO7vjQMmeVggWo8AFb6lWVw3F9mwCuEl"
)

func TestGetUMAccountAssetService_Do(t *testing.T) {
	type fields struct {
		c *Client
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes *AccountAssetResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				c: NewClient(apikey_TEST, secretKey_TEST),
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: &AccountAssetResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GetUMAccountAssetService{
				c: tt.fields.c,
			}
			gotRes, err := s.Do(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUMAccountAssetService.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetUMAccountAssetService.Do() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestGetUMPositionService_Do(t *testing.T) {
	type fields struct {
		c *Client
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes []*UMPositionResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				c: NewClient(apikey_TEST, secretKey_TEST),
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: []*UMPositionResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GetUMPositionService{
				c: tt.fields.c,
			}
			gotRes, err := s.Do(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUMPositionService.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetUMPositionService.Do() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestGetUMAccountBalanceService_Do(t *testing.T) {
	type fields struct {
		c     *Client
		asset string
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes []*UMAccountBalanceResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				c:     NewClient(apikey_TEST, secretKey_TEST),
				asset: "USDT",
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: []*UMAccountBalanceResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GetUMAccountBalanceService{
				c:     tt.fields.c,
				asset: tt.fields.asset,
			}
			gotRes, err := s.Do(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUMAccountBalanceService.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetUMAccountBalanceService.Do() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestGetUMPositionService_DoUpdate(t *testing.T) {
	type fields struct {
		c                *Client
		dualSidePosition string
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes *UMStandardResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				c:                NewClient(apikey_TEST, secretKey_TEST),
				dualSidePosition: "true",
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: &UMStandardResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GetUMPositionService{
				c:                tt.fields.c,
				dualSidePosition: tt.fields.dualSidePosition,
			}
			gotRes, err := s.DoUpdate(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUMPositionService.DoUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetUMPositionService.DoUpdate() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestGetNewOrderService_Do(t *testing.T) {
	type fields struct {
		c                *Client
		symbol           *string
		side             *OrderSide
		positionSide     PositionSide
		ordertype        OrderType
		quantity         float64
		reduceOnly       *bool
		price            float64
		newClientOrderId string
		timeInForce      TimeInForce
		isSkipCheck      bool
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes []*UMNewOrderResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "限價單",
			fields: fields{
				c: NewClient(apikey_TEST, secretKey_TEST),
				symbol: func() *string {
					s := "BTCUSDT"
					return &s
				}(),
				side: func() *OrderSide {
					o := Buy
					return &o
				}(),
				positionSide: PositionLong,
				ordertype:    LimitOrder,
				quantity:     0.002,
				price:        85000,
				timeInForce:  GTC,
				isSkipCheck:  true, //暫時不測試
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: []*UMNewOrderResponse{},
			wantErr: false,
		},
		{
			name: "市價單",
			fields: fields{
				c: NewClient(apikey_TEST, secretKey_TEST),
				symbol: func() *string {
					s := "BTCUSDT"
					return &s
				}(),
				side: func() *OrderSide {
					o := Buy
					return &o
				}(),
				positionSide: PositionLong,
				ordertype:    MarketOrder,
				quantity:     0.002,
				isSkipCheck:  false,
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: []*UMNewOrderResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.isSkipCheck {
				t.Skip()
			}
			s := &GetUMNewOrderService{
				c:                tt.fields.c,
				symbol:           tt.fields.symbol,
				side:             tt.fields.side,
				positionSide:     tt.fields.positionSide,
				ordertype:        tt.fields.ordertype,
				quantity:         tt.fields.quantity,
				reduceOnly:       tt.fields.reduceOnly,
				price:            tt.fields.price,
				newClientOrderId: tt.fields.newClientOrderId,
				timeInForce:      tt.fields.timeInForce,
			}
			gotRes, err := s.Do(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNewOrderService.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetNewOrderService.Do() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestGetUMOrderService_Do(t *testing.T) {
	type fields struct {
		c       *Client
		symbol  *string
		orderid *int64
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes []*UMOrderResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				c:       NewClient(apikey_TEST, secretKey_TEST),
				symbol:  func() *string { s := "BTCUSDT"; return &s }(),
				orderid: func() *int64 { i := int64(487888498881); return &i }(),
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: []*UMOrderResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GetUMOrderService{
				c:       tt.fields.c,
				symbol:  tt.fields.symbol,
				orderid: tt.fields.orderid,
			}
			gotRes, err := s.Do(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUMOrderService.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetUMOrderService.Do() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestGetUMLeverageService_Do(t *testing.T) {
	type fields struct {
		c        *Client
		symbol   *string
		leverage *int64
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes *UMLeverageResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				c:        NewClient(apikey_TEST, secretKey_TEST),
				symbol:   func() *string { s := "BTCUSDT"; return &s }(),
				leverage: func() *int64 { i := int64(20); return &i }(),
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: &UMLeverageResponse{
				Leverage:         20,
				Symbol:           "BTCUSDT",
				MaxNotionalValue: "100000000",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GetUMLeverageService{
				c:        tt.fields.c,
				symbol:   tt.fields.symbol,
				leverage: tt.fields.leverage,
			}
			gotRes, err := s.Do(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUMLeverageService.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetUMLeverageService.Do() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestGetUMPositionRiskService_Do(t *testing.T) {
	type fields struct {
		c      *Client
		symbol string
	}
	type args struct {
		ctx  context.Context
		opts []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes []*UMPositionRiskResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				c:      NewClient(apikey_TEST, secretKey_TEST),
				symbol: "BTCUSDT",
			},
			args: args{
				ctx: context.Background(),
			},
			wantRes: []*UMPositionRiskResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GetUMPositionRiskService{
				c:      tt.fields.c,
				symbol: tt.fields.symbol,
			}
			gotRes, err := s.Do(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUMPositionRiskService.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetUMPositionRiskService.Do() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
