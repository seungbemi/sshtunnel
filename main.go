package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/seungbemi/gofred"
	"gopkg.in/yaml.v2"
)

const (
	noSubtitle     = ""
	noArg          = ""
	noAutocomplete = ""
)

type host struct {
	RemoteUser   string   `yaml:"RemoteUser"`
	RemoteHost   string   `yaml:"RemoteHost"`
	RemotePort   string   `yaml:"RemotePort"`
	ForwardPorts []string `yaml:"ForwardPorts"`
}

// Config includes
type Config struct {
	host                  `yaml:",inline"`
	ServerAliveInterval   int    `yaml:"ServerAliveInterval"`
	ServerAliveCountMax   int    `yaml:"ServerAliveCountMax"`
	StrictHostKeyChecking string `yaml:"StrictHostKeyChecking"`
	IdentityFile          string `yaml:"IdentityFile"`
	ProxyCommand          string `yaml:"ProxyCommand"`
	LocalBindAddress      string `yaml:"LocalBindAddress"`
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
	fmt.Println(response)
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
	configPath := os.Getenv("alfred_workflow_data") + "/" + configFolder
	response := gofred.NewResponse()

	err := os.MkdirAll(configPath, os.ModePerm)
	if err != nil {
		Message(response, "error", err.Error(), true)
		return
	}

	configs, err := ioutil.ReadDir(configPath)
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
			var remote Config
			bt, err := ioutil.ReadFile(configPath + "/" + config.Name())
			if err != nil {
				Message(response, "error", err.Error(), true)
				return
			}

			err = yaml.Unmarshal(bt, &remote)
			if err != nil {
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

			name := strings.TrimSuffix(config.Name(), ".yml")
			shellCommand := runCommand(remote)
			rebootCommand := shellCommand
			status := "Off"
			command := "Start"
			n, err := exec.Command("bash", "-c", fmt.Sprintf("ps aux | grep ssh | grep -v grep | grep %s | wc -l", remote.LocalBindAddress)).CombinedOutput()
			if err == nil {
				number, err := strconv.Atoi(strings.TrimSpace(string(n)))
				if err == nil && number > 0 {
					status = "On"
					command = "Stop"
					shellCommand = fmt.Sprintf("kill -9 $(ps aux | grep ssh | grep -v grep | grep %s | awk '{print $2}')", remote.LocalBindAddress)
					rebootCommand = "(" + shellCommand + ") && " + rebootCommand
				}
			}
			item := gofred.NewItem(name, command+" "+name, noAutocomplete).AddIcon(status+".png", "").
				AddVariables(gofred.NewVariable("name", name), gofred.NewVariable("cmd", command), gofred.NewVariable("remote", remote.LocalBindAddress)).
				AddOptionKeyAction("Modify config", "modify", true).AddOptionKeyVariables(gofred.NewVariable("name", name), gofred.NewVariable("cmd", "modify")).
				AddCtrlKeyAction("Remove config", "remove", true).AddCtrlKeyVariables(gofred.NewVariable("name", name), gofred.NewVariable("cmd", "remove"))
			if valid {
				item = item.Executable(shellCommand)
				if status == "On" {
					item = item.AddCommandKeyAction("Reboot "+name, rebootCommand, true).
						AddCommandKeyVariables(gofred.NewVariable("name", name), gofred.NewVariable("cmd", "reboot"), gofred.NewVariable("remote", remote.LocalBindAddress))
				}
			}

			items = append(items, item)
		}
		if len(aliasCommand) > 0 {
			items = []gofred.Item{gofred.NewItem("Not Aliased on loopback list", "Run alias command", noAutocomplete).
				AddIcon("icon.png", "").AddVariables(gofred.NewVariable("cmd", "alias")).Executable(fmt.Sprintf(`osascript -e "do shell script \"%s\" with administrator privileges"`, aliasCommand))}
		}
		items = append(items, gofred.NewItem("Add new config", noSubtitle, "create ").AddIcon("plus.png", ""))
	} else {
		items = append(items, gofred.NewItem("Add new config", fmt.Sprintf("write name ... \"%s\"", flag.Arg(1)), noAutocomplete).
			AddIcon("plus.png", "").AddVariables(gofred.NewVariable("filename", flag.Arg(1)), gofred.NewVariable("cmd", "new")).Executable("new"))
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

func runCommand(conf Config) string {
	cmd := fmt.Sprintf("/usr/local/bin/autossh -M 0 -f -q -N")
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

	if len(conf.ProxyCommand) > 0 {
		cmd += fmt.Sprintf(" -o ProxyCommand=\"%s\"", conf.ProxyCommand)
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
