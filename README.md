# Tronco

A lossy video encoder that "compresses" video using delaunay triangulation.


## Features

- At least 15 times worse at compression than h.265
- Double the triangles!
- 3

## Requirements

- ffmpeg
- go

## Installation

1. go install:

`go install github.com/shhwip/tronco@latest`


## Usage

`tronco <video_file> <output_folder>`

You should keep in mind that this needs at least 20 times the storage space of the original file.

## Examples


