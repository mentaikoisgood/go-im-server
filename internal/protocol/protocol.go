package protocol

// Command constants 命令常量
const (
	CmdWho     = "who"     // 查詢在線用戶
	CmdRename  = "rename"  // 更改用戶名
	CmdTo      = "to"      // 發送私聊消息
	CmdMenu    = "menu"    // 顯示菜單
	CmdExit    = "exit"    // 退出（私聊模式）
	CmdPrivate = "private" // 進入私聊模式
)

// Message formats 消息格式
const (
	MsgFormatPrivate = "to=%s=%s"
	MsgFormatRename  = "rename=%s"
)

// Server responses 服務器響應
const (
	WelcomeMessage       = "WELCOME TO GO IM SERVER"
	OnlineFormat         = "[在線] %s"
	PrivateMsgFormat     = "[私聊] %s 說: %s"
	PrivateConfirmFormat = "[私聊] 您對 [%s] 說: %s"
	UserJoinedFormat     = "✅ [%s] 上線了"
	UserLeftFormat       = "❌ [%s] 離線了"
	UserRenamedFormat    = "公告：[%s] 改名為 [%s]"
	RenameSuccessFormat  = "成功：您的名稱已更新為 [%s]"
	TimeoutMessage       = "❌ 你已經超時了，將被踢出"
)

// Error messages 錯誤消息
const (
	ErrNameEmpty     = "錯誤：新名稱不能為空"
	ErrNameTaken     = "錯誤：名稱 [%s] 已經被使用"
	ErrUserOffline   = "錯誤：用戶 [%s] 不在線"
	ErrPrivateFormat = "錯誤：私聊格式錯誤，請使用 to=用戶名=消息"
)
