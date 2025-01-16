package constant

const (
	HeaderContentType      = "Content-Type"
	HeaderDisposition      = "Content-Disposition"
	HeaderAuthorization    = "Authorization"
	HeaderTransferEncoding = "Content-Transfer-Encoding"

	MineTypeJson           = "application/json"
	MineTypeZip            = "application/zip"
	MineTypeOctetStream    = "application/octet-stream"
	MineTypeFormUrlencoded = "application/x-www-form-urlencoded"
	MineTypeTextPlain      = "text/plain"
	MineBinary             = "binary"
)

type Response struct {
	Code       int         `json:"code"`
	Data       interface{} `json:"data"`
	ErrMessage string      `json:"err_message"`
}
