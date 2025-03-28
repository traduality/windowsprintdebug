#define NOSPLIT	4

// // Size: 0x20
// type Closure struct {
// 	f unsafe.Pointer  // Offset 0x00: Pointer to debugWriteOnSystemStackTrampoline.
// 	fd uintptr        // Offset 0x08: Argument.
// 	p unsafe.Pointer  // Offset 0x10: Argument.
// 	n int32           // Offset 0x18: Argument.
// 	result int32      // Offset 0x1c: Return value of debugWriteOnSystemStackTrampoline.
// }

// func DebugWrite(fd uintptr, p unsafe.Pointer, n int32) int32
TEXT ·DebugWrite(SB),NOSPLIT,$40-20
        // Stack layout:
        // arg-0x28(SP) *Closure     // 0x08 bytes
        // closure-0x20(SP) Closure  // 0x20 bytes
#define TEMP_closure(offset) closure-(0x20-offset)(SP)
        MOVD $·debugWriteOnSystemStackTrampoline(SB), R0
        MOVD R0, TEMP_closure(0x00)  // closure.f
        MOVD fd+0(FP), R0
        MOVD R0, TEMP_closure(0x08)  // closure.fd
        MOVD p+8(FP), R0
        MOVD R0, TEMP_closure(0x10)  // closure.p
        MOVW n+16(FP), R0
        MOVW R0, TEMP_closure(0x18)  // closure.n

        MOVD $TEMP_closure(0), R0
        MOVD R0, arg-0x28(SP)        // f = &closure
	CALL runtime·systemstack(SB)

        MOVW TEMP_closure(0x1c), R0  // closure.result
        MOVW R0, ret+24(FP)
        RET
#undef TEMP_closure

// This is a closure function.
//
// func debugWriteOnSystemStackTrampoline()
TEXT ·debugWriteOnSystemStackTrampoline(SB),NOSPLIT,$32-0
        // NOTE(strager): We received a pointer to the closure in R26.
        // FIXME(strager): Is this ABI-stable or should we use another mechanism
        // for accessing the closure?
        MOVD R26, closure-0(SP)

        MOVD 0x08(R26), R0                 // closure.fd
        MOVD R0, arg_fd-32(SP)
        MOVD 0x10(R26), R0                 // closure.p
        MOVD R0, arg_p-24(SP)
        MOVW 0x18(R26), R0                 // closure.n
        MOVW R0, arg_n-16(SP)
        CALL ·debugWriteOnSystemStack(SB)  // Clobbers R26.
        MOVD closure-0(SP), R26
        MOVD R0, 0x1c(R26)                 // closure.bytesWritten
        RET

// func cIsDebuggerPresent() bool
TEXT ·cIsDebuggerPresent(SB),NOSPLIT,$64-8
	MOVD ·func_IsDebuggerPresent(SB), R0
	CALL R0
        MOVD R0, ret+0(FP)
        RET

// func cOutputDebugStringA(lpOutputString unsafe.Pointer)
TEXT ·cOutputDebugStringA(SB),NOSPLIT,$64-8
        MOVD lpOutputString+0(FP), R0
	MOVD ·func_OutputDebugStringA(SB), R1
	CALL R1
        RET

// func cWriteFile(
// 	hFile uintptr,
// 	lpBuffer unsafe.Pointer,
// 	nNumberOfBytesToWrite uint32,
// 	lpNumberOfBytesWritten *uint32,
// 	lpOverlapped unsafe.Pointer,
// ) bool
TEXT ·cWriteFile(SB),NOSPLIT,$0-48
        MOVD hFile+0(FP), R0
        MOVD lpBuffer+8(FP), R1
        MOVD nNumberOfBytesToWrite+16(FP), R2
        MOVD lpNumberOfBytesWritten+24(FP), R3
        MOVD lpOverlapped+32(FP), R4
	MOVD ·func_WriteFile(SB), R5
	CALL R5
        MOVD R0, ret+40(FP)
        RET

// NOTE[align-stack-ARM64]: The Windows ARM64 ABI [1] requires that the stack
// pointer is aligned to a 16-byte boundary. Go's arm64 ABI0 also makes this
// guarantee. Therefore, unlike on x64, we do not need to align the stack
// pointer ourselves to comply with the Windows ARM64 ABI.
//
// [1] https://learn.microsoft.com/en-us/cpp/build/arm64-windows-abi-conventions?view=msvc-170#stack
