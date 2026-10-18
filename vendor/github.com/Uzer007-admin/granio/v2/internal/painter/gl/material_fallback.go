//go:build android || ios || mobile || wasm || test_web_driver

package gl

func newMaterialPipeline(*painter) materialPipeline { return nil }
