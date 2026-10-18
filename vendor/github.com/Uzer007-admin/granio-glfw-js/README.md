# granio-glfw-js

Компонент независимого проекта **Granio — Гранио**.

**Исходный репозиторий:** [glfw-js](https://github.com/fyne-io/glfw-js).
**Цель форка:** Браузерные окна и ввод для Granio.

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

[![Tests](https://github.com/fyne-io/glfw-js/actions/workflows/tests.yml/badge.svg)](https://github.com/fyne-io/glfw-js/actions/workflows/tests.yml)
[![Static Analysis](https://github.com/fyne-io/glfw-js/actions/workflows/analysis.yml/badge.svg)](https://github.com/fyne-io/glfw-js/actions/workflows/analysis.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/fyne-io/glfw-js.svg)](https://pkg.go.dev/github.com/fyne-io/glfw-js)

# glfw-js

Package glfw experimentally provides a glfw-like API
with desktop (via glfw) and browser (via HTML5 canvas) backends.

It is used for creating a GL context and receiving events.

**Note:** This package is currently in development. The API is incomplete and may change.

## Installation

```sh
go get github.com/fyne-io/glfw-js
```

## Directories

| Path                                                                | Synopsis                                                           |
|---------------------------------------------------------------------|--------------------------------------------------------------------|
| [test/events](https://pkg.go.dev/github.com/goxjs/glfw/test/events) | events hooks every available callback and outputs their arguments. |

## License

-	[MIT License](LICENSE)
