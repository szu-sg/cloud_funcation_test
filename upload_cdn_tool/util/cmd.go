// ignore_security_alert_file RCE
package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"unsafe"
)

func ExecShell(cmdString string) (output string, err error) {
	cmd := exec.Command("bash", "-c", cmdString)
	if cmd == nil {
		return "", fmt.Errorf("cannot create command")
	}
	stdOutOrErrBuffer := new(bytes.Buffer)
	cmd.Stdout = stdOutOrErrBuffer
	cmd.Stderr = stdOutOrErrBuffer
	err = cmd.Run()
	output = stdOutOrErrBuffer.String()
	return
}

// UnsafeSliceToString zero-copy slice convert to string
func UnsafeSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
