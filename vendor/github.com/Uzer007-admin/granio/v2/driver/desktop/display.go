package desktop

// DisplayInfo описывает фактический оконный backend после запуска драйвера.
// SessionType — подсказка окружения, а не доказательство использования XWayland.
type DisplayInfo struct {
	Backend     string
	Requested   string
	SessionType string
}

// DisplayDriver предоставляет диагностику выбора backend, сохраняя интерфейс Driver.
type DisplayDriver interface{ DisplayInfo() DisplayInfo }
