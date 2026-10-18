package lang

import (
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/app"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/mobileinit"

	"github.com/jeandeaual/go-locale"
)

func initRuntime() {
	locale.SetRunOnJVM(app.RunOnJVM)
}

// runtimeReady reports whether we can call into the JVM yet.
var runtimeReady = func() bool {
	return mobileinit.HasContext()
}
