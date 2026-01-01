package apperr

const sentinel_err_operation = "sentinel error"

var (
	// 汎用・システム
	ErrInternal = newSentinelError("internal error", "internal error") // 内部エラー

	// 認証・認可
	ErrUnauthorized = newSentinelError("unauthorized", "unauthorized") // 認証エラー
	ErrForbidden    = newSentinelError("forbidden", "forbidden")       // アクセス拒否

	// データ操作
	ErrNotFound   = newSentinelError("not found", "not found")               // リソースが見つからない
	ErrConflict   = newSentinelError("conflict", "conflict")                 // リソースの競合
	ErrValidation = newSentinelError("validation error", "validation error") // バリデーションエラー
)
