package atlas

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func serviceName() string {
	sum := sha256.Sum256([]byte(Home()))
	return "atlas-" + hex.EncodeToString(sum[:4])
}
func InstalledExecutable() (string, error) {
	source, e := os.Executable()
	if e != nil {
		return "", e
	}
	source, e = filepath.EvalSymlinks(source)
	if e != nil {
		return "", e
	}
	name := "atlas"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	target := filepath.Join(Home(), "bin", name)
	if source == target {
		return target, nil
	}
	if e = PrivateDir(filepath.Dir(target)); e != nil {
		return "", e
	}
	if _, e = os.Stat(target); e == nil {
		a, ae := exec.Command(source, "version").Output()
		b, be := exec.Command(target, "version").Output()
		fromInfo,fe:=os.Stat(source);toInfo,te:=os.Stat(target)
		if ae == nil && be == nil && fe==nil&&te==nil&&fromInfo.Size()==toInfo.Size()&&bytes.Equal(a, b) {
			return target, nil
		}
	}
	from, e := os.Open(source)
	if e != nil {
		return "", e
	}
	defer from.Close()
	tmp, e := os.CreateTemp(filepath.Dir(target), ".atlas-bin-*")
	if e != nil {
		return "", e
	}
	path := tmp.Name()
	defer os.Remove(path)
	_, e = io.Copy(tmp, from)
	if e == nil {
		e = tmp.Chmod(0700)
	}
	closeErr := tmp.Close()
	if e != nil {
		return "", e
	}
	if closeErr != nil {
		return "", closeErr
	}
	if e = os.Rename(path, target); e != nil {
		return "", e
	}
	return target, nil
}
func xmlText(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))
	return b.String()
}

// User services retain the receiver across logins and restart failed processes.
// No administrator access, system proxy changes, or cloud scheduler is needed.
func InstallReceiver() error {
	if _,e:=proxyConfig();e!=nil{return e}
	executable, e := InstalledExecutable()
	if e != nil {
		return e
	}
	name := serviceName()
	h, e := os.UserHomeDir()
	if e != nil {
		return e
	}
	switch runtime.GOOS {
	case "darwin":
		label := "top.atlas." + name
		path := filepath.Join(h, "Library", "LaunchAgents", label+".plist")
		if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
			return e
		}
		data := `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>Label</key><string>` + label + `</string><key>ProgramArguments</key><array><string>` + xmlText(executable) + `</string><string>receive</string></array><key>EnvironmentVariables</key><dict><key>ATLAS_HOME</key><string>` + xmlText(Home()) + `</string></dict><key>RunAtLoad</key><true/><key>KeepAlive</key><dict><key>SuccessfulExit</key><false/></dict><key>ThrottleInterval</key><integer>3</integer><key>StandardOutPath</key><string>` + xmlText(filepath.Join(Home(), "receiver.log")) + `</string><key>StandardErrorPath</key><string>` + xmlText(filepath.Join(Home(), "receiver.log")) + `</string></dict></plist>`
		if e = os.WriteFile(path, []byte(data), 0600); e != nil {
			return e
		}
		domain := "gui/" + strconv.Itoa(os.Getuid())
		check := exec.Command("launchctl", "print", domain+"/"+label)
		if check.Run()==nil{exec.Command("launchctl","bootout",domain+"/"+label).Run()}
		if state, e := Load(filepath.Join(Home(), "receiver-status.json")); e == nil && N(state, "pid") > 0 {
			heartbeat, e := time.Parse(time.RFC3339, S(state, "heartbeat"))
			if e == nil && time.Since(heartbeat) >= 0 && time.Since(heartbeat) < 10*time.Second {
				if process, e := os.FindProcess(N(state, "pid")); e == nil {
					process.Signal(os.Interrupt)
				}
			}
		}
		output, e := exec.Command("launchctl", "bootstrap", domain, path).CombinedOutput()
		if e != nil {
			return fmt.Errorf("receiver startup registration: %w: %s", e, strings.TrimSpace(string(output)))
		}
	case "linux":
		path := filepath.Join(h, ".config", "systemd", "user", name+".service")
		if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
			return e
		}
		quote := func(s string) string {
			return `"` + strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`), `%`, `%%`) + `"`
		}
		data := "[Unit]\nDescription=Atlas team receiver\nAfter=network-online.target\n[Service]\nExecStart=" + quote(executable) + " receive\nEnvironment=" + quote("ATLAS_HOME="+Home()) + "\nRestart=on-failure\nRestartSec=3\n[Install]\nWantedBy=default.target\n"
		if e = os.WriteFile(path, []byte(data), 0600); e != nil {
			return e
		}
		for _, args := range [][]string{{"--user", "daemon-reload"}, {"--user", "enable", "--now", name + ".service"}} {
			if output, e := exec.Command("systemctl", args...).CombinedOutput(); e != nil {
				return fmt.Errorf("receiver user service: %w: %s", e, strings.TrimSpace(string(output)))
			}
		}
	case "windows":
		path := filepath.Join(Home(), "receiver-task.xml")
		who := os.Getenv("USERNAME")
		if domain := os.Getenv("USERDOMAIN"); domain != "" {
			who = domain + `\` + who
		}
		data := `<?xml version="1.0" encoding="UTF-8"?><Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task"><Triggers><LogonTrigger><Enabled>true</Enabled><UserId>` + xmlText(who) + `</UserId></LogonTrigger></Triggers><Principals><Principal id="Author"><UserId>` + xmlText(who) + `</UserId><LogonType>InteractiveToken</LogonType><RunLevel>LeastPrivilege</RunLevel></Principal></Principals><Settings><MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy><DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries><StopIfGoingOnBatteries>false</StopIfGoingOnBatteries><ExecutionTimeLimit>PT0S</ExecutionTimeLimit><RestartOnFailure><Interval>PT1M</Interval><Count>999</Count></RestartOnFailure></Settings><Actions Context="Author"><Exec><Command>` + xmlText(executable) + `</Command><Arguments>service run --home &quot;` + xmlText(Home()) + `&quot;</Arguments><WorkingDirectory>` + xmlText(Home()) + `</WorkingDirectory></Exec></Actions></Task>`
		if e = os.WriteFile(path, []byte(data), 0600); e != nil {
			return e
		}
		output, e := exec.Command("schtasks", "/Create", "/F", "/TN", name, "/XML", path).CombinedOutput()
		if e != nil {
			return fmt.Errorf("receiver login task: %w: %s", e, strings.TrimSpace(string(output)))
		}
		if output,e:=exec.Command("schtasks","/Run","/TN",name).CombinedOutput();e!=nil{return fmt.Errorf("receiver login task start: %w: %s",e,strings.TrimSpace(string(output)))}
	}
	return Save(filepath.Join(Home(),"receiver-service.json"),Object{"managed":true,"name":name})
}
