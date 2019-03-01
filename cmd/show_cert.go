package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/oneplus1000/pkcs11"
	"github.com/pkg/errors"
)

type ShowCert struct {
	HsmLib string //hsm path
	Label  string //cert label
	ID     string
}

func (s ShowCert) Exec() error {
	err := s.validate()
	if err != nil {
		return errors.Wrap(err, "validate fail")
	}
	err = s.exec()
	if err != nil {
		return errors.Wrap(err, "s.exec() fail")
	}
	return nil
}

func (s ShowCert) exec() error {
	ctx := pkcs11.New(s.HsmLib)
	err := ctx.Initialize()
	if err != nil {
		return errors.Wrap(err, "ctx.Initialize() fail")
	}
	defer ctx.Destroy()

	slots, err := ctx.GetSlotList(true)
	if err != nil {
		return errors.Wrap(err, "ctx.GetSlotList(true) fail")
	}
	session, err := ctx.OpenSession(slots[0], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		return errors.Wrap(err, "ctx.OpenSession() fail")
	}
	defer ctx.CloseAllSessions(slots[0])

	err = ctx.Login(session, pkcs11.CKU_USER, "1234")
	if err != nil {
		return errors.Wrap(err, "ctx.Login() fail")
	}
	defer ctx.Logout(session)

	tmpls := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_CERTIFICATE),
		pkcs11.NewAttribute(pkcs11.CKA_CERTIFICATE_TYPE, pkcs11.CKC_X_509),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, s.Label),
	}

	err = ctx.FindObjectsInit(session, tmpls)
	if err != nil {
		return errors.Wrap(err, "ctx.FindObjectsInit() fail")
	}

	foundObjs, _, err := ctx.FindObjects(session, 4)
	if err != nil {
		return errors.Wrap(err, "ctx.FindObjects() fail")
	}

	err = ctx.FindObjectsFinal(session) // found nothing
	if err != nil {
		return errors.Wrap(err, "ctx.FindObjectsFinal() fail")
	}

	fmt.Printf("found %d object(s)\n", len(foundObjs))

	for _, f := range foundObjs {
		template := []*pkcs11.Attribute{
			pkcs11.NewAttribute(pkcs11.CKA_VALUE, nil),
		}
		attr, _ := ctx.GetAttributeValue(session, f, template)
		for _, a := range attr {
			fmt.Printf("Cert:%+v\n", a.Value)
			break
		}
	}

	return nil
}

func (s ShowCert) validate() error {
	if s.HsmLib == "" {
		return fmt.Errorf("--hsm not found")
	}
	if _, err := os.Stat(s.HsmLib); os.IsNotExist(err) {
		return fmt.Errorf("'%s' does not exist", s.HsmLib)
	}

	return nil
}

func (s ShowCert) stringToInts(str string) ([]int, error) {
	var id []int
	for _, r := range str {
		n, err := strconv.ParseInt(string(r), 16, 64)
		if err != nil {
			return []int{}, fmt.Errorf("cannot convert %s to hex id", str)
		}
		id = append(id, int(n))
	}
	return id, nil
}
