# EBM-GO

Simple Ebook Management Library In Go

## Installation

- ### Quick Build Method: 

  #### Steps:
    1. Build
      ```sh
        go build -tags sqlite_fts5 -ldflags="-s -w" -gcflags '-N -l' .
      ```

    2. Run
      ```sh
        ./ebmgo
      ```

- ### Optimized Build Method (Recommended for smaller, static binary): 
    #### Requirement
    - **Musl Toolchain** (for static build):
    
    #### Steps:
    1. Install dependencies 
      ```sh
        sudo apt update
        sudo apt install build-essential musl musl-tools musl-dev
      ```

    2. Build
      ```bash
        make
      ```
    3. Run
      ```sh
        ./ebmgo
      ```

## Usage 

### Import books

    Usage: import [options] [directory]

    Options:
      -h	Show help
    
    Examples:

    $ ebm import ./sample.pdf
    
    $ ebm import -h

### List books

    Usage: list [options]

    Options:
      -s    Filter the results by the search query
      -h	Show help

    Example:

    $ ebm list -s "modern"

### Remove books

    Usage: remove [options]

    Options:
      -h	Show help
      -ids string
            Book ID to remove. Separe by ','

    Example:

    # ebm remove -ids "1,2,3"


