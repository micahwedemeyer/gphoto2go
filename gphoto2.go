package gphoto2

// #cgo CFLAGS: -I/Users/micah/Developer/include/gphoto2
// #cgo LDFLAGS: -L/Users/micah/Developer/lib -lgphoto2
// #include <gphoto2.h>
// #include <gphoto2/gphoto2-version.h>
import "C"
import "unsafe"
import "fmt"

type Camera struct {
	camera  *C.Camera
	context *C.GPContext
}

func (c *Camera) Init() int {
	c.context = C.gp_context_new()

	C.gp_camera_new(&c.camera)
	err := C.gp_camera_init(c.camera, c.context)
	return int(err)
}

func (c *Camera) GetAbilities() (C.CameraAbilities, int) {
	var abilities C.CameraAbilities
	err := C.gp_camera_get_abilities(c.camera, &abilities)
	return abilities, int(err)
}

func (c *Camera) WaitForCameraEvent(timeout int, handler func(int, string)) {
	var eventType C.CameraEventType
	var vp unsafe.Pointer
	err := C.gp_camera_wait_for_event(c.camera, C.int(timeout), &eventType, &vp, c.context)

	if err < 0 {
		fmt.Printf(CameraResultToString(int(err)))
	}

	s := C.GoString((*C.char)(vp))
	handler(int(eventType), s)
}

func (c *Camera) Model() (string, int) {
	abilities, err := c.GetAbilities()
	modelBytes := C.GoBytes(unsafe.Pointer(&abilities.model), 255)
	model := string(modelBytes[:255])

	return model, err
}

func CameraResultToString(err int) string {
	return C.GoString(C.gp_result_as_string(C.int(err)))
}
