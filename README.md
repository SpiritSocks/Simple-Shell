# Simple-Shell

## Этап 1

- Приложение реализовано в форме консольного интерфейса (CLI).
- Приглашение к вводу формируется на основе реальных данных ОС, в которой исполняется эмулятор. Пример: `username@hostname:~$`.
- Реализован парсер, который корректно обрабатывает аргументы в кавычках.
- Есть сообщения об ошибке выполнения команд (неизвестная команда, неверные аргументы).
- Реализованы команды-заглушки: `ls`, `cd`.
- Реализована команда `exit`.

**Пример:**

```
***@**.local:~$ ls -l
Command: ls, Arguments: [-l]
```

---

## Этап 2s

- Добавлена поддержка параметров командной строки:
  - `--vfs` — путь к виртуальной файловой системе.
  - `--script` — путь к стартовому скрипту.
- Реализован запуск стартового скрипта (команды выполняются последовательно, при ошибке выполнение прерывается).
- Выводятся отладочные параметры конфигурации при старте:
  ```
  conf: vfs_path="./vfs/vfs.json", start_script="./script/start.sh"
  ```

**Пример запуска:**

```bash
go run ./cmd/emulator/main.go --vfs ./vfs/vfs.json --script ./script/start.sh
```

---

## Этап 3

- Подключена виртуальная файловая система (**VFS**) из JSON:
  - Вся работа ведётся в памяти.
  - Поддерживаются директории и файлы (содержимое файлов хранится в base64).
- Реализованы команды работы с VFS:
  - `pwd` — показать текущий виртуальный путь.
  - `cd` — переход по виртуальным каталогам.
  - `ls` — вывод содержимого текущего каталога.
- Подготовлены примеры JSON VFS и стартового скрипта для тестирования.

**Пример `vfs.json`:**

```json
{
  "name": "/",
  "is_dir": true,
  "children": {
    "home": {
      "name": "home",
      "is_dir": true,
      "children": {
        "user": {
          "name": "user",
          "is_dir": true,
          "children": {
            "hello.txt": {
              "name": "hello.txt",
              "is_dir": false,
              "content": "SGVsbG8gd29ybGQh"
            }
          }
        }
      }
    },
    "etc": {
      "name": "etc",
      "is_dir": true,
      "children": {
        "conf.txt": {
          "name": "conf.txt",
          "is_dir": false,
          "content": "Q29uZmlndXJhdGlvbiBmaWxl"
        }
      }
    }
  }
}
```

**Пример `script/start.sh`:**

```bash
ls
cd home
ls
cd user
pwd
ls
cd ..
cd ..
cd etc
ls
```

**Пример запуска:**

```bash
go run ./cmd/emulator/main.go --vfs ./vfs/vfs.json --script ./script/start.sh
```

**Вывод:**

```
/           # корень
home/
etc/
/home
user
/home/user
hello.txt
/etc
conf.txt
```

---
