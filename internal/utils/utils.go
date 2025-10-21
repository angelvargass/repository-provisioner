package utils //nolint

import (
	"crypto/rand"
	"log/slog"
	"math/big"
	"os"
)

func HandleError(errorMessage string, err error) {
	if err != nil {
		slog.Error(errorMessage, slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func GenerateRandomInteger() (int64, error) {
	maxNum := big.NewInt(100)
	nBig, err := rand.Int(rand.Reader, maxNum)
	if err != nil {
		return 0, err
	}
	return nBig.Int64(), nil
}
