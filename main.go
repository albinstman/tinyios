package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	_ "embed"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tiny"
)

// @title           tinyios
// @version         0.0.1

type GenericResponse struct {
	OK bool `json:"ok"`
}

//go:embed c.der
var cder []byte

//go:embed c.p12
var p12 []byte

type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	UDID           string `json:"Udid"`
	ProductName    string `json:"ProductName"`
	ProductType    string `json:"ProductType"`
	ProductVersion string `json:"ProductVersion"`
	ConnectionType string `json:"ConnectionType"`
}

type deviceCtxKey string

const deviceKey deviceCtxKey = "udid"

func withDevice(ctx context.Context, device ios.DeviceEntry) context.Context {
	return context.WithValue(ctx, deviceKey, device)
}

func getDevice(ctx context.Context) (ios.DeviceEntry, bool) {
	v := ctx.Value(deviceKey)
	id, ok := v.(ios.DeviceEntry)
	return id, ok
}

func deviceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		udid := r.PathValue("udid")

		devices := []byte(tiny.DeviceList())

		var resp DevicesResponse
		if err := json.Unmarshal(devices, &resp); err != nil {
			panic(err)
		}

		if len(resp.Devices) == 0 {
			fmt.Println("No devices found")
			return
		}

		found := false
		for _, device := range resp.Devices {
			if device.UDID == udid {
				found = true
				break
			}
		}

		if !found {
			w.Write([]byte("device not found"))
			return
		}

		d, err := ios.GetDevice(udid)
		if err != nil {
			//panic(err)
			w.Write([]byte("device not found (panic)"))
		}
		next.ServeHTTP(w, r.WithContext(withDevice(r.Context(), d)))
	})
}

func writeResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

// devices godoc
// @Summary      List devices
// @Description  Returns a list of all connected iOS devices
// @Tags         device
// @Produce      json
// @Success      200 {object} DevicesResponse
// @Router       /devices [get]
func devices(w http.ResponseWriter, _ *http.Request) {
	devices := []byte(tiny.DeviceList())
	writeResponse(w, 200, devices)
}

// reboot godoc
// @Summary      Reboot device
// @Description  Reboots the specified iOS device
// @Tags         device
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/reboot [post]
func reboot(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Reboot(d))
	writeResponse(w, 200, result)
}

// activated godoc
// @Summary      Check activation status
// @Description  Returns whether the device is activated
// @Tags         activation
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/activated [get]
func activated(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Activated(d))
	writeResponse(w, 200, result)
}

// activateEnable godoc
// @Summary      Enable activation
// @Description  Activates the specified iOS device
// @Tags         activation
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/activate/enable [post]
func activateEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.ActivateEnable(d))
	writeResponse(w, 200, result)
}

// supervised godoc
// @Summary      Check supervision status
// @Description  Returns whether the device is supervised
// @Tags         supervision
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/supervised [get]
func supervised(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Supervised(d))
	writeResponse(w, 200, result)
}

// superviseEnable godoc
// @Summary      Enable supervision
// @Description  Prepares and enables supervision on the device
// @Tags         supervision
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/supervise/enable [post]
func superviseEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Prepare(d, cder, "tinyios", "en-US", "en"))
	writeResponse(w, 200, result)
}

// erase godoc
// @Summary      Erase device
// @Description  Erases all content and settings from the device
// @Tags         device
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/erase [post]
func erase(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Erase(d))
	writeResponse(w, 200, result)
}

// paired godoc
// @Summary      Check pairing status
// @Description  Returns whether the device is paired
// @Tags         pairing
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/paired [get]
func paired(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Paired(d))
	writeResponse(w, 200, result)
}

// pairEnable godoc
// @Summary      Enable pairing
// @Description  Pairs the device using the provided certificate
// @Tags         pairing
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/pair/enable [post]
func pairEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.PairEnable(d, p12))
	writeResponse(w, 200, result)
}

// devmode godoc
// @Summary      Check developer mode status
// @Description  Returns whether developer mode is enabled on the device
// @Tags         developer
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/devmode [get]
func devmode(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Devmode(d))
	writeResponse(w, 200, result)
}

// devmodeEnable godoc
// @Summary      Enable developer mode
// @Description  Enables developer mode on the device
// @Tags         developer
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/devmode/enable [post]
func devmodeEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.DevmodeEnable(d))
	writeResponse(w, 200, result)
}

// image godoc
// @Summary      Check developer disk image status
// @Description  Returns whether the developer disk image is mounted
// @Tags         developer
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/image [get]
func image(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Image(d))
	writeResponse(w, 200, result)
}

// imageEnable godoc
// @Summary      Mount developer disk image
// @Description  Mounts the developer disk image on the device
// @Tags         developer
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/image/enable [post]
func imageEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.ImageEnable(d))
	writeResponse(w, 200, result)
}

// profileList godoc
// @Summary      List profiles
// @Description  Returns a list of configuration profiles installed on the device
// @Tags         profiles
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/profiles/list [get]
func profileList(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.ProfileList(d))
	writeResponse(w, 200, result)
}

type ProfileAddRequest struct {
	B64Profile string `json:"b64profile"`
}

// profileAdd godoc
// @Summary      Add profile
// @Description  Installs a configuration profile on the device
// @Tags         profiles
// @Accept       json
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Param        profile body ProfileAddRequest true "Base64 encoded profile"
// @Success      200 {object} GenericResponse
// @Failure      400 {string} string "invalid JSON"
// @Router       /{udid}/profiles/add [post]
func profileAdd(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())

	// Decode JSON
	var u ProfileAddRequest
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	data, err := base64.StdEncoding.DecodeString(u.B64Profile)
	if err != nil {
		http.Error(w, "invalid base64 profile: "+err.Error(), http.StatusBadRequest)
		return
	}

	result := []byte(tiny.ProfileAdd(d, data, p12))
	writeResponse(w, 200, result)
}

// appList godoc
// @Summary      List applications
// @Description  Returns a list of applications installed on the device
// @Tags         apps
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/apps/list [get]
func appList(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.AppList(d))
	writeResponse(w, 200, result)
}

// appRun godoc
// @Summary      Run application
// @Description  Launches an application on the device
// @Tags         apps
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Param        bundleid formData string true "Application bundle identifier"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/apps/run [post]
func appRun(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.AppRun(d, r.FormValue("bundleid")))
	writeResponse(w, 200, result)
}

type AppInstallRequest struct {
	URL string `json:"url"`
}

// appInstall godoc
// @Summary      Install application
// @Description  Installs an application from a URL on the device
// @Tags         apps
// @Accept       json
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Param        request body AppInstallRequest true "Application IPA URL"
// @Success      200 {object} GenericResponse
// @Failure      400 {string} string "invalid JSON"
// @Router       /{udid}/apps/install [post]
func appInstall(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())

	var u AppInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	result := []byte(tiny.AppInstall(d, u.URL))
	writeResponse(w, 200, result)
}

// appKill godoc
// @Summary      Kill application
// @Description  Terminates a running application by process ID
// @Tags         apps
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Param        pid formData string true "Process ID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/apps/kill [post]
func appKill(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.AppKill(d, r.FormValue("pid")))
	writeResponse(w, 200, result)
}

// processes godoc
// @Summary      List processes
// @Description  Returns a list of running processes on the device
// @Tags         device
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/processes [get]
func processes(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Processes(d))
	writeResponse(w, 200, result)
}

// wdaRun godoc
// @Summary      Run WebDriverAgent
// @Description  Starts WebDriverAgent on the device
// @Tags         wda
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/wda/run [post]
func wdaRun(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.WdaRun(d))
	writeResponse(w, 200, result)
}

// wdaKill godoc
// @Summary      Kill WebDriverAgent
// @Description  Stops WebDriverAgent on the device
// @Tags         wda
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Success      200 {object} GenericResponse
// @Router       /{udid}/wda/kill [post]
func wdaKill(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.WdaKill(d))
	writeResponse(w, 200, result)
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v\n%s", rec, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	proxyUrl := os.Getenv("HTTP_PROXY")
	if os.Getenv("HTTPS_PROXY") != "" {
		proxyUrl = os.Getenv("HTTPS_PROXY")
	}

	if proxyUrl != "" {
		parsedUrl, err := url.Parse(proxyUrl)
		if err != nil {
			log.Fatalf("could not parse proxy url %s: %v", proxyUrl, err)
		}
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(parsedUrl)}
	}

	root := http.NewServeMux()
	root.HandleFunc("GET /devices", devices)

	deviceMux := http.NewServeMux()
	deviceMux.HandleFunc("POST /{udid}/reboot", reboot)
	deviceMux.HandleFunc("GET /{udid}/activated", activated)
	deviceMux.HandleFunc("POST /{udid}/activate/enable", activateEnable)
	deviceMux.HandleFunc("GET /{udid}/supervised", supervised)
	deviceMux.HandleFunc("POST /{udid}/supervise/enable", superviseEnable)
	deviceMux.HandleFunc("POST /{udid}/erase", erase)
	deviceMux.HandleFunc("GET /{udid}/paired", paired)
	deviceMux.HandleFunc("POST /{udid}/pair/enable", pairEnable)
	deviceMux.HandleFunc("GET /{udid}/devmode", devmode)
	deviceMux.HandleFunc("POST /{udid}/devmode/enable", devmodeEnable)
	deviceMux.HandleFunc("GET /{udid}/image", image)
	deviceMux.HandleFunc("POST /{udid}/image/enable", imageEnable)
	deviceMux.HandleFunc("GET /{udid}/profiles/list", profileList)
	deviceMux.HandleFunc("POST /{udid}/profiles/add", profileAdd)
	deviceMux.HandleFunc("GET /{udid}/apps/list", appList)
	deviceMux.HandleFunc("POST /{udid}/apps/run", appRun)
	deviceMux.HandleFunc("POST /{udid}/apps/install", appInstall)
	deviceMux.HandleFunc("POST /{udid}/apps/kill", appKill)
	deviceMux.HandleFunc("GET /{udid}/processes", processes)
	deviceMux.HandleFunc("POST /{udid}/wda/run", wdaRun)
	deviceMux.HandleFunc("POST /{udid}/wda/kill", wdaKill)

	root.Handle("/{udid}/", deviceMiddleware(deviceMux))

	var handler http.Handler = root
	handler = RecoveryMiddleware(handler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Wait for signal
	<-stop
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed:%+v", err)
	}

	log.Println("Server exited cleanly")
}
