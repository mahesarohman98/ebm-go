# EBM-GO

Simple Ebook Management Library In Go

## Installation

    go build -tags sqlite_fts5 .

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
