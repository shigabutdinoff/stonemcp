package engine

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

// DefaultAddr задаёт адрес игры MindConnect на этом компьютере.
const DefaultAddr = "127.0.0.1:64649"

// ModeSync задаёт имя единственного поддержанного режима MindConnect.
const ModeSync = "sync"

// Config задаёт соединение и лимиты движка.
// Нули выбирают значения по умолчанию.
type Config struct {
	// Addr задаёт адрес игры, по умолчанию локальный порт MindConnect.
	Addr string
	// Logger принимает сообщения соединения, nil отключает журнал.
	Logger *slog.Logger
	// ReconnectDelays задаёт паузы перед повторным подключением,
	// по умолчанию 1 с, 3 с и 5 с. Последняя пауза повторяется.
	ReconnectDelays []time.Duration
	// ReadTimeout ограничивает обмен внутри шага, по умолчанию 2 с.
	ReadTimeout time.Duration
	// ProbeInterval задаёт период проверки связи в заморозке, по умолчанию 1 с.
	ProbeInterval time.Duration
	// FrameBudget задаёт бюджет запросов кадра, по умолчанию 512.
	// Шаги pre и post расходуют бюджет независимо друг от друга.
	// Один Evaluate может превысить бюджет, сохраняя общий кадр значений.
	FrameBudget int
	// MaxPacket ограничивает размер пакета, по умолчанию SafePacket.
	MaxPacket int
	// QueueLimit ограничивает число незавершённых команд, по умолчанию 32.
	QueueLimit int
	// HoldLimit ограничивает число одновременных удержаний, по умолчанию 32.
	// Значение должно быть меньше FrameBudget.
	HoldLimit int
	// CommandTTL ограничивает срок отправки команды, по умолчанию 2 с.
	// Остановка кадров игры срок не расходует.
	CommandTTL time.Duration
	// CommandTimeout задаёт защитное ожидание команды, по умолчанию 30 с.
	CommandTimeout time.Duration
}

func (c Config) defaults() Config {
	orDefault(&c.Addr, DefaultAddr)
	c.ReconnectDelays = slices.Clone(c.ReconnectDelays)
	if c.ReconnectDelays == nil {
		c.ReconnectDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}
	}
	orDefault(&c.ReadTimeout, 2*time.Second)
	orDefault(&c.ProbeInterval, time.Second)
	orDefault(&c.FrameBudget, 512)
	orDefault(&c.MaxPacket, mc.SafePacket)
	orDefault(&c.QueueLimit, 32)
	orDefault(&c.HoldLimit, 32)
	orDefault(&c.CommandTTL, 2*time.Second)
	orDefault(&c.CommandTimeout, 30*time.Second)
	return c
}

func orDefault[T comparable](v *T, def T) {
	var zero T
	if *v == zero {
		*v = def
	}
}

func (c Config) validate() error {
	if c.QueueLimit <= 0 || c.HoldLimit <= 0 || c.CommandTTL <= 0 || c.CommandTimeout <= 0 {
		return fmt.Errorf("недопустимые лимиты команд")
	}
	if c.FrameBudget <= c.HoldLimit {
		return fmt.Errorf("бюджет кадра должен превышать предел удержаний")
	}
	if !validDelays(c.ReconnectDelays) || c.ReadTimeout <= 0 || c.ProbeInterval <= 0 || c.FrameBudget <= len(baseVars) || c.MaxPacket < 0 {
		return fmt.Errorf("недопустимые лимиты движка")
	}
	return fitAll(baseRequests, c.MaxPacket)
}
