package gphoto2

// #cgo CFLAGS: -I/Users/micah/Developer/include/gphoto2
// #cgo LDFLAGS: -L/Users/micah/Developer/lib -lgphoto2
// #include <gphoto2.h>
// #include <stdlib.h>
import "C"
import "unsafe"
import "fmt"
import "strings"
import "io"
import "bufio"

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

func (c *Camera) TriggerCapture() int {
	err := C.gp_camera_trigger_capture(c.camera, c.context)
	return int(err)
}

type CameraEventType int

const (
	EVENT_UKNOWN     CameraEventType = C.GP_EVENT_UNKNOWN
	EVENT_TIMEOUT    CameraEventType = C.GP_EVENT_TIMEOUT
	EVENT_FILE_ADDED CameraEventType = C.GP_EVENT_FILE_ADDED
)

type CameraEvent struct {
	Type   CameraEventType
	Folder string
	File   string
}

func (c *Camera) AsyncWaitForEvent(timeout int) chan *CameraEvent {
	var eventType C.CameraEventType
	var vp unsafe.Pointer
	defer C.free(vp)

	ch := make(chan *CameraEvent)

	go func() {
		C.gp_camera_wait_for_event(c.camera, C.int(timeout), &eventType, &vp, c.context)
		ch <- cCameraEventToGoCameraEvent(vp, eventType)
	}()

	return ch
}

func cCameraEventToGoCameraEvent(voidPtr unsafe.Pointer, eventType C.CameraEventType) *CameraEvent {
	ce := new(CameraEvent)
	ce.Type = CameraEventType(eventType)

	if ce.Type == EVENT_FILE_ADDED {
		cameraFilePath := (*C.CameraFilePath)(voidPtr)
		ce.File = C.GoString((*C.char)(&cameraFilePath.name[0]))
		ce.Folder = C.GoString((*C.char)(&cameraFilePath.folder[0]))
	}

	return ce
}

func (c *Camera) ListFolders(folder string) ([]string, int) {
	if folder == "" {
		folder = "/"
	}

	var cameraList *C.CameraList
	C.gp_list_new(&cameraList)
	defer C.free(unsafe.Pointer(cameraList))

	cFolder := C.CString(folder)
	defer C.free(unsafe.Pointer(cFolder))

	err := C.gp_camera_folder_list_folders(c.camera, cFolder, cameraList, c.context)
	folderMap, _ := cameraListToMap(cameraList)

	names := make([]string, len(folderMap))
	i := 0
	for key, _ := range folderMap {
		names[i] = key
		i += 1
	}

	return names, int(err)
}

func (c *Camera) RListFolders(folder string) []string {
	folders := make([]string, 0)
	path := folder
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	subfolders, _ := c.ListFolders(path)
	for _, sub := range subfolders {
		subPath := path + sub
		folders = append(folders, subPath)
		folders = append(folders, c.RListFolders(subPath)...)
	}

	return folders
}

func (c *Camera) ListFiles(folder string) ([]string, int) {
	if folder == "" {
		folder = "/"
	}

	if !strings.HasSuffix(folder, "/") {
		folder = folder + "/"
	}

	var cameraList *C.CameraList
	C.gp_list_new(&cameraList)
	defer C.free(unsafe.Pointer(cameraList))

	cFolder := C.CString(folder)
	defer C.free(unsafe.Pointer(cFolder))

	err := C.gp_camera_folder_list_files(c.camera, cFolder, cameraList, c.context)
	fileNameMap, _ := cameraListToMap(cameraList)

	names := make([]string, len(fileNameMap))
	i := 0
	for key, _ := range fileNameMap {
		names[i] = key
		i += 1
	}

	return names, int(err)
}

func cameraListToMap(cameraList *C.CameraList) (map[string]string, int) {
	size := int(C.gp_list_count(cameraList))
	vals := make(map[string]string)

	if size < 0 {
		return vals, size
	}

	for i := 0; i < size; i++ {
		var cKey *C.char
		var cVal *C.char

		C.gp_list_get_name(cameraList, C.int(i), &cKey)
		C.gp_list_get_value(cameraList, C.int(i), &cVal)
		defer C.free(unsafe.Pointer(cKey))
		defer C.free(unsafe.Pointer(cVal))
		key := C.GoString(cKey)
		val := C.GoString(cVal)

		vals[key] = val
	}

	return vals, 0
}

func (c *Camera) Model() (string, int) {
	abilities, err := c.GetAbilities()
	model := C.GoString((*C.char)(&abilities.model[0]))

	return model, err
}

func CameraResultToString(err int) string {
	return C.GoString(C.gp_result_as_string(C.int(err)))
}

// Need to find a good buffer size
// For now, let's try 1MB
const fileReaderBufferSize = 1 * 1024 * 1024

type cameraFileReader struct {
	camera   *Camera
	folder   string
	fileName string
	fullSize uint64
	offset   uint64

	buffer [fileReaderBufferSize]byte
}

func (cfr *cameraFileReader) Read(p []byte) (int, error) {
	n := uint64(len(p))

	if n == 0 {
		return 0, nil
	}

	bufLen := uint64(len(cfr.buffer))
	remaining := cfr.fullSize - cfr.offset

	toRead := bufLen
	if toRead > remaining {
		toRead = remaining
	}

	if toRead > n {
		toRead = n
	}

	cFileName := C.CString(cfr.fileName)
	cFolderName := C.CString(cfr.folder)
	defer C.free(unsafe.Pointer(cFileName))
	defer C.free(unsafe.Pointer(cFolderName))

	cBuffer := (*C.char)(C.malloc(C.size_t(toRead)))
	defer C.free(unsafe.Pointer(cBuffer))

	cOffset := C.uint64_t(cfr.offset)

	cToRead := C.uint64_t(toRead)

	err := C.gp_camera_file_read(cfr.camera.camera, cFolderName, cFileName, C.GP_FILE_TYPE_NORMAL, cOffset, cBuffer, &cToRead, cfr.camera.context)
	if err < 0 {
		fmt.Printf("File Read Error: %s\n", CameraResultToString(int(err)))
	}

	amountRead := int(cToRead)

	// Note: This is a double (triple?) buffering performance issue. It would be nice to write directly to the byte slice, but I'm not sure how to do that safely
	bytes := C.GoBytes(unsafe.Pointer(cBuffer), C.int(amountRead))
	for i, b := range bytes {
		p[i] = b
	}

	cfr.offset += uint64(cToRead)

	if cfr.offset < cfr.fullSize {
		return amountRead, nil
	} else {
		return amountRead, io.EOF
	}
}

func (c *Camera) FileReader(folder string, fileName string) io.Reader {
	cfr := new(cameraFileReader)
	cfr.camera = c
	cfr.folder = folder
	cfr.fileName = fileName
	cfr.offset = 0

	cFileName := C.CString(cfr.fileName)
	cFolderName := C.CString(cfr.folder)
	defer C.free(unsafe.Pointer(cFileName))
	defer C.free(unsafe.Pointer(cFolderName))

	var cCameraFileInfo C.CameraFileInfo

	err := C.gp_camera_file_get_info(c.camera, cFolderName, cFileName, &cCameraFileInfo, c.context)
	if err < 0 {
		fmt.Printf("File Get Info Error: %s\n", CameraResultToString(int(err)))
	}

	cfr.fullSize = uint64(cCameraFileInfo.file.size)

	return cfr
}

func (c *Camera) BufferedFileReader(folder string, fileName string) io.Reader {
	return bufio.NewReaderSize(c.FileReader(folder, fileName), fileReaderBufferSize)
}
