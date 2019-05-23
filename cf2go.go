package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app  = kingpin.New("cf2go", "A command line toolset to manage CF landscapes")
	list = app.Command("list", "List landscapes configuration")

	jump            = app.Command("jump", "SSH to a jumpbox system")
	jumpLandscapeID = jump.Arg("landscape-id", "Call 'cf2go list' to get available landscape ids").Required().String()

	tunnel            = app.Command("tunnel", "SSH tunnel via jumpbox to target")
	tunnelLandscapeID = tunnel.Arg("landscape-id", "Call 'cf2go list' to get available landscape ids").Required().String()
	tunnelTarget      = tunnel.Arg("target", "Supported targets: [director|concourse], director is default").String()

	details            = app.Command("details", "Display landscape details")
	detailsLandscapeID = details.Arg("landscape-id", "Call 'cf2go list' to get available landscape ids").Required().String()

	login            = app.Command("login", "login to cf api endpoint")
	loginLandscapeID = login.Arg("landscape-id", "Call 'cf2go list' to get available landscape ids").Required().String()

	url = app.Flag("url", "URL from which the landscape configuration is requested").Default("https://raw.githubusercontent.com/sklevenz/cf2go/master/landscape.json").String()
)

type detailsCommandStruct struct{}

type listCommandStruct struct{}

type loginCommandStruct struct{}

type jumpCommandStruct struct{}

type tunnelCommandStruct struct{}

type landscapeStruct struct {
	LandscapeType string `json:"type"`
	Description   string `json:"description"`
	Owner         string `json:"owner"`
	JumpboxIP     string `json:"jumpbox"`
	ConcourseIP   string `json:"concourse"`
	DirectorIP    string `json:"director"`
	Domain        string `json:"domain"`
}

func (n *tunnelCommandStruct) run(c *kingpin.ParseContext) error {
	config := parseConfiguration(readConfiguration())

	landscape, ok := (*config)[*tunnelLandscapeID]
	if !ok {
		log.Printf("Unknown landscape id: %v", *tunnelLandscapeID)
		os.Exit(1)
	}

	s := *tunnelTarget
	var param string
	switch {
	case s == "" || s == "director":
		param = fmt.Sprintf("localhost:25555:%v:25555", landscape.DirectorIP)
	case s == "concourse":
		param = fmt.Sprintf("localhost:8080:%v:8080", landscape.ConcourseIP)
	default:
		log.Printf("Unknown tunnel target: %v", *tunnelTarget)
		os.Exit(1)
	}

	log.Println("Tunnel open: " + param)
	cmdStr := "ssh"
	cmd := exec.Command(cmdStr, "-nNTL", param, landscape.JumpboxIP)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Printf("Execute: %v %v", cmdStr, landscape.JumpboxIP)
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
	return nil
}

func (n *loginCommandStruct) run(c *kingpin.ParseContext) error {
	config := parseConfiguration(readConfiguration())

	landscape, ok := (*config)[*loginLandscapeID]
	if !ok {
		log.Printf("Unknown landscape id: %v", *loginLandscapeID)
		os.Exit(1)
	}

	cmdStr := "cf"
	cmd := exec.Command(cmdStr, "api", fmt.Sprintf("https://api.cf.%s", landscape.Domain), "--skip-ssl-validation")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Printf("Execute: %v %v", cmdStr, landscape.JumpboxIP)
		log.Printf("Error: %v", err)
		os.Exit(1)
	}

	return nil
}

func (n *jumpCommandStruct) run(c *kingpin.ParseContext) error {
	config := parseConfiguration(readConfiguration())

	landscape, ok := (*config)[*jumpLandscapeID]
	if !ok {
		log.Printf("Unknown landscape id: %v", *jumpLandscapeID)
		os.Exit(1)
	}

	cmdStr := "ssh"
	cmd := exec.Command(cmdStr, landscape.JumpboxIP)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Printf("Execute: %v %v", cmdStr, landscape.JumpboxIP)
		log.Printf("Error: %v", err)
		os.Exit(1)
	}

	return nil
}

func (n *detailsCommandStruct) run(c *kingpin.ParseContext) error {
	config := parseConfigurationRaw(readConfiguration())

	landscape, ok := (*config)[*detailsLandscapeID]
	if !ok {
		log.Printf("Unknown landscape id: %v", *detailsLandscapeID)
		os.Exit(1)
	}

	var keys []string
	landscapeMap := landscape.(map[string]interface{})
	for k := range landscapeMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("details called: " + *detailsLandscapeID)
	fmt.Printf("landscape - %v: \n", landscape)
	fmt.Printf("value: %v \n", landscapeMap["domain"])

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"property", "value"})

	for _, k := range keys {
		v := (landscapeMap)[k]
		table.Append([]string{k, v.(string)})
	}
	table.Render() // Send output

	return nil
}

func (n *listCommandStruct) run(c *kingpin.ParseContext) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"landscape-id", "type", "description", "owner", "jumpbox ip", "concourse ip", "director ip", "domain"})

	config := parseConfiguration(readConfiguration())
	var keys []string
	for k := range *config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := (*config)[k]
		table.Append([]string{k, v.LandscapeType, v.Description, v.Owner, v.JumpboxIP, v.ConcourseIP, v.DirectorIP, v.Domain})
	}
	table.Render() // Send output

	return nil
}

func init() {
}

func parseConfiguration(jsonValue string) *map[string]landscapeStruct {
	var cfg map[string]landscapeStruct
	err := json.Unmarshal([]byte(jsonValue), &cfg)
	if err != nil {
		log.Printf("json = '%s'", jsonValue)
		log.Println("error:", err)
		os.Exit(1)
	}

	return &cfg
}

func parseConfigurationRaw(jsonValue string) *map[string]interface{} {
	var cfg map[string]interface{}
	err := json.Unmarshal([]byte(jsonValue), &cfg)
	if err != nil {
		log.Printf("json = '%s'", jsonValue)
		log.Println("error:", err)
		os.Exit(1)
	}

	return &cfg
}

func readConfiguration() string {
	response, err := http.Get(*url)

	if err != nil {
		log.Printf("Error: %v for url: %s", err, *url)
		os.Exit(1)
	}

	if response.StatusCode != 200 {
		log.Printf("Http status (%v): '%v' for requesting url: %s", response.StatusCode, response.Status, *url)
		os.Exit(1)
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error while reading content: %s", err)
		os.Exit(1)
	}
	return string(content)
}

func main() {
	app.Version("0.1." + REVISION).Author("Stephan Klevenz")
	app.HelpFlag.Short('h')

	listCommand := &listCommandStruct{}
	list.Action(listCommand.run)
	jumpCommand := &jumpCommandStruct{}
	jump.Action(jumpCommand.run)
	loginCommand := &loginCommandStruct{}
	login.Action(loginCommand.run)
	tunnelCommand := &tunnelCommandStruct{}
	tunnel.Action(tunnelCommand.run)
	detailsCommand := &detailsCommandStruct{}
	details.Action(detailsCommand.run)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
