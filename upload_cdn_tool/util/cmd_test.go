// ignore_security_alert_file RCE
package util

import "testing"

func TestExecShell(t *testing.T) {
	type args struct {
		cmdString string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			args: args{
				cmdString: `
echo "123"
sleep 1
echo "456"
				`,
			},
			wantOutput: "123\n456\n",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutput, err := ExecShell(tt.args.cmdString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecShell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutput != tt.wantOutput {
				t.Errorf("ExecShell() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}
