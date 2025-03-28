#define NOSPLIT	4

// // Size: 0x20
// type Closure struct {
// 	f unsafe.Pointer  // Offset 0x00: Pointer to debugWriteOnSystemStackTrampoline.
// 	fd uintptr        // Offset 0x08: Argument.
// 	p unsafe.Pointer  // Offset 0x10: Argument.
// 	n int32           // Offset 0x18: Argument.
// 	result int32      // Offset 0x1c: Return value of asmDebugWrite.
// }

// func DebugWrite(fd uintptr, p unsafe.Pointer, n int32) int32
TEXT ·DebugWrite(SB),NOSPLIT,$40-20
        // Stack layout:
        // arg-0x28(SP) *Closure     // 0x08 bytes
        // closure-0x20(SP) Closure  // 0x20 bytes
#define TEMP_closure(offset) closure-(0x20-offset)(SP)
        MOVQ $·debugWriteOnSystemStackTrampoline(SB), AX
        MOVQ AX, TEMP_closure(0x00)  // closure.f
        MOVQ fd+0(FP), AX
        MOVQ AX, TEMP_closure(0x08)  // closure.fd
        MOVQ p+8(FP), AX
        MOVQ AX, TEMP_closure(0x10)  // closure.p
        MOVL n+16(FP), AX
        MOVL AX, TEMP_closure(0x18)  // closure.n

        MOVQ $TEMP_closure(0), AX
        MOVQ AX, arg-0x28(SP)        // f = &closure
	CALL runtime·systemstack(SB)

        MOVQ TEMP_closure(0x1c), AX  // closure.result
        MOVQ AX, ret+24(FP)
        RET
#undef TEMP_closure

// This is a closure function.
//
// func debugWriteOnSystemStackTrampoline()
TEXT ·debugWriteOnSystemStackTrampoline(SB),NOSPLIT,$32-0
        // NOTE(strager): We received a pointer to the closure in DX.
        // FIXME(strager): Is this ABI-stable or should we use another mechanism
        // for accessing the closure?
        MOVQ DX, closure-0(SP)

        MOVQ 0x08(DX), AX                  // closure.fd
        MOVQ AX, arg_fd-32(SP)
        MOVQ 0x10(DX), AX                  // closure.p
        MOVQ AX, arg_p-24(SP)
        MOVL 0x18(DX), AX                  // closure.n
        MOVL AX, arg_n-16(SP)
        CALL ·debugWriteOnSystemStack(SB)  // Clobbers DX.
        MOVQ closure-0(SP), DX
        MOVL AX, 0x1c(DX)                  // closure.bytesWritten
        RET

// func cIsDebuggerPresent() bool
TEXT ·cIsDebuggerPresent(SB),NOSPLIT,$64-8
        // See NOTE[align-stack-x64].
	MOVQ SP, CX
	ANDQ $~15, SP
	MOVQ CX, 32(SP)

        // NOTE(strager): 32 bytes at the bottom of the stack are reserved for the callee.
	MOVQ ·func_IsDebuggerPresent(SB), BX  // NOTE(strager): Might clobber CX.
	CALL BX

        MOVQ 32(SP), CX
        MOVQ CX, SP

        MOVQ AX, ret+0(FP)
        RET

// func cOutputDebugStringA(lpOutputString unsafe.Pointer)
TEXT ·cOutputDebugStringA(SB),NOSPLIT,$64-8
        MOVQ lpOutputDebugString+0(FP), AX

        // See NOTE[align-stack-x64].
	MOVQ SP, CX
	ANDQ $~15, SP
	MOVQ CX, 32(SP)

        // NOTE(strager): 32 bytes at the bottom of the stack are reserved for the callee.
	MOVQ ·func_OutputDebugStringA(SB), BX  // NOTE(strager): Might clobber CX.
        MOVQ AX, CX            // lpOutputString
	CALL BX

        MOVQ 32(SP), CX
        MOVQ CX, SP
        RET

// func cWriteFile(
// 	hFile uintptr,
// 	lpBuffer unsafe.Pointer,
// 	nNumberOfBytesToWrite uint32,
// 	lpNumberOfBytesWritten *uint32,
// 	lpOverlapped unsafe.Pointer,
// ) bool
TEXT ·cWriteFile(SB),NOSPLIT,$64-48
        MOVQ hFile+0(FP), R10
        MOVQ lpBuffer+8(FP), DX
        MOVQ nNumberOfBytesToWrite+16(FP), R8
        MOVQ lpNumberOfBytesWritten+24(FP), R9
        MOVQ lpOverlapped+32(FP), R11

        // See NOTE[align-stack-x64].
	MOVQ SP, CX
	ANDQ $~15, SP
	MOVQ CX, 40(SP)

        // NOTE(strager): 40 bytes (5 parameters) at the bottom of the stack are
        // reserved for the callee.
	MOVQ ·func_WriteFile(SB), BX  // NOTE(strager): Might clobber CX.
        MOVQ R10, CX                 // hFile
        MOVQ R11, 32(SP)             // lpOverlapped
	CALL BX

        MOVQ 40(SP), CX
        MOVQ CX, SP

        MOVQ AX, ret+40(FP)
        RET

// NOTE[align-stack-x64]: The Windows x64 ABI [1] requires that the stack
// pointer is aligned to a 16-byte boundary. Go's amd64 ABI0 only guarantees
// that the stack pointer is aligned to an 8-byte boundary. When calling Windows
// APIs, we must align the stack pointer ourselves to comply with the Windows
// x64 ABI.
//
// [1] https://learn.microsoft.com/en-us/cpp/build/stack-usage?view=msvc-170#stack-allocation
