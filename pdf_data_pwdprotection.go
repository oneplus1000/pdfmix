package pdfmix

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var protectionPadding = []byte{
	0x28, 0xBF, 0x4E, 0x5E, 0x4E, 0x75, 0x8A, 0x41, 0x64, 0x00, 0x4E, 0x56, 0xFF, 0xFA, 0x01, 0x08,
	0x2E, 0x2E, 0x00, 0xB6, 0xD0, 0x68, 0x3E, 0x80, 0x2F, 0x0C, 0xA9, 0xFE, 0x64, 0x53, 0x69, 0x7A,
}

type pwdProtectionInfo struct {
	encryptionKey []byte
}

func (p *PdfData) pwdProtection(passInfo *PasswordInfo) error {

	encryNodes, encryptionKey, err := p.createEncryptNodes(passInfo)
	if err != nil {
		return errors.Wrap(err, "createEncryptNodes(...) fail")
	}
	encryNodeID := p.newRealID()
	p.objects[encryNodeID] = &encryNodes
	p.pwdProtectionInfo = &pwdProtectionInfo{
		encryptionKey: encryptionKey,
	}
	rootID := createRealObjectID(0)
	if root, ok := p.objects[rootID]; ok {
		en := initNodeKeyUseNameAndNodeContentUseRefTo("Encrypt", encryNodeID)
		root.append(en)
	}
	return nil
}

func extractHexStringForEncrypt(str string) (string, bool) {
	r := []rune(str)
	size := len(r)
	if size <= 2 {
		return "", false
	}
	var buff bytes.Buffer
	if r[0] == '<' && r[size-1] == '>' {
		for i := 1; i < size-1; i++ {
			buff.WriteString(fmt.Sprintf("%s", string(r[i])))
		}
	}
	return buff.String(), true
}

func (p *PdfData) readTrailerID() ([]byte, error) {
	idOfTrailer := createRealObjectID(uint32(0)) //id ของ trailer คือ 0 แน่นอน
	trailerIDNode, err := newQuery(p).findPdfNodeByKeyName(idOfTrailer, "ID")
	if err == ErrKeyNameNotFound || err == ErrObjectIDNotFound {
		return nil, nil //ok ไม่เจอก็ข้ามไป
	} else if err != nil {
		return nil, errors.Wrap(err, "")
	} else {
		if trailerIDNode.content.use == constNodeKeyUseIndex {
			idNodes := p.objects[trailerIDNode.content.refTo]
			for _, idNode := range *idNodes {
				if idNode.key.use == constNodeKeyUseIndex &&
					idNode.content.use == constNodeContentUseString {
					id, ok := extractHexStringForEncrypt(idNode.content.str)
					if ok {
						return hex.DecodeString(id)
					}
				}
				break //ใช้แค่อันแรก
			}
		}
	}
	return nil, nil
}

func (p *PdfData) createEncryptNodes(passInfo *PasswordInfo) (pdfNodes, []byte, error) {

	trailerID, err := p.readTrailerID()
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}

	userPass := passInfo.UserPass
	ownerPass := passInfo.OwnerPass
	permissions := passInfo.Permissions
	protection := 192 | permissions
	if ownerPass == nil || len(ownerPass) == 0 {
		ownerPass = randomPwd(24)
	}
	userPass = append(userPass, protectionPadding...)
	userPassWithPadding := userPass[0:32]
	ownerPass = append(ownerPass, protectionPadding...)
	ownerPassWithPadding := ownerPass[0:32]
	oValue, err := createPwdProtectionOValue(userPassWithPadding, ownerPassWithPadding)
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	uValue, encryptionKey, err := createPwdProtectionUValue(userPassWithPadding, oValue, trailerID, protection)
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	pValue := -((protection ^ 255) + 1)

	encryNodes := pdfNodes{}
	filter := initNodeKeyUseNameAndNodeContentUseString("Filter", "/Standard")
	vVal := initNodeKeyUseNameAndNodeContentUseString("V", "1")
	rVal := initNodeKeyUseNameAndNodeContentUseString("R", "2")
	oVal := initNodeKeyUseNameAndNodeContentUseString("O", fmt.Sprintf("(%s)", escapeValue(oValue)))
	uVal := initNodeKeyUseNameAndNodeContentUseString("U", fmt.Sprintf("(%s)", escapeValue(uValue)))
	pVal := initNodeKeyUseNameAndNodeContentUseString("P", fmt.Sprintf("%d", pValue))
	encryNodes.append(filter)
	encryNodes.append(vVal)
	encryNodes.append(rVal)
	encryNodes.append(oVal)
	encryNodes.append(uVal)
	encryNodes.append(pVal)

	return encryNodes, encryptionKey, nil
}

/*
//createContentPwdProtectionAll เข้ารหัส content
func (p *PdfData) createContentPwdProtectionAll(contentObjectIDs map[int]objectID) error {
	for pageIndex, objID := range contentObjectIDs {
		stm, err := p.getStreamOfContentOfPage(contentObjectIDs, pageIndex) //ดึงข้อมูลเดิมออกมา
		if err != nil {
			return errors.Wrap(err, "getStreamOfContentOfPage(...) fail")
		}
		stmCip, err := p.createContentPwdProtection(stm, objID.id, p.pwdProtectionInfo.encryptionKey)
		if err != nil {
			return errors.Wrap(err, "createContentPwdProtection(...) fail")
		}
		fmt.Printf("\n%x\n\n", stmCip.Bytes())
		err = p.replaceStramObj(contentObjectIDs[pageIndex], stmCip) // ยัดลง pdfdata
		if err != nil {
			return errors.Wrapf(err, "p.reWriteStramObj(contentObjectIDs[%d], stm) fail", pageIndex)
		}
	}
	return nil
}
*/
/*
func (p *PdfData) createContentPwdProtection(stm *bytes.Buffer, id uint32, encryptionKey []byte) (*bytes.Buffer, error) {
	objKey := createPwdProtectionObjectKey(id, encryptionKey)
	cip, err := rc4Cip(objKey, stm.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "rc4Cip(...) fail")
	}
	stmCip := bytes.NewBuffer(cip)
	return stmCip, nil
}
*/

func createPwdProtectionObjectKey(n uint32, encryptionKey []byte) []byte {
	tmp := make([]byte, 8, 8)
	binary.LittleEndian.PutUint32(tmp, n)
	tmp2 := append(encryptionKey, tmp[0], tmp[1], tmp[2], 0, 0)
	tmp3 := md5.Sum(tmp2)
	return tmp3[0:10]
}

func createPwdProtectionOValue(userPassWithPadding []byte, ownerPassWithPadding []byte) ([]byte, error) {
	tmp := md5.Sum(ownerPassWithPadding)
	ownerRC4key := tmp[0:5]
	cip, err := rc4.NewCipher(ownerRC4key)
	if err != nil {
		return nil, err
	}
	dest := make([]byte, len(userPassWithPadding))
	cip.XORKeyStream(dest, userPassWithPadding)
	return dest, nil
}

func createPwdProtectionUValue(userPassWithPadding []byte, oValue []byte, trailerID []byte, protection int) ([]byte, []byte, error) {

	var tmp bytes.Buffer
	tmp.Write(userPassWithPadding)
	tmp.Write(oValue)
	tmp.WriteByte(byte(protection))
	tmp.WriteByte(byte(0xff))
	tmp.WriteByte(byte(0xff))
	tmp.WriteByte(byte(0xff))
	if trailerID != nil {
		tmp.Write(trailerID)
	}

	tmp2 := md5.Sum(tmp.Bytes())
	encryptionKey := tmp2[0:5]
	cip, err := rc4.NewCipher(encryptionKey)
	if err != nil {
		return nil, nil, err
	}
	dest := make([]byte, len(protectionPadding))
	cip.XORKeyStream(dest, protectionPadding)
	return dest, encryptionKey, nil
}

func randomPwd(strlen int) []byte {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdef0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return result
}

func escapeValue(b []byte) string {
	s := string(b)
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "(", "\\(", -1)
	s = strings.Replace(s, ")", "\\)", -1)
	s = strings.Replace(s, "\r", "\\r", -1)
	return s
}

func rc4Cip(key []byte, src []byte) ([]byte, error) {
	cip, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	dest := make([]byte, len(src))
	cip.XORKeyStream(dest, src)
	return dest, nil
}
