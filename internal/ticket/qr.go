package ticket

import "github.com/skip2/go-qrcode"

func QRCode(code string) ([]byte, error) {
	return qrcode.Encode(code, qrcode.Medium, 256)
}
