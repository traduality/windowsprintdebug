// This file should not be included in the Go package.
// clang-format off
//go:build (windows && !windows)
// clang-format on

#include <Windows.h>
#include <stdint.h>
#include <stdio.h>
#include <string>
#include <vector>

static bool verbose = false;

// https://protobuf.dev/programming-guides/encoding/#structure
enum Protobuf_Type {
  // VARINT (int32, int64, uint32, uint64, sint32, sint64, bool, enum)
  protobuf_type_varint = 0,
  // LEN (string, bytes, embedded messages, packed repeated fields)
  protobuf_type_len = 2,
};

struct Protobuf_Message {
  std::vector<uint8_t> bytes;

  void write_varint(uint64_t value) {
    while (value >= 128) {
      this->bytes.push_back((value & 0x7f) | 0x80);
      value >>= 7;
    }
    this->bytes.push_back(value & 0xff);
  }

  void write_tag(uint32_t field_number, Protobuf_Type wire_type) {
    uint64_t tag = ((uint64_t)field_number << 3) | (uint64_t)wire_type;
    this->write_varint(tag);
  }

  void write_raw(void* data, size_t data_size) {
    size_t old_size = this->bytes.size();
    this->bytes.resize(old_size + data_size);
    memcpy(&this->bytes[old_size], data, data_size);
  }

  void write_embedded_message(const Protobuf_Message* embedded) {
    this->write_varint(embedded->bytes.size());
    this->bytes.insert(this->bytes.end(), embedded->bytes.begin(),
                       embedded->bytes.end());
  }
};

static void write_protocol_raw(void* data, size_t data_size) {
  DWORD bytes_written;
  BOOL ok = WriteFile(GetStdHandle(STD_OUTPUT_HANDLE), data, data_size,
                      &bytes_written, /*lpOverlapped:*/ NULL);
  if (!ok) {
    fprintf(stderr, "error: WriteFile to protocol failed: %d\n",
            GetLastError());
    exit(1);
  }
  if (bytes_written != data_size) {
    fprintf(stderr,
            "error: incomplete write to protocol (tried to write %zu bytes, "
            "but wrote %u bytes)\n",
            data_size, bytes_written);
    exit(1);
  }
}

int wmain(int argc, wchar_t** argvw) {
  BOOL ok;

  setvbuf(stderr, NULL, _IONBF, 0);

  std::wstring command_line = L"";
  for (int i = 1; i < argc; i += 1) {
    if (!command_line.empty()) {
      command_line.append(L" ");
    }
    command_line.append(argvw[i]);
  }
  STARTUPINFOEXW startup_info = {0};
  startup_info.StartupInfo.cb = sizeof(startup_info);
  // TODO(strager): Test with STARTF_USESTDHANDLES.
  PROCESS_INFORMATION process_info;
  ok = CreateProcessW(/*lpApplicationName:*/ NULL, command_line.data(),
                      /*lpProcessAttributes:*/ NULL,
                      /*lpThreadAttributes:*/ NULL,
                      /*bInheritHandles:*/ false,
                      DEBUG_ONLY_THIS_PROCESS | EXTENDED_STARTUPINFO_PRESENT,
                      /*lpEnvironment:*/ NULL,
                      /*lpCurrentDirectory:*/ NULL,
                      (STARTUPINFOW*)&startup_info, &process_info);
  if (!ok) {
    fprintf(stderr, "CreateProcess failed: %d\n", GetLastError());
    exit(1);
  }
  CloseHandle(process_info.hThread);

  bool keep_debugging = true;
  while (keep_debugging) {
    DEBUG_EVENT event;
    WaitForDebugEvent(&event, INFINITE);

    DWORD continue_status;
    switch (event.dwDebugEventCode) {
      case EXCEPTION_DEBUG_EVENT:
        if (verbose) {
          fprintf(stderr, "note: EXCEPTION_DEBUG_EVENT\n");
        }
        continue_status = DBG_EXCEPTION_NOT_HANDLED;
        break;

      case EXIT_PROCESS_DEBUG_EVENT:
        if (verbose) {
          fprintf(stderr, "note: EXIT_PROCESS_DEBUG_EVENT\n");
        }
        // Call ContinueDebugEvent, then exit the loop.
        keep_debugging = false;
        break;

      case OUTPUT_DEBUG_STRING_EVENT: {
        if (verbose) {
          fprintf(stderr, "note: OUTPUT_DEBUG_STRING_EVENT\n");
        }
        static char buffer[65536];
        SIZE_T bytes_read = 0;
        BOOL ok = ReadProcessMemory(
            process_info.hProcess, event.u.DebugString.lpDebugStringData,
            buffer, event.u.DebugString.nDebugStringLength, &bytes_read);
        if (!ok) {
          fprintf(stderr, "error: ReadProcessMemory failed: %d\n",
                  GetLastError());
          exit(1);
        }
        if (bytes_read != event.u.DebugString.nDebugStringLength) {
          fprintf(stderr,
                  "error: ReadProcessMemory returned incomplete data\n");
          exit(1);
        }

        // ProtocolMessageOutputDebugString
        Protobuf_Message ods_message = Protobuf_Message();
        // ProtocolMessageOutputDebugString.data
        ods_message.write_tag(1, protobuf_type_len);
        ods_message.write_varint(event.u.DebugString.nDebugStringLength);
        ods_message.write_raw(buffer, event.u.DebugString.nDebugStringLength);
        // ProtocolMessageOutputDebugString.isUnicode
        ods_message.write_tag(2, protobuf_type_varint);
        ods_message.write_varint(event.u.DebugString.fUnicode == 0 ? 0 : 1);

        // ProtocolMessage
        Protobuf_Message protocol_message = Protobuf_Message();
        // ProtocolMessage.outputDebugString
        protocol_message.write_tag(1, protobuf_type_len);
        protocol_message.write_embedded_message(&ods_message);

        // Protocol.messages (incomplete)
        Protobuf_Message protocol_incomplete = Protobuf_Message();
        protocol_incomplete.write_tag(1, protobuf_type_len);
        protocol_incomplete.write_embedded_message(&protocol_message);
        write_protocol_raw(protocol_incomplete.bytes.data(),
                           protocol_incomplete.bytes.size());
        break;
      }

      case CREATE_PROCESS_DEBUG_EVENT:
      case CREATE_THREAD_DEBUG_EVENT:
      case EXIT_THREAD_DEBUG_EVENT:
      case LOAD_DLL_DEBUG_EVENT:
      case RIP_EVENT:
      case UNLOAD_DLL_DEBUG_EVENT:
      default:
        if (verbose) {
          fprintf(stderr, "note: other event\n");
        }
        continue_status = DBG_CONTINUE;
        break;
    }
    ContinueDebugEvent(event.dwProcessId, event.dwThreadId, continue_status);
  }

  if (verbose) {
    fprintf(stderr,
            "note: exited debugger loop; waiting for process to signal exit\n");
  }
  WaitForSingleObject(process_info.hProcess, INFINITE);
  DWORD exit_code;
  ok = GetExitCodeProcess(process_info.hProcess, &exit_code);
  if (!ok) {
    fprintf(stderr, "error: GetExitCodeProcess failed: %d\n", GetLastError());
    exit(1);
  }

  {
    // ProtocolProcessExit
    Protobuf_Message protocol_process_exit = Protobuf_Message();
    // ProtocolProcessExit.exitCode
    protocol_process_exit.write_tag(1, protobuf_type_varint);
    protocol_process_exit.write_varint(exit_code);

    // Protocol.processExit (incomplete)
    Protobuf_Message protocol_incomplete = Protobuf_Message();
    protocol_incomplete.write_tag(2, protobuf_type_len);
    protocol_incomplete.write_embedded_message(&protocol_process_exit);
    write_protocol_raw(protocol_incomplete.bytes.data(),
                       protocol_incomplete.bytes.size());
  }

  exit(0);
}
