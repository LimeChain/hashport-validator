package recovery

type RecoveryService struct {
}

func NewRecoveryService() *RecoveryService {
	return &RecoveryService{}
}

func Recover() error {
	_, err := cryptoTransferWatcherRecovery()
	if err != nil {
		return err
	}

	_, err = cryptoTransferHandlerRecovery()
	if err != nil {
		return err
	}

	_, err = consensusMessageWatcherRecovery()
	if err != nil {
		return err
	}

	_, err = consensusMessageHandlerRecovery()
	if err != nil {
		return err
	}
	return nil
}

func cryptoTransferWatcherRecovery() (interface{}, error) {
	return nil, nil
}

func cryptoTransferHandlerRecovery() (interface{}, error) {
	return nil, nil
}

func consensusMessageWatcherRecovery() (interface{}, error) {
	return nil, nil
}

func consensusMessageHandlerRecovery() (interface{}, error) {
	return nil, nil
}
