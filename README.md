# StoneMCP

StoneMCP даёт ассистенту играть в Stone Story RPG в реальном времени. Это
сервер [MCP](https://modelcontextprotocol.io): он подключается к игре через
MindConnect и отдаёт её ассистенту инструментами. Ассистент читает состояние
забега, меняет оружие, пьёт зелья, ждёт событий и останавливает игру на кадре.

Работает на macOS, Linux и Windows. Игре нужен MindConnect 0.3 в режиме
`sync`.

## Документация

Как поставить и подключить, написано [ниже](#установка). Настройки собраны в
разделе [Настройка](#настройка), инструменты в разделе
[Инструменты](#инструменты). Краткую справку по флагам печатает
`stonemcp -h`.

## Разработка

Stone Story RPG для тестов не нужна. Тесты запускают тестовую игру из
`internal/fakegame/bin`, там лежат готовые бинарники для darwin/amd64,
darwin/arm64, linux/amd64, linux/arm64 и windows/amd64.

```sh
git clone https://github.com/shigabutdinoff/stonemcp.git
cd stonemcp
go vet ./...
go test -race ./...
make lint
```

Исходный код тестовой игры закрыт и подключён сабмодулем
`internal/fakegame/src`. Клонируйте без `--recurse-submodules`: без доступа
git остановится на этом сабмодуле с ошибкой, хотя сам репозиторий скачается.
Пересобрать бинарники может только тот, у кого есть доступ, командой
`make fakegame`.

## Установка

### Из исходников

Нужен Go 1.27 или новее.

```sh
go install github.com/shigabutdinoff/stonemcp/cmd/stonemcp@latest
```

Бинарник появится в `$(go env GOPATH)/bin`. Из клона репозитория то же делает
`make build`, он кладёт `stonemcp` в текущий каталог. `make release` собирает
бинарники для шести платформ в `dist/`: macOS, Linux и Windows под amd64 и
arm64. Файл выпущенной версии называется
`stonemcp-<версия>-<номер сборки>-<os>-<arch>`, рабочей сборки
`stonemcp-dev-<os>-<arch>`.

### Подключение к ассистенту

StoneMCP общается с клиентом по stdio, поэтому клиенту достаточно знать
команду запуска. В Claude Code:

```sh
claude mcp add stonemcp -- stonemcp
```

В клиентах с файлом настроек:

```json
{
  "mcpServers": {
    "stonemcp": {
      "command": "stonemcp"
    }
  }
}
```

### Подключение игры

1. Запустите Stone Story RPG.
2. Сделайте первой строкой программы Mind Stone `//MindConnect: sync`.
3. Включите Mind Stone и зайдите в локацию.

StoneMCP подключается к игре по адресу `127.0.0.1:64649`. Порядок запуска не
важен: сервер работает и без игры, а после обрыва подключается заново сам.
Есть ли связь, покажет инструмент `stone_status`.

## Настройка

| Флаг | Переменная окружения | Ключ в файле | По умолчанию | Что задаёт |
|---|---|---|---|---|
| `--addr` | `STONEMCP_ADDR` | `addr` | `127.0.0.1:64649` | адрес игры |
| `--mode` | `STONEMCP_MODE` | `mode` | `sync` | режим MindConnect, поддержан только `sync` |
| `--log-file` | `STONEMCP_LOG_FILE` | `log_file` | stderr | файл журнала |
| `--config`, `-c` | `STONEMCP_CONFIG` | | | файл конфигурации JSON |

Переменная окружения сильнее флага, флаг сильнее файла. Stdout занят
протоколом MCP, поэтому журнал туда не пишется.
`stonemcp --version` печатает версию, номер, дату и коммит сборки:

```
Build version: 27.0
Build number: 27A41
Build date: 2026-10-13
Build commit: 3f2a9c1
```

`stonemcp completion <оболочка>` печатает скрипт автодополнения для bash,
zsh, fish или powershell, `stonemcp help` справку по командам.

Файл конфигурации:

```json
{
  "addr": "127.0.0.1:64649",
  "mode": "sync",
  "log_file": "/tmp/stonemcp.log"
}
```

## Инструменты

| Инструмент | Что делает |
|---|---|
| `stone_status` | показывает, подключена ли игра: версия, режим, кадр, fps, локация, последняя ошибка |
| `stone_snapshot` | читает состояние забега из последнего кадра: герой, враг, локация |
| `stone_screenshot` | читает ASCII-экран игры текстом |
| `stone_evaluate` | читает переменные и вызывает функции Stonescript; игру не меняет |
| `stone_wait_for` | ждёт условия или таймаута вместо опроса состояния |
| `stone_equip` | экипирует предмет в обе руки |
| `stone_equip_left` | экипирует предмет в левую руку |
| `stone_equip_right` | экипирует предмет в правую руку |
| `stone_loadout` | переключает сохранённый набор экипировки по номеру |
| `stone_activate` | активирует способность или предмет |
| `stone_brew` | смешивает зелье из ингредиентов |
| `stone_print` | показывает текст на экране игры |
| `stone_freeze` | останавливает игру на ближайшем кадре |
| `stone_release` | отпускает замороженную игру |

Действия возвращают ответ игры и свежее состояние, поэтому после них не нужно
отдельно читать снимок.

## Профилирование памяти

Бенчмарк `BenchmarkExchangeValues` собирает значения ответа на кадр экрана
90×40 так же, как движок: запросы режутся на пакеты по `SafePacket`, ответ
каждого пакета читается частями по 1460 байт. Профиль `profiles/base.pprof`
снят до оптимизации, `profiles/result.pprof` после неё. Оба сняты при
20000 итерациях, поэтому их разность сравнима.

Снять профиль:

```sh
go test -run '^$' -bench BenchmarkExchangeValues -benchtime 20000x -benchmem \
	-memprofile profiles/result.pprof ./internal/mindconnect
```

Сравнить с базой:

```sh
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

```text
File: mindconnect.test
Type: alloc_space
Time: 2026-09-30 17:01:50 MSK
Showing nodes accounting for -4.24GB, 62.55% of 6.78GB total
Dropped 30 nodes (cum <= 0.03GB)
      flat  flat%   sum%        cum   cum%
   -3.08GB 45.39% 45.39%    -4.25GB 62.68%  github.com/shigabutdinoff/stonemcp/internal/mindconnect.BenchmarkExchangeValues
   -1.16GB 17.17% 62.55%    -1.16GB 17.17%  slices.Clone[go.shape.[]github.com/shigabutdinoff/stonemcp/internal/mindconnect.Value,go.shape.string] (inline)
         0     0% 62.55%    -1.16GB 17.17%  github.com/shigabutdinoff/stonemcp/internal/mindconnect.(*ResponseReader).Response
         0     0% 62.55%    -4.25GB 62.68%  testing.(*B).run1.func1
         0     0% 62.55%    -4.25GB 62.68%  testing.(*B).runN
```

| Путь | ns/op | B/op | allocs/op |
|---|---|---|---|
| до, `Response` и `append` | 37828 | 361950 | 69 |
| после, `AppendValues` | 19104 | 137693 | 52 |

`Response` копирует значения пакета через `slices.Clone`, а движок
наращивал общий срез через `append`. Теперь `ResponseReader.AppendValues`
дописывает значения прямо в срез вызывающего, а движок один раз выделяет
срез на все пакеты кадра.

## Лицензия

[MIT](LICENSE).
