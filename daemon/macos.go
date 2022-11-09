package daemon

import (
	"log"
	"os"
	"time"

	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/netclient/ncutils"
)

const MacServiceName = "com.gravitl.netclient"
const MacExecDir = "/usr/local/bin/"

// SetupMacDaemon - Creates a daemon service from the netclient under LaunchAgents for MacOS
func SetupMacDaemon() error {

	binarypath, err := os.Executable()
	if err != nil {
		return err
	}

	if ncutils.FileExists(MacExecDir + "netclient") {
		logger.Log(0, "updating netclient binary in", MacExecDir)
	}
	err = ncutils.Copy(binarypath, MacExecDir+"netclient")
	if err != nil {
		logger.Log(0, err.Error())
		return err
	}

	err = CreateMacService(MacServiceName)
	if err != nil {
		return err
	}
	_, err = ncutils.RunCmd("launchctl load /Library/LaunchDaemons/"+MacServiceName+".plist", true)
	return err
}

// CleanupMac - Removes the netclient checkin daemon from LaunchDaemons
func CleanupMac() {
	_, err := ncutils.RunCmd("launchctl unload /Library/LaunchDaemons/"+MacServiceName+".plist", true)
	if ncutils.FileExists("/Library/LaunchDaemons/" + MacServiceName + ".plist") {
		err = os.Remove("/Library/LaunchDaemons/" + MacServiceName + ".plist")
	}
	if err != nil {
		logger.Log(1, err.Error())
	}

	os.RemoveAll(ncutils.GetNetclientPath())
	os.Remove(MacExecDir + "netclient")
}

// RestartLaunchD - restart launch daemon
func RestartLaunchD() {
	ncutils.RunCmd("launchctl unload /Library/LaunchDaemons/"+MacServiceName+".plist", true)
	time.Sleep(time.Second >> 2)
	ncutils.RunCmd("launchctl load /Library/LaunchDaemons/"+MacServiceName+".plist", true)
}

// StopLaunchD - stop launch daemon
func StopLaunchD() {
	ncutils.RunCmd("launchctl unload  /Library/LaunchDaemons/"+MacServiceName+".plist", true)
}

// CreateMacService - Creates the mac service file for LaunchDaemons
func CreateMacService(servicename string) error {
	_, err := os.Stat("/Library/LaunchDaemons")
	if os.IsNotExist(err) {
		os.Mkdir("/Library/LaunchDaemons", 0755)
	} else if err != nil {
		log.Println("couldnt find or create /Library/LaunchDaemons")
		return err
	}
	daemonstring := MacDaemonString()
	daemonbytes := []byte(daemonstring)

	if !ncutils.FileExists("/Library/LaunchDaemons/com.gravitl.netclient.plist") {
		err = os.WriteFile("/Library/LaunchDaemons/com.gravitl.netclient.plist", daemonbytes, 0644)
	}
	return err
}

// MacDaemonString - the file contents for the mac netclient daemon service (launchdaemon)
func MacDaemonString() string {
	return `<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\" >
<plist version='1.0'>
<dict>
	<key>Label</key><string>com.gravitl.netclient</string>
	<key>ProgramArguments</key>
		<array>
			<string>/usr/local/bin/netclient</string>
			<string>daemon</string>
		</array>
	<key>StandardOutPath</key><string>/var/log/com.gravitl.netclient.log</string>
	<key>StandardErrorPath</key><string>/var/log/com.gravitl.netclient.log</string>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>AbandonProcessGroup</key><true/>
	<key>EnvironmentVariables</key>
		<dict>
			<key>PATH</key>
			<string>/usr/local/sbin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
		</dict>
</dict>
</plist>
`
}

// MacTemplateData - struct to represent the mac service
type MacTemplateData struct {
	Label    string
	Interval string
}
