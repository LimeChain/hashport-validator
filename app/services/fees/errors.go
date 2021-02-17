package fees

import "errors"

var (
	InvalidTransferAmount = errors.New("INVALID_TRANSFER_AMOUNT")
	InvalidTransferFee    = errors.New("INVALID_TRANSFER_FEE")
	InvalidGasPrice       = errors.New("INVALID_GAS_PRICE")
	InsufficientFee       = errors.New("INSUFFICIENT_FEE")
	RateProviderFailure   = errors.New("RATE_PROVIDER_FAILURE")
)
