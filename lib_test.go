package windowsprintdebug

import (
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/traduality/windowsprintdebug/debugger"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var compileTestprogOnce sync.Once = sync.Once{}
var compiledTestprogEXEPath string

// Returns the path to the compiled testprog.exe.
func compileTestprogIfNeeded() string {
	compileTestprogOnce.Do(func() {
		var err error
		var ok bool

		var scriptPath string
		_, scriptPath, _, ok = runtime.Caller(0)
		if !ok {
			panic("could not determine path of .go file")
		}
		var windowsprintdebugDirectory string = filepath.Dir(scriptPath)

		var testprogEXEPath string = filepath.Join(windowsprintdebugDirectory, "testprog", "testprog.exe")
		var cmd *exec.Cmd = exec.Command("go", "build", "-o", testprogEXEPath)
		cmd.Dir = filepath.Join(windowsprintdebugDirectory, "testprog")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			panic(err)
		}

		compiledTestprogEXEPath = testprogEXEPath
	})
	return compiledTestprogEXEPath
}

func TestEmptyDebugWriteDoesNotCallOutputDebugString(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestEmptyDebugWriteDoesNotCallOutputDebugString")
	require.Equal(t, 0, len(result.Messages))
}

func TestShortOutputDebugStringIsWrittenToDebugger(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestShortOutputDebugStringIsWrittenToDebugger")
	require.Equal(t, 1, len(result.Messages))
	require.NotNil(t, result.Messages[0].OutputDebugString)
	require.Equal(t, []byte("hello\000"), result.Messages[0].OutputDebugString.Data)
}

func TestLongOutputDebugStringIsWrittenToDebuggerInPieces(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestLongOutputDebugStringIsWrittenToDebuggerInPieces")
	// Assumption: Each write is limited to 127 bytes. This corresponds to
	// BUFFER_SIZE in the implementation.
	var expectedWrites [][]byte = [][]byte{
		// Each write (except the last) is 128 bytes long.
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque elementum euismod erat eu egestas. In lobortis eros a rho\000"),
		[]byte("ncus blandit. Ut id elementum metus, vitae vestibulum augue. Phasellus sit amet leo quam. Vivamus laoreet massa neque, ac luctu\000"),
		[]byte("s sem posuere eu. In dapibus euismod justo, a tempor elit commodo quis. Proin lobortis fringilla sagittis. Donec et eleifend do\000"),
		[]byte("lor. Nunc laoreet, nisi nec faucibus finibus, justo elit tincidunt leo, sit amet cursus orci dolor semper orci. Cras le\000"),
	}
	var actualWrites [][]byte = result.OutputDebugStringCalls()
	require.EqualValues(t, expectedWrites, actualWrites)
}

func TestOutputDebugStringWithoutNullTerminatorIsWrittenToDebuggerWithNullTerminator(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestOutputDebugStringWithoutNullTerminatorIsWrittenToDebuggerWithNullTerminator")
	require.Equal(t, 1, len(result.Messages))
	require.NotNil(t, result.Messages[0].OutputDebugString)
	require.Equal(t, []byte("hello\000"), result.Messages[0].OutputDebugString.Data)
}

func TestDebugWriteReturnsNumberOfBytesWritten(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestDebugWriteReturnsNumberOfBytesWritten")
	// testprog performs the assertions for us, setting a non-zero exit code
	// on failure.
	require.EqualValues(t, 0, *result.ProcessExit.ExitCode)
}

func TestRegisteredPrintOutputIsWrittenToDebugger(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestRegisteredPrintOutputIsWrittenToDebugger")
	var output string = string(result.ConcatenatedOutputDebugString())
	require.Equal(t, "hello123world\n", output)
}

func TestRegisteredFmtPrintfOutputIsWrittenToDebugger(t *testing.T) {
	t.Skip() // TODO(strager)
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestRegisteredFmtPrintfOutputIsWrittenToDebugger")
	var output string = string(result.ConcatenatedOutputDebugString())
	require.Equal(t, "hello123world\n", output)
}

func TestRegisteredOSStderrOutputIsWrittenToDebugger(t *testing.T) {
	t.Skip() // TODO(strager)
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestRegisteredOSStderrOutputIsWrittenToDebugger")
	var output string = string(result.ConcatenatedOutputDebugString())
	require.Equal(t, "helloworld\n", output)
}

func TestRegisteredLogOutputIsWrittenToDebugger(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestRegisteredLogOutputIsWrittenToDebugger")
	var output string = string(result.ConcatenatedOutputDebugString())
	require.Contains(t, output, "hello 42\n")
}

func TestRegisteredSlogOutputIsWrittenToDebugger(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestRegisteredSlogOutputIsWrittenToDebugger")
	var output string = string(result.ConcatenatedOutputDebugString())
	require.Contains(t, output, "hello value=42\n")
}

func TestRegisteredPanicOutputIsWrittenToDebugger(t *testing.T) {
	var testprogEXE string = compileTestprogIfNeeded()
	var result debugger.RunResult = debugger.RunCommandWithDebugger(testprogEXE, "TestRegisteredPanicOutputIsWrittenToDebugger")
	var output string = string(result.ConcatenatedOutputDebugString())
	require.Contains(t, output, "panic: this is a panic!\n")
	require.Contains(t, output, "main.main()")
	require.Contains(t, output, "testprog/main.go")
}

func TestRegisterDoesNotInterfereWithFilesystemWrites(t *testing.T) {
	var err error

	var testprogEXE string = compileTestprogIfNeeded()
	var tempFile string = filepath.Join(t.TempDir(), "file.txt")
	var result debugger.RunResult = debugger.RunCommandWithDebugger(
		testprogEXE,
		"TestRegisterDoesNotInterfereWithFilesystemWrites",
		tempFile,
	)

	var fileContent []byte
	fileContent, err = os.ReadFile(tempFile)
	require.NoError(t, err)
	require.Equal(t, "hello\n", string(fileContent))

	var output string = string(result.ConcatenatedOutputDebugString())
	require.Equal(t, "", output)
}

func TestRuntimeDebugSetCrashOutputLogsPanicToBothFileAndOutputDebugString(t *testing.T) {
	var err error

	var testprogEXE string = compileTestprogIfNeeded()
	var tempFile string = filepath.Join(t.TempDir(), "file.txt")
	var result debugger.RunResult = debugger.RunCommandWithDebugger(
		testprogEXE,
		"TestRuntimeDebugSetCrashOutputLogsPanicToBothFileAndOutputDebugString",
		tempFile,
	)

	var fileContent []byte
	fileContent, err = os.ReadFile(tempFile)
	require.NoError(t, err)
	require.Contains(t, string(fileContent), "panic: This is a panic.\n")
	require.Contains(t, string(fileContent), "main.main()")
	require.Contains(t, string(fileContent), "testprog/main.go")

	var output string = string(result.ConcatenatedOutputDebugString())
	require.Equal(t, string(fileContent), output)
}

func TestDebugWriteDoesNotCrashWithoutDebugger(t *testing.T) {
	var err error
	var testprogEXE string = compileTestprogIfNeeded()
	var cmd *exec.Cmd = exec.Command(testprogEXE, "TestDebugWriteDoesNotCrashWithoutDebugger")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err)
}

func TestLogToFileWritesEverythingToFileWithoutDebugger(t *testing.T) {
	var err error

	var tempFile string = filepath.Join(t.TempDir(), "file.txt")
	var testprogEXE string = compileTestprogIfNeeded()
	var cmd *exec.Cmd = exec.Command(
		testprogEXE,
		"TestLogToFileWritesEverythingToFileWithoutDebugger",
		tempFile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		// Expected.
	} else {
		require.NoError(t, err)
	}

	var fileContent []byte
	fileContent, err = os.ReadFile(tempFile)
	require.NoError(t, err)
	require.Contains(t, string(fileContent), "print message here\n")
	require.Contains(t, string(fileContent), "regular log message here\n")
	require.Contains(t, string(fileContent), "slog message here\n")
	require.Contains(t, string(fileContent), "panic: panic message here\n")
	require.Contains(t, string(fileContent), "main.main()")      // panic stack trace.
	require.Contains(t, string(fileContent), "testprog/main.go") // panic stack trace.
	// TODO(strager): Also test fmt.Printf, os.Stdout, and os.Stderr.
}

// TODO(strager): Ensure UTF-8 code points are not split.
