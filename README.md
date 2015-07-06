# gPhoto2-Go

A more Go idiomatic interface to the gPhoto2 library.

## Warning

I'm not a great Golang programmer, so this could be a complete disaster. But, if you want to use the gPhoto2 library and you'd prefer to avoid sticking cgo references all over your main Go program, this library might help.

I'm also not a great C programmer, having grown up with garbage collectors and other niceties. Therefore, this library will be riddled with memory leaks. I would appreciate any help in making it more memory efficient.

## Installation

    go get github.com/micahwedemeyer/gphoto2go

## Requirements

You will also need libgphoto2 installed. If you are on Mac OS X, I recommend installing it with homebrew.

    brew install libgphoto2

In order to compile your Go program, you will need to set CFLAGS and LDFLAGS to find the libgphoto2 libraries. I will update with more instructions on that once I understand, but the following worked for me with the homebrew-installed libgphoto2

    // #cgo CFLAGS: -I/Users/micah/Developer/include/gphoto2
    // #cgo LDFLAGS: -L/Users/micah/Developer/lib -lgphoto2
    // #include <gphoto2.h>

## Usage

The main goal with this library is to present a Go-friendly interface to the C methods of gPhoto2

### Camera Initializing

    camera := new(gphoto2.Camera)
    err := camera.Init()


This will create a new Camera struct and intitialize it, which prompts gphoto2 to auto-detect any connected USB cameras.

### Taking a Photo

    camera.TriggerCapture()

This will trigger the camera.

### Downloading the Photos from the Camera

    cameraFileReader := camera.BufferedFileReader("/store_00020001/DCIM/100CANON", "IMG_8085.JPG")
    fileWriter := os.Create("/tmp/myfile.jpg")
    io.Copy(fileWriter, cameraFileReader)

### Interpreting errors

Most of the functions will return an error code. If it is less than zero, that means an error has occurred. The library can translate the error integer
into a human readable string.

    err := camera.TriggerCapture()
    if err < 0 {
        fmt.Printf(gphoto2.CameraResultToString(err))
    }
