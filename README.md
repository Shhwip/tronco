# Tronco

A lossy video encoder that "compresses" video using delaunay triangulation.


## Features

- At least 15 times worse at compression than h.265
- Double the triangles!
- 3

## Requirements

- [ffmpeg](https://www.ffmpeg.org/download.html)
- [go](https://go.dev/doc/install)

## Installation

1. go install:

`go install github.com/shhwip/tronco@latest`


## Usage

`tronco <video_file> <output_folder>`

You should keep in mind that this needs at least 20 times the storage space of the original file.

## Examples


after:

https://github.com/user-attachments/assets/540eccaf-101f-4bfb-9edd-8a3fc0c68807



