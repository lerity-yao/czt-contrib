package xerr

import (
	"fmt"
	"runtime"
	"sync"
)

// Code 错误码枚举类型
type (
	Code uint32

	alertLevel string // alertLevel 不导出，外部只能使用预定义常量 AlertP0~AlertP3

	CodeError struct {
		code       Code
		message    string
		alertLevel alertLevel
		alertData  map[string]any
	}

	CodeErrorOptions func(c *CodeError)
)

type wrapError struct {
	msg    string
	err    error
	frames []runtime.Frame
}

const (
	OK                        Code = 0      // OK 成功返回 SUCCESS
	ServerCommonError         Code = 100001 // ServerCommonError 服务器开小差啦,稍后再来试一试
	RequestParamError         Code = 100002 // RequestParamError 参数错误
	TokenExpireError          Code = 100003 // TokenExpireError token失效，请重新登陆
	TokenGenerateError        Code = 100004 // TokenGenerateError 生成token失败
	DbError                   Code = 100005 // DbError 数据库繁忙,请稍后再试
	DbUpdateAffectedZeroError Code = 100006 // DbUpdateAffectedZeroError 更新数据影响行数为0
	MdCommonError             Code = 100007 // MdCommonError 中间件错误
	PermitNoAccess            Code = 100008 // PermitNoAccess 无权限操作
	SignParamError            Code = 100009 // SignParamError 签名错误

	AlertP0 alertLevel = "P0" // AlertP0 最高告警级别
	AlertP1 alertLevel = "P1" // AlertP1 高告警级别
	AlertP2 alertLevel = "P2" // AlertP2 中告警级别
	AlertP3 alertLevel = "P3" // AlertP3 低告警级别
)

var messages = map[Code]string{
	OK:                        "SUCCESS",
	ServerCommonError:         "服务器开小差啦,稍后再来试一试",
	RequestParamError:         "参数错误",
	TokenExpireError:          "token失效，请重新登陆",
	TokenGenerateError:        "生成token失败",
	DbError:                   "数据库繁忙,请稍后再试",
	DbUpdateAffectedZeroError: "更新数据影响行数为0",
	MdCommonError:             "中间件错误",
	PermitNoAccess:            "无权限操作",
	SignParamError:            "签名错误",
}

// Msg 返回错误码对应的描述信息，未匹配则返回默认兜底消息
func (c Code) Msg() string {
	if msg, ok := messages[c]; ok {
		return msg
	}
	return messages[ServerCommonError]
}

// Error 实现 error 接口
func (e *CodeError) Error() string {
	return fmt.Sprintf("ErrCode: %d, ErrMsg: %s", e.code, e.message)
}

// Is 实现 errors.Is 穿透匹配，判断 target 是否为 CodeError
func (e *CodeError) Is(target error) bool {
	_, ok := target.(*CodeError)
	return ok
}

func (e *CodeError) GetCode() Code                { return e.code }
func (e *CodeError) GetMessage() string           { return e.message }
func (e *CodeError) GetAlertLevel() string        { return string(e.alertLevel) }
func (e *CodeError) GetAlertData() map[string]any { return e.alertData }

// NewCodeError 构建新的 CodeError，允许自定义 code 和 message，支持 WithAlertLevel/WithAlertData 选项
func NewCodeError(code Code, message string, opts ...CodeErrorOptions) *CodeError {
	e := &CodeError{code: code, message: message}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// NewErrCodeMsg 构建新的 CodeError，允许自定义 errCode 和 errMsg
func NewErrCodeMsg(errCode uint32, errMsg string, opts ...CodeErrorOptions) *CodeError {
	return NewCodeError(Code(errCode), errMsg, opts...)
}

// NewErrCode 用自定义的 errCode 查找预设消息，未匹配则返回兜底消息
func NewErrCode(errCode uint32, opts ...CodeErrorOptions) *CodeError {
	return NewCodeError(Code(errCode), Code(errCode).Msg(), opts...)
}

// NewErrMsg 固定 errCode 为 ServerCommonError，自定义 errMsg
func NewErrMsg(errMsg string, opts ...CodeErrorOptions) *CodeError {
	return NewCodeError(ServerCommonError, errMsg, opts...)
}

// WithAlertLevel 设置报警级别
func WithAlertLevel(level alertLevel) CodeErrorOptions {
	return func(e *CodeError) { e.alertLevel = level }
}

// WithAlertData 设置报警数据
func WithAlertData(key string, value any) CodeErrorOptions {
	return func(e *CodeError) {
		if e.alertData == nil {
			e.alertData = make(map[string]any)
		}
		e.alertData[key] = value
	}
}

// wrapError 携带调用堆栈的错误包装，实现 error / Unwrap / Format 三接口
func (e *wrapError) Error() string { return e.msg }
func (e *wrapError) Unwrap() error { return e.err }

// Format 实现 fmt.Formatter：fmt.Printf("%+v", err) 输出消息+堆栈
func (e *wrapError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n", e.msg)
			for _, f := range e.stackTrace() {
				fmt.Fprintf(s, "    %s\n        %s:%d\n", f.Function, f.File, f.Line)
			}
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, e.msg)
	case 'q':
		fmt.Fprintf(s, "%q", e.msg)
	}
}

func (e *wrapError) stackTrace() []runtime.Frame {
	return e.frames
}

// callerBufPool 复用 []uintptr 缓冲区，避免 Wrapf 每次分配 256 字节
var callerBufPool = sync.Pool{
	New: func() any {
		buf := make([]uintptr, 32)
		return &buf
	},
}

// Wrapf 用格式化消息包装一个错误，类似 errors.Wrapf，格式为 "格式化消息: 原始错误"
// 支持 format 中使用多个 %w，可以将额外错误加入链中（如 asynq.SkipRetry）
// 额外捕获调用堆栈，可通过 fmt.Printf("%+v") 获取
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	pcsPtr := callerBufPool.Get().(*[]uintptr)
	n := runtime.Callers(2, *pcsPtr) // skip: 0=Callers 1=Wrapf 2=调用方
	frames := make([]runtime.Frame, 0, n)
	iter := runtime.CallersFrames((*pcsPtr)[:n])
	for {
		f, more := iter.Next()
		frames = append(frames, f)
		if !more {
			break
		}
	}
	callerBufPool.Put(pcsPtr)
	wrapped := fmt.Errorf(format+": %w", append(args, err)...)
	return &wrapError{
		msg:    wrapped.Error(),
		err:    wrapped,
		frames: frames,
	}
}
