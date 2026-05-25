package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateReferenceNo() string {
	return fmt.Sprintf("FLP-%d-%06d", time.Now().Year(), rand.Intn(1000000))
}

func GenerateVANumber() string {
	return fmt.Sprintf("8808%d", 100000000+rand.Intn(900000000))
}

func GenerateQRISString(referenceNo string, amount int64) string {
	return fmt.Sprintf("QRIS|FLIPAY|%s|%d", referenceNo, amount)
}
