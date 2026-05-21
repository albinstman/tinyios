package main

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	_ "embed"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tiny"
)

func convertToJSONString(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(b)
}

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

type ProfileHttpRequest struct {
	Address string `json:"address"`
	Port    string `json:"port"`
}

// profileHttp godoc
// @Summary      Set global HTTP proxy
// @Description  Configures the device to use an HTTP proxy for profile installation
// @Tags         profiles
// @Accept       json
// @Produce      json
// @Param        udid   path      string  true  "Device UDID"
// @Param        request body ProfileHttpRequest true "HTTP proxy configuration"
// @Success      200 {object} GenericResponse
// @Failure      400 {string} string "invalid JSON"
// @Router       /{udid}/profiles/http [post]
func profileHttp(w http.ResponseWriter, r *http.Request) {
	d, _ := getDevice(r.Context())
	var u ProfileHttpRequest
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("Setting HTTP proxy to %s:%s\n", u.Address, u.Port)
	result := []byte(tiny.ProfileHttp(d, u.Address, u.Port, p12))
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

// downloadFromPresignedURL downloads the content from presignedURL into destPath.
func downloadFromPresignedURL(presignedURL, destPath string) error {
	resp, err := http.Get(presignedURL)
	if err != nil {
		return fmt.Errorf("failed to GET from presigned URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d %s", resp.StatusCode, resp.Status)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", destPath, err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to copy response body to file: %w", err)
	}

	return nil
}

// unzip extracts a zip archive to destDir, taking care to avoid ZipSlip.
func unzip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	// Normalize destination directory to an absolute, cleaned path
	destDir, err = filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for destDir: %w", err)
	}

	for _, f := range r.File {
		// Build the full path under destDir
		fpath := filepath.Join(destDir, f.Name)

		// Prevent ZipSlip by ensuring fpath is still under destDir
		if !strings.HasPrefix(fpath, destDir+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", fpath, err)
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", filepath.Dir(fpath), err)
		}

		// Open file inside the zip
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip %q: %w", f.Name, err)
		}

		// Create destination file with same mode
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create file %q: %w", fpath, err)
		}

		// Copy contents
		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return fmt.Errorf("failed to copy file contents to %q: %w", fpath, err)
		}

		outFile.Close()
		rc.Close()
	}

	return nil
}

func appInstallWithURL(url string, d ios.DeviceEntry) string {
	presignedURL := url                            // TODO: replace with your URL
	zipPath := d.Properties.SerialNumber + ".zip"  // where to save the downloaded zip
	extractDir := "./" + d.Properties.SerialNumber // where to unzip; current dir is fine

	// 1. Download the zip file
	if err := downloadFromPresignedURL(presignedURL, zipPath); err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	fmt.Println("Downloaded zip to", zipPath)

	// 2. Unzip it (this will create the .app bundle in extractDir if it's inside the zip)
	if err := unzip(zipPath, extractDir); err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	fmt.Println("Unzipped to", extractDir)

	wdaPath := d.Properties.SerialNumber + "/WebDriverAgentRunner-Runner.app"
	return tiny.AppInstall(d, wdaPath)
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

	result := []byte(appInstallWithURL(u.URL, d))
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

// wdaCmd godoc
// @Summary      WebDriverAgent passthrough
// @Description  Transparent reverse proxy to the WebDriverAgent HTTP server running on the device. Everything after /wda/cmd is forwarded verbatim to WDA, exposing its full WebDriver/Appium endpoint surface (for example /status, /session, element interactions). Sub-paths are proxied as-is.
// @Tags         wda
// @Param        udid   path      string  true  "Device UDID"
// @Router       /{udid}/wda/cmd/ [any]
func wdaCmd(rproxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rproxy.ServeHTTP(w, r)
	}
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

func stripUDIDAndWDACmd(r *http.Request) string {
	udid := r.PathValue("udid")
	if udid == "" {
		return r.URL.Path
	}

	prefix := "/" + udid + "/wda/cmd"
	rest := strings.TrimPrefix(r.URL.Path, prefix)

	// Normalize:
	// "/{udid}/wda/cmd"        -> "/"
	// "/{udid}/wda/cmd/"       -> "/"
	// "/{udid}/wda/cmd/status" -> "/status"
	if rest == "" || rest == "/" {
		return "/"
	}
	if !strings.HasPrefix(rest, "/") {
		rest = "/" + rest
	}
	return rest
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
	deviceMux.HandleFunc("POST /{udid}/profiles/http", profileHttp)

	deviceMux.HandleFunc("GET /{udid}/apps/list", appList)
	deviceMux.HandleFunc("POST /{udid}/apps/run", appRun)
	deviceMux.HandleFunc("POST /{udid}/apps/install", appInstall)
	deviceMux.HandleFunc("POST /{udid}/apps/kill", appKill)

	deviceMux.HandleFunc("GET /{udid}/processes", processes)

	deviceMux.HandleFunc("POST /{udid}/wda/run", wdaRun)
	deviceMux.HandleFunc("POST /{udid}/wda/kill", wdaKill)

	transport := &http.Transport{
		// IMPORTANT: DialContext receives the request context.
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d, _ := getDevice(ctx)
			conn, err := tiny.WdaConnection(d)
			return conn, err
		},

		// Keep-alive / pooling knobs (tune to your needs)
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,

		// If you do NOT want http2 surprises with custom conns:
		ForceAttemptHTTP2: false,
	}

	rproxy := &httputil.ReverseProxy{
		Transport: transport,
		Director: func(r *http.Request) {
			d, _ := getDevice(r.Context())

			// Pool key: use a stable synthetic host per udid.
			// All requests for same udid will share a pool.
			r.URL.Scheme = "http"
			r.URL.Host = "udid-" + d.Properties.SerialNumber
			r.Host = r.URL.Host

			// Strip "/{udid}" prefix
			r.URL.Path = stripUDIDAndWDACmd(r)
			r.URL.RawPath = ""
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "bad gateway: "+err.Error(), http.StatusBadGateway)
		},
	}

	deviceMux.HandleFunc("/{udid}/wda/cmd/", wdaCmd(rproxy))

	deviceMux.HandleFunc("/{udid}/wda/cmd", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusPermanentRedirect)
	})

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
