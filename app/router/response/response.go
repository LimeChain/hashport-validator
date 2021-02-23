package response

type ErrResponse struct {
	Err error `json:"-"` // low-level runtime error

	ErrorMessage string `json:"error,omitempty"` // application-level error message, for debugging
}

func ErrorResponse(err error) *ErrResponse {
	return &ErrResponse{
		Err:          err,
		ErrorMessage: err.Error(),
	}
}

type MetadataResponse struct {
	TransactionFee         string `json:"txFee"`
	TransactionFeeCurrency string `json:"txFeeCurrency"`
	GasPriceGwei           string `json:"gasPriceGwei"`
}
