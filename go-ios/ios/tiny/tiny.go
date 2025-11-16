package tiny

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/amfi"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/danielpaulus/go-ios/ios/zipconduit"
	"golang.org/x/sync/errgroup"

	log "github.com/sirupsen/logrus"
)

type WdaSession struct {
	Udid    string
	stopWda context.CancelFunc
}

func (session *WdaSession) Write(p []byte) (n int, err error) {
	log.
		WithField("udid", session.Udid).
		Debugf("WDA_LOG %s", p)

	return len(p), nil
}

var globalSessions = sync.Map{}

func convertToJSONString(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(b)
}

func Reboot(device ios.DeviceEntry) string {
	err := diagnostics.Reboot(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func DeviceList() string {
	dl, _ := ios.DetailedList()
	return dl
}

func GetDevice(udid string) (ios.DeviceEntry, error) {
	return ios.GetDevice(udid)

}

func Activated(device ios.DeviceEntry) string {
	activated, err := mobileactivation.IsActivated(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"activated": false})
	}
	return convertToJSONString(map[string]bool{"activated": activated})
}

func ActivateEnable(device ios.DeviceEntry) string {
	err := mobileactivation.Activate(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func Supervised(device ios.DeviceEntry) string {
	conn, err := ios.ConnectLockdownWithSession(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"supervised": false})
	}
	defer conn.Close()
	v, _ := conn.GetValueForDomain("DeviceIsChaperoned", "com.apple.mobile.chaperone")
	return convertToJSONString(map[string]bool{"supervised": v.(bool)})
}

func Prepare(device ios.DeviceEntry, cder []byte, orgname string, locale string, lang string) string {
	skip := mcinstall.GetAllSetupSkipOptions()
	if orgname == "" {
		orgname = "ios"
	}
	err := mcinstall.Prepare(device, skip, cder, orgname, locale, lang)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func Erase(device ios.DeviceEntry) string {
	err := mcinstall.Erase(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func Paired(device ios.DeviceEntry) string {
	_, err := ios.ReadPairRecord(device.Properties.SerialNumber)
	if err != nil {
		return convertToJSONString(map[string]bool{"paired": false})
	}
	return convertToJSONString(map[string]bool{"paired": true})
}

func PairEnable(device ios.DeviceEntry, p12 []byte) string {
	err := ios.PairSupervised(device, p12, "a")
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func Devmode(device ios.DeviceEntry) string {
	enabled, _ := imagemounter.IsDevModeEnabled(device)
	return convertToJSONString(map[string]bool{"devmode": enabled})
}

func DevmodeEnable(device ios.DeviceEntry) string {
	err := amfi.EnableDeveloperMode(device, true)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func Image(device ios.DeviceEntry) string {
	conn, err := imagemounter.NewImageMounter(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"devimage": false})
	}
	signatures, err := conn.ListImages()
	if err != nil {
		return convertToJSONString(map[string]bool{"devimage": false})
	}
	if len(signatures) == 0 {
		return convertToJSONString(map[string]bool{"devimage": false})
	}
	return convertToJSONString(map[string]bool{"devimage": true})
}

func ImageEnable(device ios.DeviceEntry) string {
	basedir := "./devimages"
	var err error
	path, err := imagemounter.DownloadImageFor(device, basedir)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	err = imagemounter.MountImage(device, path)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func ProfileList(device ios.DeviceEntry) string {
	profileService, err := mcinstall.New(device)
	if err != nil {
		return convertToJSONString(map[string][]any{
			"profiles": {},
		})
	}
	list, err := profileService.HandleList()
	if err != nil {
		return convertToJSONString(map[string][]any{
			"profiles": {},
		})
	}
	return convertToJSONString(map[string][]mcinstall.ProfileInfo{
		"profiles": list,
	})
}

func ProfileAdd(device ios.DeviceEntry, profileData []byte, p12 []byte) string {
	profileService, err := mcinstall.New(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	err = profileService.AddProfileSupervised(profileData, p12, "a")
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func AppList(device ios.DeviceEntry) string {
	svc, err := installationproxy.New(device)
	if err != nil {
		return convertToJSONString(map[string][]any{
			"apps": {},
		})
	}
	response, err := svc.BrowseUserApps()
	if err != nil {
		return convertToJSONString(map[string][]any{
			"apps": {},
		})
	}
	return convertToJSONString(map[string][]installationproxy.AppInfo{
		"apps": response,
	})
}

func AppInstall(device ios.DeviceEntry, path string) string {
	conn, err := zipconduit.New(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	err = conn.SendFile(path)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func Processes(device ios.DeviceEntry) string {
	service, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		return convertToJSONString(map[string][]any{
			"processes": {},
		})
	}

	defer service.Close()
	processList, err := service.ProcessList()
	if err != nil {
		return convertToJSONString(map[string][]any{
			"processes": {},
		})
	}
	var applicationProcessList []instruments.ProcessInfo
	for _, processInfo := range processList {
		if processInfo.IsApplication {
			applicationProcessList = append(applicationProcessList, processInfo)
		}
	}
	return convertToJSONString(map[string][]instruments.ProcessInfo{
		"processes": applicationProcessList,
	})
}

func AppRun(device ios.DeviceEntry, bundleID string) string {
	pControl, err := instruments.NewProcessControl(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	defer pControl.Close()
	opts := map[string]any{}
	opts["KillExisting"] = 1
	_, err = pControl.LaunchAppWithArgs(bundleID, []any{}, map[string]any{}, opts)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func AppKill(device ios.DeviceEntry, bundleID string) string {
	if bundleID == "" {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	svc, err := installationproxy.New(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	defer svc.Close()
	response, err := svc.BrowseAllApps()
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	var processName string
	for _, app := range response {
		if app.CFBundleIdentifier() == bundleID {
			processName = app.CFBundleExecutable()
			break
		}
	}
	if processName == "" {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	service, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	defer service.Close()
	processList, err := service.ProcessList()
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	pControl, err := instruments.NewProcessControl(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	defer pControl.Close()
	for _, p := range processList {
		if p.Name == processName {
			err = pControl.KillProcess(p.Pid)
			if err != nil {
				return convertToJSONString(map[string]bool{"ok": false})
			}
			break
		}
	}
	return convertToJSONString(map[string]bool{"ok": true})
}

func WdaRun(device ios.DeviceEntry) string {
	var bundleID, testbundleID, xctestconfig string
	svc, err := installationproxy.New(device)
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	defer svc.Close()

	response, err := svc.BrowseAllApps()
	if err != nil {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	for _, app := range response {
		if strings.Contains(app.CFBundleIdentifier(), "WebDriverAgentRunner") {
			bundleID = app.CFBundleIdentifier()
			testbundleID = app.CFBundleIdentifier()
			xctestconfig = "WebDriverAgentRunner.xctest"
			break
		}
	}

	if bundleID == "" || testbundleID == "" || xctestconfig == "" {
		return convertToJSONString(map[string]bool{"ok": false})
	}

	wdaCtx, stopWda := context.WithCancel(context.Background())

	session := &WdaSession{
		Udid:    device.Properties.SerialNumber,
		stopWda: stopWda,
	}

	go func() {
		_, err := testmanagerd.RunTestWithConfig(wdaCtx, testmanagerd.TestConfig{
			BundleId:           bundleID,
			TestRunnerBundleId: testbundleID,
			XctestConfigName:   xctestconfig,
			Env:                map[string]any{},
			Args:               []string{},
			Device:             device,
			Listener:           testmanagerd.NewTestListener(session, session, os.TempDir()),
		})
		if err != nil {
			log.
				WithField("udid", session.Udid).
				WithError(err).
				Error("Failed running WDA")
		}
		globalSessions.Delete(device.Properties.SerialNumber)
	}()

	globalSessions.Store(device.Properties.SerialNumber, session)

	return convertToJSONString(map[string]bool{"ok": true})
}

func WdaKill(device ios.DeviceEntry) string {
	session, loaded := globalSessions.Load(device.Properties.SerialNumber)
	if !loaded {
		return convertToJSONString(map[string]bool{"ok": false})
	}

	wdaSession, ok := session.(*WdaSession)
	if !ok {
		return convertToJSONString(map[string]bool{"ok": false})
	}
	wdaSession.stopWda()

	log.
		WithField("udid", wdaSession.Udid).
		Debug("Requested to stop WDA")

	return convertToJSONString(map[string]bool{"ok": true})
}

func Forward(upstream, downstream net.Conn) error {
	var (
		g    errgroup.Group
		once sync.Once
	)
	closeBoth := func() { // executed once
		upstream.Close() // closes read+write, unblocking the peer goroutine
		downstream.Close()
	}

	// Copy upstream → downstream
	g.Go(func() error {
		defer once.Do(closeBoth)
		_, err := io.Copy(downstream, upstream)
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("error copying upstream->downstream: %w", err)
	})

	// Copy downstream → upstream
	g.Go(func() error {
		defer once.Do(closeBoth)
		_, err := io.Copy(upstream, downstream)
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("error copying downstream->upstream: %w", err)
	})

	// Wait() returns the first non-nil error (if any)
	return g.Wait()
}
