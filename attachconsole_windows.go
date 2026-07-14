//go:build windows

// rain-qa-func — QA 测试工具集
// attachconsole_windows.go: 修复 GUI 子系统（-H windowsgui）下 CLI 输出不可见的问题。
//
// Go module: git.devcloud.ztgame.com/v-tangfangda/rain-qa-func
package main

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ATTACH_PARENT_PROCESS 是 Windows AttachConsole API 的特殊参数值 (DWORD)-1 = 0xFFFFFFFF，
// 含义为「attach 调用进程的父进程所附加的控制台」。
// golang.org/x/sys/windows 未导出此常量，这里手动定义。
const ATTACH_PARENT_PROCESS = ^uint32(0)

// WriteConsoleInput 的输入事件类型常量。
const (
	keyEventType = 0x0001 // KEY_EVENT
	vkReturn     = 0x0D   // VK_RETURN，回车键
)

// consoleAttached 记录当前进程是否成功 attach 了父控制台，
// 供 releaseParentConsole / nudgeConsoleRefresh 判断是否需要清理/刷新。
var consoleAttached bool

// keyEventRecord 对应 Windows API 的 KEY_EVENT_RECORD（16 字节），
// 用于 WriteConsoleInput 构造键盘事件。字段对齐必须与 Windows 结构一致。
type keyEventRecord struct {
	KeyDown         uint32 // BOOL：TRUE=按下，FALSE=释放
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	Char            uint16 // WCHAR：Unicode 字符（回车='\r'）
	ControlKeyState uint32 // Ctrl/Alt/Shift 等状态（单键回车无需设置）
}

// inputRecord 对应 Windows API 的 INPUT_RECORD（20 字节）。
// EventType=keyEventType(1) 时使用 KeyEvent 字段，其余事件类型此处用不到。
type inputRecord struct {
	EventType uint16
	KeyEvent  keyEventRecord
}

// attachParentConsole 在 GUI 子系统程序带命令行参数启动时，attach 父进程的控制台，
// 并把 os.Stdout/os.Stderr 重新指向控制台屏幕缓冲区，使 cobra 的 CLI 输出可见。
//
// 背景：rain-qa-func.exe 用 -ldflags="-H windowsgui" 构建（PE Subsystem=GUI），
// Windows 加载器启动 GUI 子系统程序时不会附加父进程的控制台，导致在 cmd/PowerShell
// 里带参数运行（如 --help、proto-test 子命令）时，cobra 写到 os.Stdout/os.Stderr 的
// 输出全部丢失（GetStdHandle 返回无效句柄），表现为「命令敲下去没反应」。
//
// 仅在带命令行参数时（CLI 模式）attach：
//   - 无参数（GUI 模式）直接返回，避免 AttachConsole 干扰 Wails/WebView2 启动导致 GUI 闪退，
//     且 GUI 模式本就不需要 CLI 输出。
//
// 双击启动场景（父进程=explorer.exe，无控制台）下 AttachConsole 会失败，本函数直接返回，
// 不影响 runWails() 启动 GUI，且双击不会弹出 cmd 黑框。
//
// 实现要点：AttachConsole 成功后进程的标准句柄仍是启动时的（无效）句柄，
// 必须重新 CreateFile("CONOUT$") 打开控制台屏幕缓冲区，再 SetStdHandle 设为新句柄、
// 重建 os.Stdout/os.Stderr，否则 cobra 仍会写到旧的无效 fd。
//
// 副作用与对策：
//   - GUI 子系统进程 AttachConsole 共享父控制台后，cmd.exe 会因控制台被子进程持有而等待，
//     导致 CLI 输出后提示符不刷新。需在退出前 releaseParentConsole() 释放 attachment。
//   - PowerShell 的 PSReadLine 因 CONOUT$ 直写绕过管道，屏幕状态模型与实际不同步，
//     exe 退出后不重绘提示符。需在 FreeConsole 前 nudgeConsoleRefresh() 写假回车触发重绘。
func attachParentConsole() {
	// 仅 CLI 模式（带参数）才 attach；无参数 GUI 模式直接返回，避免干扰 Wails 启动
	if len(os.Args) <= 1 {
		return
	}
	// golang.org/x/sys/windows 未封装 AttachConsole，直接通过 kernel32.dll 调用
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	attachConsole := kernel32.NewProc("AttachConsole")

	// 返回值是 BOOL：非 0 成功，0 失败（父进程无控制台时返回 0）
	r, _, _ := attachConsole.Call(uintptr(ATTACH_PARENT_PROCESS))
	if r == 0 {
		// 父进程无控制台（双击从 explorer.exe 启动），失败是预期行为，跳过
		return
	}

	// 重新打开 CONOUT$（控制台屏幕缓冲区）作为新的 stdout/stderr 句柄
	conout, err := windows.UTF16PtrFromString("CONOUT$")
	if err != nil {
		return
	}
	h, err := windows.CreateFile(
		conout,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0, 0,
	)
	if err != nil {
		return
	}
	// stdout 与 stderr 共用同一个控制台屏幕缓冲区句柄
	windows.SetStdHandle(windows.STD_OUTPUT_HANDLE, h)
	windows.SetStdHandle(windows.STD_ERROR_HANDLE, h)
	os.Stdout = os.NewFile(uintptr(h), "/dev/stdout")
	os.Stderr = os.NewFile(uintptr(h), "/dev/stderr")
	consoleAttached = true
}

// releaseParentConsole 在程序退出前释放对父控制台的 attachment（FreeConsole）。
//
// GUI 子系统进程 AttachConsole 共享父进程（cmd/PowerShell）的控制台后，cmd.exe 会因
// 控制台被子进程持有而等待。退出前 FreeConsole 让 cmd.exe 检测到子进程已脱离、正常刷新提示符。
// 仅在已 attach 成功时生效；非 Windows 平台是空操作（见 attachconsole_other.go）。
func releaseParentConsole() {
	if !consoleAttached {
		return
	}
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	freeConsole := kernel32.NewProc("FreeConsole")
	freeConsole.Call()
	consoleAttached = false
}

// nudgeConsoleRefresh 往控制台输入缓冲区写一个「回车键」事件，
// 触发 PowerShell 的 PSReadLine 重新同步屏幕状态并重绘提示符。
//
// 背景：GUI 子系统 exe AttachConsole 直接写 CONOUT$ 后，PSReadLine 的屏幕状态模型
// 与实际屏幕不同步，exe 退出后它不重绘提示符（用户看到光标闪烁、误以为程序卡住，
// 实际 PowerShell 已回到命令行在等下一条命令）。往 CONIN$ 输入缓冲区塞一个假回车键事件，
// PSReadLine 下次 ReadConsole 读到后当作用户按回车 → 执行当前空命令行 → 渲染新提示符。
//
// 副作用：cmd.exe 也会收到此回车（执行空命令、多一个空提示符行，无害）。
// 必须在 releaseParentConsole（FreeConsole）之前调用，否则进程已脱离控制台、CONIN$ 无法访问。
// 由 main 通过 defer（LIFO，先于 releaseParentConsole 执行）触发。
func nudgeConsoleRefresh() {
	if !consoleAttached {
		return
	}
	conin, err := windows.UTF16PtrFromString("CONIN$")
	if err != nil {
		return
	}
	h, err := windows.CreateFile(
		conin,
		windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0, 0,
	)
	if err != nil {
		return
	}
	defer windows.CloseHandle(h)

	// 一次按键 = keydown + keyup 两个 INPUT_RECORD
	records := [2]inputRecord{
		{
			EventType: keyEventType,
			KeyEvent: keyEventRecord{
				KeyDown:        1,
				RepeatCount:    1,
				VirtualKeyCode: vkReturn,
				Char:           '\r',
			},
		},
		{
			EventType: keyEventType,
			KeyEvent: keyEventRecord{
				KeyDown:        0,
				RepeatCount:    1,
				VirtualKeyCode: vkReturn,
				Char:           '\r',
			},
		},
	}

	// golang.org/x/sys/windows 未封装 WriteConsoleInput，直接通过 kernel32.dll 调用
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	writeConsoleInput := kernel32.NewProc("WriteConsoleInputW")
	var written uint32
	_, _, _ = writeConsoleInput.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&records[0])),
		uintptr(len(records)),
		uintptr(unsafe.Pointer(&written)),
	)
}
