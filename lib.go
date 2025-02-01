package windowsprintdebug

import (
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

var lock sync.Mutex = sync.Mutex{}

// Protected by lock.
var initialized bool = false
var registered bool = false

var logToFile atomic.Pointer[os.File] = atomic.Pointer[os.File]{}

var func_IsDebuggerPresent uintptr
var func_OutputDebugStringA uintptr
var func_WriteFile uintptr

var oldOverrideWrite writeFunc

func Initialize() error {
	var err error

	lock.Lock()
	defer lock.Unlock()
	if initialized {
		return nil
	}

	var kernel32 syscall.Handle
	kernel32, err = syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return err
	}
	func_OutputDebugStringA, err = syscall.GetProcAddress(kernel32, "OutputDebugStringA")
	if err != nil {
		return err
	}
	func_IsDebuggerPresent, err = syscall.GetProcAddress(kernel32, "IsDebuggerPresent")
	if err != nil {
		return err
	}
	func_WriteFile, err = syscall.GetProcAddress(kernel32, "WriteFile")
	if err != nil {
		return err
	}

	initialized = true
	return nil
}

func Register() error {
	var err error

	err = Initialize()
	if err != nil {
		return err
	}

	lock.Lock()
	defer lock.Unlock()
	if registered {
		return nil
	}

	// TODO(strager): Make these two operations atomic.
	oldOverrideWrite = overrideWrite
	overrideWrite = DebugWrite

	log.Default().SetOutput(OutputDebugStringWriter)

	registered = true
	return nil
}

func SetLogToFile(file *os.File) error {
	var err error

	var duplicatedFile *os.File
	if file == nil {
		// Do not log to a file.
		duplicatedFile = nil
	} else {
		duplicatedFile, err = duplicateFile(file)
		if err != nil {
			return err
		}
	}
	logToFile.Store(duplicatedFile)
	return nil
}

func duplicateFile(file *os.File) (*os.File, error) {
	var err error
	var fileName string = file.Name()
	var process syscall.Handle
	process, err = syscall.GetCurrentProcess()
	if err != nil {
		return nil, err
	}
	var duplicatedFileHandle syscall.Handle
	err = syscall.DuplicateHandle(
		process, syscall.Handle(file.Fd()),
		process, &duplicatedFileHandle,
		/*dwDesiredAccess:*/ 0,
		/*bInheritHandle:*/ false,
		/*dwFlags:*/ syscall.DUPLICATE_SAME_ACCESS,
	)
	runtime.KeepAlive(file) // Prevent file.Close() from being called before syscall.DuplicateHandle returns.
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(duplicatedFileHandle), fileName), nil
}

type outputDebugStringWriter struct{}

func (w *outputDebugStringWriter) Write(data []byte) (int, error) {
	if len(data) > 0 {
		var bytesWritten int32 = DebugWrite(2, unsafe.Pointer(&data[0]), int32(len(data)))
		return int(bytesWritten), nil
	}
	return 0, nil
}

var OutputDebugStringWriter io.Writer = &outputDebugStringWriter{}

type writeFunc = func(fd uintptr, p unsafe.Pointer, n int32) int32

// See src/runtime/time_nofake.go and src/runtime/os_windows.go in the Go source tree.
//
//go:linkname overrideWrite runtime.overrideWrite
var overrideWrite writeFunc

// Precondition: Initialize() was previously called.
func DebugWrite(fd uintptr, p unsafe.Pointer, n int32) int32

// This function is called on the system stack.
//
// This function might be called from panic's crash handler which disables
// allocation. Therefore, this function should not allocate from the heap.
//
//go:nosplit
func debugWriteOnSystemStack(fd uintptr, p unsafe.Pointer, n int32) int32 {
	if fd == 1 || fd == 2 {
		// OutputDebugStringA raises a DBG_PRINTEXCEPTION_C (0x40010006)
		// exception. The exception handler in Go's runtime exits the
		// process when it sees this exception, which is undesired.
		//
		// Work around this by only calling OutputDebugStringA if a
		// debugger is attached. This is racy, as a debugger might be
		// detached in the middle of this function (causing the Go
		// runtime to exit the program).
		//
		// TODO(strager): Fix this bug in the Go runtime.
		if cIsDebuggerPresent() {
			// Here we call OutputDebugStringA in chunks. This is necessary
			// because OutputDebugStringA requires a null-terminated string,
			// but p is not null-terminated.
			var pSlice []byte = unsafe.Slice((*byte)(p), n)
			var buffer [128]byte = [128]byte{}
			for len(pSlice) > 0 {
				var copiedSize int = copy(buffer[:len(buffer)-1], pSlice)
				pSlice = pSlice[copiedSize:]
				buffer[copiedSize] = 0 // Null terminator for the C string.
				cOutputDebugStringA(unsafe.Pointer(&buffer[0]))
			}
		}

		var logFile *os.File = logToFile.Load()
		if logFile == nil {
			return n
		} else {
			var bytesWritten uint32 = 0
			_ = cWriteFile(logFile.Fd(), p, uint32(n), &bytesWritten, unsafe.Pointer(nil))
			return int32(bytesWritten)
		}
	} else {
		// Emulate the behavior of the original runtime.write function.
		var bytesWritten uint32 = 0
		_ = cWriteFile(fd, p, uint32(n), &bytesWritten, unsafe.Pointer(nil))
		return int32(bytesWritten)
		// Do not also log to logToFile.
	}
}

func cIsDebuggerPresent() bool

//go:noescape
func cOutputDebugStringA(lpOutputString unsafe.Pointer)

//go:noescape
func cWriteFile(
	hFile uintptr,
	lpBuffer unsafe.Pointer,
	nNumberOfBytesToWrite uint32,
	lpNumberOfBytesWritten *uint32,
	lpOverlapped unsafe.Pointer,
) bool
