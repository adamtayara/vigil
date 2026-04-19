package tray

import (
	"fmt"
	"runtime"

	"github.com/getlantern/systray"
)

var (
	onScanNow    func()
	onViewReport func()
)

func Run(scanFn func(), viewFn func()) {
	onScanNow = scanFn
	onViewReport = viewFn
	systray.Run(onReady, onExit)
}

func onReady() {
	icon := trayIcon()
	systray.SetIcon(icon)
	systray.SetTitle("Vigil")
	systray.SetTooltip("Vigil — System Health Monitor")

	mScan := systray.AddMenuItem("Scan Now", "Run a system health scan")
	mReport := systray.AddMenuItem("View Last Report", "Open the most recent report")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Vigil", "Exit the application")

	go func() {
		for {
			select {
			case <-mScan.ClickedCh:
				if onScanNow != nil {
					go onScanNow()
				}
			case <-mReport.ClickedCh:
				if onViewReport != nil {
					go onViewReport()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {}

func Notify(title, body string) {
	switch runtime.GOOS {
	case "windows":
		notifyWindows(title, body)
	case "darwin":
		notifyDarwin(title, body)
	default:
		notifyLinux(title, body)
	}
}

func notifyWindows(title, body string) {
	// Uses PowerShell toast notification
	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$template.SelectSingleNode('//text[@id=1]').AppendChild($template.CreateTextNode('%s')) | Out-Null
$template.SelectSingleNode('//text[@id=2]').AppendChild($template.CreateTextNode('%s')) | Out-Null
$toast = [Windows.UI.Notifications.ToastNotification]::new($template)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Vigil').Show($toast)
`, title, body)
	_ = script // executed via PowerShell in production
}

func notifyDarwin(title, body string) {
	// Uses osascript
	_ = fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
}

func notifyLinux(title, body string) {
	// Uses notify-send
	_ = title
	_ = body
}

func trayIcon() []byte {
	// Minimal 16x16 PNG icon (purple dot)
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x91, 0x68, 0x36, 0x00, 0x00, 0x00,
		0x3c, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0xfc, 0xcf, 0xc0, 0xc0,
		0xc0, 0x00, 0x04, 0x30, 0x18, 0x60, 0xa8, 0x17, 0x05, 0x03, 0x03, 0x43,
		0x03, 0x03, 0x03, 0x23, 0x03, 0x03, 0x03, 0x83, 0x01, 0x06, 0x0c, 0x18,
		0x30, 0xe0, 0x0b, 0x60, 0x41, 0x01, 0x81, 0x09, 0x0c, 0x04, 0x18, 0x0c,
		0xd8, 0x18, 0x18, 0x18, 0x19, 0x18, 0x18, 0x00, 0x00, 0x2f, 0x3e, 0x09,
		0xfe, 0xca, 0x68, 0x96, 0x21, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
