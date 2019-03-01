package pdfmix

import (
	"fmt"

	"github.com/oneplus1000/pkcs11"
	"github.com/pkg/errors"
)

//SoftHsmWrap softshm wrap
type SoftHsmWrap struct {
	//private
	ctx                *pkcs11.Ctx
	session            pkcs11.SessionHandle
	isSession, isLogin bool
	//public
	LibraryPath string
	CertLabel   string //label ของ cert
	PrivLabel   string //lavel ของ private key
	PIN         string //pin
	SlotIndex   int
}

//Ctx hsm ctx
func (sw *SoftHsmWrap) Ctx() *pkcs11.Ctx {
	return sw.ctx
}

//Session hsm session
func (sw *SoftHsmWrap) Session() pkcs11.SessionHandle {
	return sw.session
}

//Open session
func (sw *SoftHsmWrap) Open() error {
	//fmt.Printf("sw.LibraryPath=%s\n", sw.LibraryPath)
	ctx := pkcs11.New(sw.LibraryPath)
	err := ctx.Initialize()
	if err != nil {
		return errors.Wrap(err, "p.Initialize() fail")
	}
	sw.ctx = ctx
	slots, err := sw.ctx.GetSlotList(true)
	if err != nil {
		return errors.Wrap(err, "p.GetSlotList() fail")
	}

	session, err := sw.ctx.OpenSession(slots[sw.SlotIndex], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		return errors.Wrap(err, "p.OpenSession() fail")
	}
	fmt.Printf("slots =%+v  %d  %s\n", slots, sw.SlotIndex, sw.PIN)
	sw.isSession = true
	sw.session = session
	err = sw.ctx.Login(session, pkcs11.CKU_USER, sw.PIN)
	if err != nil {
		return errors.Wrap(err, "p.Login() fail")
	}
	sw.isLogin = true

	return nil
}

//Close close ctx
func (sw *SoftHsmWrap) Close() {
	if sw.isLogin {
		sw.ctx.Logout(sw.session)
	}
	if sw.isSession {
		sw.ctx.CloseSession(sw.session)
	}
	if sw.ctx != nil {
		sw.ctx.Finalize()
		sw.ctx.Destroy()
	}
}

//GetPrivateKey private key
func (sw *SoftHsmWrap) GetPrivateKey() (pkcs11.ObjectHandle, error) {

	searchTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_RSA),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, sw.PrivLabel),
	}
	err := sw.ctx.FindObjectsInit(sw.session, searchTemplate)
	if err != nil {
		var oh pkcs11.ObjectHandle
		return oh, errors.Wrap(err, "")
	}
	foundObjs, _, err := sw.ctx.FindObjects(sw.session, 10)
	if err != nil {
		var oh pkcs11.ObjectHandle
		return oh, errors.Wrap(err, "")
	}
	err = sw.ctx.FindObjectsFinal(sw.session) // found nothing
	if err != nil {
		var oh pkcs11.ObjectHandle
		return oh, errors.Wrap(err, "")
	}

	return foundObjs[0], nil
}

//GetSigningCertificate signing certificate
func (sw *SoftHsmWrap) GetSigningCertificate() ([]byte, error) {

	//fmt.Printf("sw.CKAID=%+v\n", sw.CKAID)
	searchTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_CERTIFICATE),
		pkcs11.NewAttribute(pkcs11.CKA_CERTIFICATE_TYPE, pkcs11.CKC_X_509),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, sw.CertLabel),
	}
	err := sw.ctx.FindObjectsInit(sw.session, searchTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	foundObjs, _, err := sw.ctx.FindObjects(sw.session, 4)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	err = sw.ctx.FindObjectsFinal(sw.session) // found nothing
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	if len(foundObjs) != 1 {
		return nil, fmt.Errorf("found object != 1 (found = %d)", len(foundObjs))
	}

	var ckaVal []byte
	for _, f := range foundObjs {
		template := []*pkcs11.Attribute{
			pkcs11.NewAttribute(pkcs11.CKA_VALUE, nil),
		}
		attr, _ := sw.ctx.GetAttributeValue(sw.session, f, template)
		for _, a := range attr {
			//fmt.Printf("a=%+v\n", a)
			ckaVal = a.Value
			break
		}
	}
	return ckaVal, nil
}

//ToHsmByte form slice of int to hsm byte ส่วนใหญ่ใช้ใน id
func ToHsmByte(id []int) []byte {
	var arr []byte
	j := len(id) - 1
	for {
		last := id[j]
		j--
		first := 0
		if j >= 0 {
			first = id[j] << 4
		}
		j--
		//fmt.Printf("j=%d  %d %d \n", j, first, last)
		comb := first + last
		arr = append(arr, byte(comb))
		if j < 0 {
			break
		}
	}
	//fmt.Printf("-->%+v\n", arr)
	max := len(arr)
	result := make([]byte, max)
	for i := 0; i < max; i++ {
		result[max-i-1] = arr[i]
	}
	return result
}
