package main

import (
	"fmt"
	"github.com/traduality/Traduality/lib/windowsprintdebug"
	"log"
	"log/slog"
	"os"
	"runtime/debug"
	"unsafe"
)

var loremIpsumText string = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque elementum euismod erat eu egestas. In lobortis eros a rhoncus blandit. Ut id elementum metus, vitae vestibulum augue. Phasellus sit amet leo quam. Vivamus laoreet massa neque, ac luctus sem posuere eu. In dapibus euismod justo, a tempor elit commodo quis. Proin lobortis fringilla sagittis. Donec et eleifend dolor. Nunc laoreet, nisi nec faucibus finibus, justo elit tincidunt leo, sit amet cursus orci dolor semper orci. Cras lectus urna, interdum sed gravida sit amet, vulputate quis ipsum. Duis euismod dolor in ante egestas, eget accumsan turpis dapibus. Nunc eu mauris vitae quam ultricies placerat. Vivamus aliquam efficitur lorem, eget rhoncus ante tristique vitae. Donec quis augue id risus ultrices accumsan. Nunc ut purus nec libero suscipit bibendum at et tortor. Mauris placerat libero ut ipsum posuere rutrum. Maecenas quis commodo nisl, ac auctor purus. Aenean euismod arcu elit, eu varius dolor porttitor et nullam."

func main() {
	var err error

	switch os.Args[1] {
	case "TestEmptyDebugWriteDoesNotCallOutputDebugString":
		windowsprintdebug.Initialize()
		var data []byte = []byte("hello\000")
		windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[0]), 0)
		windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[0]), 0)
		windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[0]), 0)

	case "TestShortOutputDebugStringIsWrittenToDebugger":
		windowsprintdebug.Initialize()
		var data []byte = []byte("hello\000")
		windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[0]), 5)

	case "TestOutputDebugStringWithoutNullTerminatorIsWrittenToDebuggerWithNullTerminator":
		windowsprintdebug.Initialize()
		var data []byte = []byte("helloworld")
		windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[0]), 5) // Just "hello".

	case "TestDebugWriteReturnsNumberOfBytesWritten":
		windowsprintdebug.Initialize()
		var data []byte = []byte("hello world!")
		var written int32
		written = windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[0]), 5) // Just "hello".
		if written != 5 {
			fmt.Fprintf(os.Stderr, "error: DebugWrite returned %d, expected 5\n", written)
			os.Exit(1)
		}
		written = windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[5]), 7) // Just " world!".
		if written != 7 {
			fmt.Fprintf(os.Stderr, "error: DebugWrite returned %d, expected 7\n", written)
			os.Exit(2)
		}

	case "TestLongOutputDebugStringIsWrittenToDebuggerInPieces":
		windowsprintdebug.Initialize()
		var data []byte = []byte(loremIpsumText + "\000")
		windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[0]), 500)

	case "TestRegisteredPrintOutputIsWrittenToDebugger":
		windowsprintdebug.Register()
		print("hello", 123)
		print("world\n")

	case "TestFmtPrintfOutputIsWrittenToDebugger":
		windowsprintdebug.Register()
		fmt.Printf("hello%d", 123)
		fmt.Printf("world\n")

	case "TestRegisteredOSStderrOutputIsWrittenToDebugger":
		windowsprintdebug.Register()
		os.Stderr.Write([]byte("hello"))
		os.Stderr.Write([]byte("world\n"))

	case "TestRegisteredLogOutputIsWrittenToDebugger":
		windowsprintdebug.Register()
		log.Printf("hello %d", 42)

	case "TestRegisteredSlogOutputIsWrittenToDebugger":
		windowsprintdebug.Register()
		slog.Info("hello", slog.Int("value", 42))

	case "TestRegisteredPanicOutputIsWrittenToDebugger":
		windowsprintdebug.Register()
		panic("this is a panic!")

	case "TestRegisterDoesNotInterfereWithFilesystemWrites":
		err = windowsprintdebug.Register()
		if err != nil {
			print(err)
			os.Exit(1)
		}
		err = os.WriteFile(os.Args[2], []byte("hello\n"), 0600)
		if err != nil {
			print(err)
			os.Exit(1)
		}

	case "TestRuntimeDebugSetCrashOutputLogsPanicToBothFileAndOutputDebugString":
		err = windowsprintdebug.Register()
		if err != nil {
			print(err)
			os.Exit(1)
		}
		var crashOutputFile *os.File
		crashOutputFile, err = os.Create(os.Args[2])
		if err != nil {
			print(err)
			os.Exit(1)
		}
		debug.SetCrashOutput(crashOutputFile, debug.CrashOptions{})
		crashOutputFile.Close()

		panic("This is a panic.")

	case "TestDebugWriteDoesNotCrashWithoutDebugger":
		err = windowsprintdebug.Initialize()
		if err != nil {
			print(err)
			os.Exit(1)
		}
		var data []byte = []byte("hello\000")
		windowsprintdebug.DebugWrite(2, unsafe.Pointer(&data[0]), 5)

	case "TestLogToFileWritesEverythingToFileWithoutDebugger":
		var logFile *os.File
		logFile, err = os.Create(os.Args[2])
		if err != nil {
			print(err)
			os.Exit(1)
		}
		err = windowsprintdebug.SetLogToFile(logFile)
		if err != nil {
			print(err)
			os.Exit(1)
		}
		logFile.Close()

		err = windowsprintdebug.Register()
		if err != nil {
			print(err)
			os.Exit(1)
		}

		print("print message here\n")
		log.Println("regular log message here")
		slog.Info("slog message here")
		panic("panic message here")

	default:
		slog.Error("invalid test name", slog.String("test_name", os.Args[1]))
		os.Exit(1)
	}
}
