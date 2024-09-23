# Tronco

A lossy video encoder that "compresses" video using delaunay triangulation.

## Table of Contents
- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)

## Features

- At least 15 times worse at compression than h.265
- Double the triangles!
- 

## Requirements

- ffmpeg
- go
- linux? I have not tried this on other platforms

## Installation

1. Clone the repository:

`git clone https://github.com/Shhwip/tronco.git && cd tronco`

2. Run the build script:

`./build.sh`


## Usage

`tronco <video_file> <output_folder>`

You should keep in mind that this needs at least 20 times the storage space of the original file.

