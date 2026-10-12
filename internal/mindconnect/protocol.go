package mindconnect

const (
	// FieldSep разделяет поля пакета и ответа.
	FieldSep = "\x1f"
	// EndStep завершает шаг кадра; клиент шлёт его отдельной записью.
	EndStep = "\x03"
	// SignalPrefix открывает сигнал кадра, который шлёт игра.
	SignalPrefix = "\x06"
	// MaxPacket задаёт размер приёмного буфера игры.
	MaxPacket = 65536
	// MaxResponse ограничивает ответ в байтах, сверх него Feed даёт ошибку.
	MaxResponse = 1 << 20
	// SafePacket задаёт наибольший размер пакета, который собирает кодек.
	SafePacket = 8192
	// Sentinel замыкает каждый пакет; его значение True или False.
	Sentinel = "ai.enabled"
	// MaxID задаёт наибольший id пакета; id идут от 1 до MaxID.
	MaxID = 9999
)
