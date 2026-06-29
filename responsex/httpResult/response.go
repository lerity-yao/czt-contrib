package httpResult

import (
	"errors"
	"net/http"

	"github.com/lerity-yao/czt-contrib/responsex/xerr"
	"google.golang.org/grpc/status"
)

const (
	LogFType       = "xr_type"
	LogFResult     = "xr_result"
	LogFStack      = "xr_stack"
	LogFAlertLevel = "xr_alert_level"
	LogFAlertData  = "xr_alert_data"

	LogApiSuccess = "API-SUCCESS"
	LogApiError   = "API-ERROR"
	LogRpcSuccess = "RPC-SUCCESS"
	LogRpcError   = "RPC-ERROR"
)

type (
	ResponseSuccessBean struct {
		Code    uint32      `json:"code"`    // 业务状态码
		Msg     string      `json:"msg"`     // 业务消息
		Data    interface{} `json:"data"`    // 返回数据
		TraceId string      `json:"traceId"` // 链路跟踪traceId
		SpanId  string      `json:"spanId"`  // 链路跟踪spanId
	}

	NullJson struct{}

	ResponseErrorBean struct {
		Code    uint32 `json:"code"`
		Msg     string `json:"msg"`
		TraceId string `json:"traceId"` // 链路跟踪traceId
		SpanId  string `json:"spanId"`  // 链路跟踪spanId
	}
)

type (
	grpcStatus interface {
		GRPCStatus() *status.Status
	}

	// ErrData 错误解析结果，拦截器可读写
	ErrData struct {
		Code       uint32         `json:"code"`
		Msg        string         `json:"msg"`
		AlertLevel string         `json:"alertLevel"`
		AlertData  map[string]any `json:"alertData"`
	}

	// ErrInterceptor 错误拦截器，在错误解析后、响应写入前执行
	// 可用于告警通知、指标采集、错误码映射等
	ErrInterceptor func(r *http.Request, err error, data *ErrData)
)

// Success 请求成功返回数据, traceId 为链路跟踪traceId, spanId为链路跟踪spanId
func Success(traceId, spanId string, data interface{}) *ResponseSuccessBean {
	return &ResponseSuccessBean{
		Code:    uint32(xerr.OK),
		Msg:     xerr.OK.Msg(),
		Data:    data,
		TraceId: traceId,
		SpanId:  spanId,
	}
}

// Error 请求失败返回数据, traceId 为链路跟踪traceId, spanId为链路跟踪spanId
func Error(traceId, spanId string, errCode uint32, errMsg string) *ResponseErrorBean {
	return &ResponseErrorBean{
		Code:    errCode,
		Msg:     errMsg,
		TraceId: traceId,
		SpanId:  spanId,
	}
}

// ParseErr 从错误链中解析错误码和消息，使用 errors.As 穿透查找
func ParseErr(err error) ErrData {
	data := ErrData{
		Code: uint32(xerr.ServerCommonError),
		Msg:  xerr.ServerCommonError.Msg(),
	}
	var ce *xerr.CodeError
	if errors.As(err, &ce) {
		data.Code = uint32(ce.GetCode())
		data.Msg = ce.GetMessage()
		data.AlertLevel = ce.GetAlertLevel()
		data.AlertData = ce.GetAlertData()
	} else {
		var gs grpcStatus
		if errors.As(err, &gs) {
			s := gs.GRPCStatus()
			if s.Message() != "" {
				data.Msg = s.Message()
			}
			if s.Code() != 0 {
				data.Code = uint32(s.Code())
			}
		}
	}
	return data
}
