package security

import (
	"testing"
	"time"
)

// TestParseDuration 覆盖标准时长、天数扩展及非法输入。
func TestParseDuration(t *testing.T) {
	// args 表示单个时长解析用例的输入。
	type args struct {
		d string // d 是待解析的时长文本。
	}
	tests := []struct {
		name    string        // name 是子测试名称。
		args    args          // args 保存本用例输入。
		want    time.Duration // want 是期望时长。
		wantErr bool          // wantErr 表示是否期望解析失败。
	}{
		{
			name:    "5h20m",
			args:    args{"5h20m"},
			want:    time.Hour*5 + 20*time.Minute,
			wantErr: false,
		},
		{
			name:    "1d5h20m",
			args:    args{"1d5h20m"},
			want:    24*time.Hour + time.Hour*5 + 20*time.Minute,
			wantErr: false,
		},
		{
			name:    "1d",
			args:    args{"1d"},
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "invalid day value",
			args:    args{"invalid"},
			wantErr: true,
		},
		{
			name:    "invalid day suffix",
			args:    args{"1dnope"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDuration() got = %v, want %v", got, tt.want)
			}
		})
	}
}
