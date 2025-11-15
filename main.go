package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "embed"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tiny"
)

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

func reboot(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Reboot(d))
	w.Write(result)
}

func activated(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Activated(d))
	w.Write(result)
}

func activateEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.ActivateEnable(d))
	w.Write(result)
}

func supervised(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Supervised(d))
	w.Write(result)
}

func superviseEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Prepare(d, cder, "tinyios", "en-US", "en"))
	w.Write(result)
}

func erase(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Erase(d))
	w.Write(result)
}

func paired(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Paired(d))
	w.Write(result)
}

func pairEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.PairEnable(d, p12))
	w.Write(result)
}

func devmode(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Devmode(d))
	w.Write(result)
}

func devmodeEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.DevmodeEnable(d))
	w.Write(result)
}

func image(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Image(d))
	w.Write(result)
}

func imageEnable(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.ImageEnable(d))
	w.Write(result)
}

func profileList(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.ProfileList(d))
	w.Write(result)
}

func profileAdd(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.ProfileAdd(d, []byte(r.FormValue("b64profile")), p12))
	w.Write(result)
}

func appList(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.AppList(d))
	w.Write(result)
}

func appRun(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.AppRun(d, r.FormValue("bundleid")))
	w.Write(result)
}

func appInstall(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.AppInstall(d, r.FormValue("url")))
	w.Write(result)
}

func appKill(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.AppKill(d, r.FormValue("pid")))
	w.Write(result)
}

func processes(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.Processes(d))
	w.Write(result)
}

func wdaRun(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.WdaRun(d))
	w.Write(result)
}

func wdaKill(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	result := []byte(tiny.WdaKill(d))
	w.Write(result)
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	root := http.NewServeMux()
	root.HandleFunc("GET /devices", func(w http.ResponseWriter, r *http.Request) {
		devices := []byte(tiny.DeviceList())
		w.Write(devices)
	})

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
