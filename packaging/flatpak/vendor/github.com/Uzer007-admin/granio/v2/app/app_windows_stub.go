//go:build windows && mobile

package app

import (
	"log"

	"github.com/Uzer007-admin/granio/v2"
)

func NewWithID(_ string) granio.App {
	log.Fatal("Cannot launch the mobile simulator mode on Windows")
	return nil
}
