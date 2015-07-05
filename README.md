# gPhoto2-Go

A more Go idiomatic interface to the gPhoto2 library.

## Warning

I'm not a great Golang programmer, so this could be a complete disaster. But, if you want to use the gPhoto2 library and you'd prefer to avoid sticking cgo references all over your main Go program, this library might help.

## Installation

`go get github.com/micahwedemeyer/gphoto2go`


## Requirements

You will also need libgphoto2 installed. If you are on Mac OS X, I recommend installing it with homebrew.

`brew install libgphoto2`

In order to compile your Go program, you will need to set CFLAGS and LDFLAGS to find the libgphoto2 libraries. I will update with more instructions on that once I understand, but the following worked for me with the homebrew-installed libgphoto2

    // #cgo CFLAGS: -I/Users/micah/Developer/include/gphoto2
    // #cgo LDFLAGS: -L/Users/micah/Developer/lib -lgphoto2
    // #include <gphoto2.h>

## Usage

The main goal with this library is to present a Go-friendly interface to the C methods of gPhoto2

### Camera Initializing

    camera := new(gphoto2.Camera)
    err := camera.Init()

    if err < 0 {
        fmt.Printf(gphoto2.CameraResultToString(err))
    }

This will create a new Camera struct and intitialize it, which prompts gphoto2 to auto-detect any connected USB cameras.

### Taking a Photo

    camera.trigger_capture()

This will trigger the camera.

### Downloading the Photos from the Camera

Coming soon...
