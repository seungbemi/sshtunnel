package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gofred"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	noSubtitle     = ""
	noArg          = ""
	noAutocomplete = ""
)

type host struct {
	RemoteUser   string   `json:"RemoteUser"`
	RemoteHost   string   `json:"RemoteHost"`
	RemotePort   string   `json:"RemotePort"`
	ForwardPorts []string `json:"ForwardPorts"`
}

// Config includes
type Config struct {
	host                  `json:",inline"`
	ServerAliveInterval   int    `json:"ServerAliveInterval"`
	ServerAliveCountMax   int    `json:"ServerAliveCountMax"`
	StrictHostKeyChecking string `json:"StrictHostKeyChecking"`
	LocalUser             string `json:"LocalUser"`
	IdentityFile          string `json:"IdentityFile"`
	LocalBindAddress      string `json:"LocalBindAddress"`
}

// Message adds simple message
func Message(response *gofred.Response, title, subtitle string, err bool) {
	msg := gofred.NewItem(title, subtitle, noAutocomplete)
	// if err {
	// 	msg = msg.AddIcon(iconError, defaultIconType)
	// } else {
	// 	msg = msg.AddIcon(iconDone, defaultIconType)
	// }
	response.AddItems(msg)
}

func init() {
	flag.Parse()
}

const configFolder = "conf"

func main() {
	path := os.Getenv("PATH")
	if !strings.Contains(path, "/usr/local/bin") {
		os.Setenv("PATH", path+":/usr/local/bin")
	}
	response := gofred.NewResponse()
	configs, err := ioutil.ReadDir(configFolder)
	if err != nil {
		Message(response, "error", err.Error(), true)
		return
	}
	loopback, err := exec.Command("bash", "-c", "ifconfig | grep 127.0.0 | awk '{print $2}'").CombinedOutput()
	if err != nil {
		Message(response, "error", err.Error(), true)
		return
	}

	lists := strings.Fields(string(loopback))

	items := []gofred.Item{}
	if flag.Arg(0) != "create" {
		aliasCommand := ""
		for _, config := range configs {
			file, err := os.Open(configFolder + "/" + config.Name())
			if err != nil {
				Message(response, "error", err.Error(), true)
				return
			}

			var remote Config
			decoder := json.NewDecoder(file)

			if err := decoder.Decode(&remote); err != nil {
				Message(response, "error", err.Error(), true)
				return
			}
			valid := valid(remote)
			found := false
			for _, list := range lists {
				if remote.LocalBindAddress == list {
					found = true
					break
				}
			}

			if !found {
				if len(aliasCommand) > 0 {
					aliasCommand += " && "
				}
				aliasCommand += fmt.Sprintf("ifconfig lo0 alias %s", remote.LocalBindAddress)
				continue
			}

			shellCommand := runCommand(config.Name(), remote)
			status := "Off"
			command := "Start"
			n, err := exec.Command("bash", "-c", fmt.Sprintf("ps aux | grep ssh | grep -v grep | grep %s | wc -l", remote.LocalBindAddress)).CombinedOutput()
			if err == nil {
				number, err := strconv.Atoi(strings.TrimSpace(string(n)))
				if err == nil && number > 0 {
					status = "On"
					command = "Stop"
					shellCommand = fmt.Sprintf("%s#%s#kill -9 $(ps aux | grep ssh | grep -v grep | grep %s | awk '{print $2}')",
						config.Name(), remote.LocalBindAddress, remote.LocalBindAddress)
				}
			}
			item := gofred.NewItem(config.Name(), command+" "+config.Name(), noAutocomplete).AddIcon(status+".png", "").
				AddCommandKeyAction("Modify config", "modify "+config.Name(), true).
				AddOptionKeyAction("Remove config", "remove "+config.Name(), true)
			if valid {
				item = item.Executable(shellCommand)
			}
			items = append(items, item)
		}
		if len(aliasCommand) > 0 {
			items = []gofred.Item{gofred.NewItem("Not Aliased on loopback list", "Run alias command", noAutocomplete).
				AddIcon("icon.png", "").Executable(fmt.Sprintf(`osascript -e "do shell script \"%s\" with administrator privileges"`, aliasCommand))}
		}
		items = append(items, gofred.NewItem("Add new config", noSubtitle, "create ").AddIcon("plus.png", ""))
	} else {
		response.AddVariable("filename", flag.Arg(1))
		items = append(items, gofred.NewItem("Add new config", fmt.Sprintf("write name ... \"%s\"", flag.Arg(1)), noAutocomplete).
			AddIcon("plus.png", "").Executable("newConfig"))
	}
	response.AddItems(items...)
	fmt.Println(response)
}

func valid(conf Config) bool {
	if len(conf.LocalBindAddress) == 0 {
		return false
	}
	if len(conf.RemoteUser) == 0 {
		return false
	}
	if len(conf.RemoteHost) == 0 {
		return false
	}
	return true
}

func runCommand(name string, conf Config) string {
	cmd := fmt.Sprintf("%s#%s#/usr/local/bin/autossh -M 0 -f -q -N", name, conf.LocalBindAddress)
	if len(conf.RemotePort) > 0 {
		cmd += fmt.Sprintf(" -p %s", conf.RemotePort)
	}
	if len(conf.IdentityFile) > 0 {
		cmd += fmt.Sprintf(" -i %s", conf.IdentityFile)
	}
	if conf.ServerAliveInterval > 0 {
		cmd += fmt.Sprintf(" -o ServerAliveInterval=%d", conf.ServerAliveInterval)
	}
	if conf.ServerAliveCountMax > 0 {
		cmd += fmt.Sprintf(" -o ServerAliveCountMax=%d", conf.ServerAliveCountMax)
	}
	if len(conf.StrictHostKeyChecking) > 0 {
		cmd += fmt.Sprintf(" -o StrictHostKeyChecking=%s", conf.StrictHostKeyChecking)
	}
	if len(conf.ForwardPorts) > 0 {
		forwardPorts := ""
		for _, str := range conf.ForwardPorts {
			forwardPorts += fmt.Sprintf(" -L %s%s", conf.LocalBindAddress, str)
		}
		cmd += forwardPorts
	}
	cmd += fmt.Sprintf(" %s@%s", conf.RemoteUser, conf.RemoteHost)

	return cmd
}