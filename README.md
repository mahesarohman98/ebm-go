# EBM-GO

**EBM-GO** (Ebook Management Library in Go) is a simple, fast, and lightweight ebook management tool built in Go. It supports importing, listing, and removing ebooks with ease.

----------

## Features

-   Import and manage ebooks quickly.
-   Search ebooks using filters.
-   Static binary build support.
-   Minimal dependencies.

----------

## Installation

You can install EBM-GO using either the quick build method or the optimized build method.

### Quick Build (Development)

#### Build:

```bash
go build -tags sqlite_fts5 -ldflags="-s -w" -gcflags '-N -l' .

```

#### Run:

```bash
./ebmgo

```

### Optimized Build (Recommended)

Build a smaller, static binary using **Musl Toolchain**:

#### Requirements:

```bash
sudo apt update
sudo apt install build-essential musl musl-tools musl-dev

```

#### Build:

```bash
make

```

#### Run:

```bash
./ebmgo

```

----------

## Usage

### Import Books

```bash
ebm import [options] [directory]

```

**Options:**

-   `-h` — Show help

**Examples:**

```bash
ebm import ./sample.pdf
ebm import -h

```

----------

### List Books

```bash
ebm list [options]

```

**Options:**

-   `-f` — The fields to display when listing books in the db. Available fields: title, authors, formats. Default: title,authors. (default "title,authors")
-   `-s` — Filter results by search query
-   `-h` — Show help

**Example:**

```bash
ebm list -s "modern"

```

----------

### Remove Books

```bash
ebm remove [options]

```

**Options:**

-   `-ids string` — Comma-separated book IDs to remove
-   `-h` — Show help

**Example:**

```bash
ebm remove -ids "1,2,3"

```

----------

## License

This project is licensed under the terms of the GNU General Public License v3.0. See the LICENSE file for details.

----------

## Author

**Mahesa Rohman**
