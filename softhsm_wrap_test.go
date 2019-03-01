package pdfmix

import (
	"testing"

	"github.com/pkg/errors"
)

var hsmLibPath = "/usr/local/Cellar/softhsm/2.3.0/lib/softhsm/libsofthsm2.so"
var hsmPin = "1234"
var hsmTokenLabel = "tpb"
var hsmTokenLabelPriv = "tpb"
var hsmCKAID = []int{0xd, 0xd, 0xd, 0xd, 0xd, 0xd, 0xd, 0xd} //ccccccc

func _TestSoftHsmWrap(t *testing.T) {
	err := testSoftHsmWrap()
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func _TestToHsmByte(t *testing.T) {
	id := []int{0xc, 0xc, 0xc, 0xc, 0xc, 0xc, 0xc}
	result := ToHsmByte(id)
	if len(result) != 4 {
		t.Errorf("result must be 4!!!")
	}
	if !(result[0] == 12 && result[1] == 204 && result[2] == 204 && result[3] == 204) {
		t.Errorf("convert Wrong!!")
	}
}

func _TestToHsmByte2(t *testing.T) {
	id := []int{0xd, 0xd, 0xd, 0xd, 0xd, 0xd, 0xd, 0xd}
	result := ToHsmByte(id)
	if len(result) != 4 {
		t.Errorf("result must be 4!!!")
	}
	if !(result[0] == 221 && result[1] == 221 && result[2] == 221 && result[3] == 221) {
		t.Errorf("convert Wrong!!")
	}
}

func testSoftHsmWrap() error {
	sw := SoftHsmWrap{
		LibraryPath: hsmLibPath,
		PIN:         hsmPin,
		CertLabel:   hsmTokenLabel,
		PrivLabel:   hsmTokenLabelPriv,
		SlotIndex:   0,
	}
	//fmt.Printf("%+v\n", sw)
	err := sw.Open()
	if err != nil {
		return errors.Wrap(err, "")
	}
	defer sw.Close()
	/*_, err = sw.GetSigningCertificate()
	if err != nil {
		return errors.Wrap(err, "")
	}*/
	return nil
}
