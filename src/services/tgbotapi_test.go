package services

import "testing"

func TestTGSendMessage(t *testing.T) {
	type args struct {
		chatid  int64
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TGSendMessage(tt.args.chatid, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("TGSendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
