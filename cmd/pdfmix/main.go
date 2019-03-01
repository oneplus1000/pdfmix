package main

import (
	"flag"
	"log"

	"github.com/oneplus1000/pdfmix/cmd"
)

func main() {
	//cmd
	cmdShowCert := flag.Bool("cert", false, "show cert in hsm")
	//prop
	propHsmLib := flag.String("hsm", "", "path to hsm .dll or .so")
	propLabel := flag.String("label", "", "label")
	propID := flag.String("id", "", "id")
	//parse
	flag.Parse()

	if *cmdShowCert {
		c := cmd.ShowCert{
			HsmLib: *propHsmLib,
			Label:  *propLabel,
			ID:     *propID,
		}
		err := c.Exec()
		if err != nil {
			echoErr(err)
		}
	}

}

func echoErr(err error) {
	log.Panic(err)
}
