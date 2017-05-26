# gPhoto2-Go

A more Go idiomatic interface to the gPhoto2 library.

## Installation

```
go get github.com/charles-d-burton/gphoto2go
```

## Requirements

You will also need libgphoto2 installed. If you are on Mac OS X, I recommend installing it with homebrew.
```
brew install libgphoto2
```

This should now compile by itself thanks to this `// #cgo pkg-config: libgphoto2`

## Usage

The main goal with this library is to present a Go-friendly interface to the C methods of gPhoto2

### Camera Initializing

```go
camera := new(gphoto2go.Camera)
err := camera.Init()
```

This will create a new Camera struct and intitialize it, which prompts gphoto2 to auto-detect any connected USB cameras.

### Taking a Photo

```go
camera.TriggerCapture()
```
This will trigger the camera.

### Downloading the Photos from the Camera

```go
folders := camera.RListFolders("/")
for _, folder := range folders {
    files, _ := camera.ListFiles(folder)
    for _, fileName := range files {
        cameraFileReader := camera.FileReader(folder, fileName)
        fileWriter := os.Create("/tmp/" + fileName)
        io.Copy(fileWriter, cameraFileReader)

        // Important, since there is memory used in the transfer that needs to be freed up
        cameraFileReader.Close()
    }
}
```
### Interpreting errors

Most of the functions will return an error code. If it is less than zero, that means an error has occurred. The library can translate the error integer
into a human readable string.

```go
err := camera.TriggerCapture()
if err < 0 {
    fmt.Printf(gphoto2go.CameraResultToString(err))
}
```
