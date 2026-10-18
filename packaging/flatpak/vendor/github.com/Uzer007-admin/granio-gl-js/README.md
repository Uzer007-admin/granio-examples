# granio-gl-js

Компонент независимого проекта **Granio — Гранио**.

**Исходный репозиторий:** [gl-js](https://github.com/fyne-io/gl-js).
**Цель форка:** OpenGL, GLES и WebGL для графического драйвера Granio.

История изменений, лицензии и сведения об исходных авторах сохранены.
Go-модули переведены на собственные пути `github.com/Uzer007-admin/…`;
минимальная версия Go — **1.27.1**. Ядро и общие API подключаются из
[`granio/v2`](https://github.com/Uzer007-admin/granio).

Состояние переноса и результаты проверок приведены в
[реестре экосистемы](https://github.com/Uzer007-admin/granio/blob/dev/granio/docs/ECOSYSTEM.md).
Наличие форка само по себе не означает проверку всех платформ и сценариев.

---

## Документация исходного проекта

Команды установки и ссылки в сохранённом тексте ниже относятся к исходнику.
Для Granio используются новые пути модулей, перечисленные выше.

[![Tests](https://github.com/fyne-io/gl-js/actions/workflows/tests.yml/badge.svg)](https://github.com/fyne-io/gl-js/actions/workflows/tests.yml)
[![Static Analysis](https://github.com/fyne-io/gl-js/actions/workflows/analysis.yml/badge.svg)](https://github.com/fyne-io/gl-js/actions/workflows/analysis.yml)
[![GoDoc](https://godoc.org/github.com/fyne-io/gl-js?status.svg)](https://godoc.org/github.com/fyne-io/gl-js)

# GL

Package gl is a Go cross-platform binding for OpenGL, with an OpenGL ES 2-like API.

It supports:

- **macOS**, **Linux** and **Windows** via OpenGL 2.1 backend,

- **iOS** and **Android** via OpenGL ES 2.0 backend,

- **Modern Browsers** (desktop and mobile) via WebGL 1.0 backend.

This is a fork of golang.org/x/mobile/gl package with [CL 8793](https://go-review.googlesource.com/8793)
merged in and Windows support added. This package is fully functional, but may eventually become superceded by
the new x/mobile/gl plan. It will exist and be fully supported until it can be safely replaced by a better package.

## Installation

```bash
go get github.com/fyne-io/gl-js
```

## Usage

This OpenGL binding has a ContextWatcher, which implements [glfw.ContextWatcher](https://godoc.org/github.com/goxjs/glfw#ContextWatcher)
interface. Recommended usage is with github.com/fyne-io/glfw-js package, which accepts a ContextWatcher in its Init, and takes on the responsibility
of notifying it when context is made current or detached.

```go
if err := glfw.Init(gl.ContextWatcher); err != nil {
	// Handle error.
}
defer glfw.Terminate()
```

If you're not using a ContextWatcher-aware glfw library, you must call methods of gl.ContextWatcher yourself whenever
you make a context current or detached.

```Go
window.MakeContextCurrent()
gl.ContextWatcher.OnMakeCurrent(nil)

glfw.DetachCurrentContext()
gl.ContextWatcher.OnDetach()
```
