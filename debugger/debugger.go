package debugger

import (
	"bytes"
	"google.golang.org/protobuf/proto"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// Generate the .pb.go files:
//go:generate protoc --proto_path=./ --go_out=./ --go_opt=module=github.com/traduality/Traduality/lib/windowsprintdebug/debugger protocol.proto

type RunResult struct {
	Messages    []*ProtocolMessage
	ProcessExit *ProtocolProcessExit
}

func (result *RunResult) OutputDebugStringCalls() [][]byte {
	var calls [][]byte = make([][]byte, 0, len(result.Messages))
	var message *ProtocolMessage
	for _, message = range result.Messages {
		if message.OutputDebugString != nil {
			calls = append(calls, message.OutputDebugString.Data)
		}
	}
	return calls
}

func (result *RunResult) ConcatenatedOutputDebugString() []byte {
	var output []byte = []byte{}
	var message *ProtocolMessage
	for _, message = range result.Messages {
		if message.OutputDebugString != nil {
			output = append(output, message.OutputDebugString.Data[:len(message.OutputDebugString.Data)-1]...)
		}
	}
	return output
}

func RunCommandWithDebugger(command ...string) RunResult {
	var err error

	compileDebuggerIfNeeded()

	var stdout = &bytes.Buffer{}
	var cmd *exec.Cmd = exec.Command(compiledDebuggerEXEPath, command...)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	var protocol Protocol
	err = proto.Unmarshal(stdout.Bytes(), &protocol)
	if err != nil {
		panic(err)
	}

	return RunResult{
		Messages:    protocol.Messages,
		ProcessExit: protocol.ProcessExit,
	}
}

var compileDebuggerOnce sync.Once = sync.Once{}
var compiledDebuggerEXEPath string

// Sets compiledDebuggerEXEPath.
func compileDebuggerIfNeeded() {
	compileDebuggerOnce.Do(func() {
		var err error
		var ok bool

		var scriptPath string
		_, scriptPath, _, ok = runtime.Caller(0)
		if !ok {
			panic("could not determine path of .go file")
		}
		var debuggerDirectory string = filepath.Dir(scriptPath)

		var debuggerEXEPath string = filepath.Join(debuggerDirectory, "debugger.exe")
		var cmd *exec.Cmd = exec.Command(
			"cl.exe",
			"/nologo",
			"/std:c++20",
			"/EHa",
			"/Fe:"+debuggerEXEPath,
			filepath.Join(debuggerDirectory, "main.cpp"),
		)
		cmd.Dir = debuggerDirectory // Place .obj files here.
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			panic(err)
		}

		compiledDebuggerEXEPath = debuggerEXEPath
	})
}
