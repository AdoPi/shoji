package convert // import "github.com/adopi/shoji/convert"

import (
	"log"
	"bufio"
	"strings"
	"os"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"crypto/sha256"
	"encoding/hex"
	"regexp"

	"gopkg.in/yaml.v3"
)

type SshConfig struct {
	Hosts []Host
}

type Host struct {
	Name	string	
	User string 
	Port string 
	Identity string
	Hostname string
	Data string // Contains every other options
}

// Generate SSH config from Yaml
func FromYamlToSSH(yamlFilePath string, outputDirectory string, outputSshConfigFile string) {
	filename, err := filepath.Abs(yamlFilePath)
	if err != nil {
		panic(err)
	}
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var sshConfig SshConfig

	err = yaml.Unmarshal(yamlFile, &sshConfig)
	if err != nil {
		panic(err)
	}

	if err := os.Mkdir(outputDirectory, os.ModePerm); err != nil {
		if os.IsExist(err) {
			// continue
		} else {
			log.Fatal(err)
		}
	}

	// [ssh_key]filename
	// var identities map [string]string

	outputStr := ""

	identities := make(map[string]string)
	for _,v := range sshConfig.Hosts {
		host := v.Name
		var f string

		if v.Identity != "" {
			if filename, found := identities[v.Identity]; found {
				f = filename
			} else {
				// Save key content and associate a filename to it
				hash := sha256.Sum256([]byte(v.Identity))
				id := hex.EncodeToString(hash[:])

				reg := regexp.MustCompile(`[^0-9a-zA-Z\.\-\_]`)
				h := host
				if reg.MatchString(host) {
					h = ""
				} 

				u := v.User
				if u != "" {
					u += "-" 
				}

				keyFile := u + h + "-" + id[:8] + ".key"
				f = filepath.Join(outputDirectory, keyFile)

				identities[v.Identity] = string(f)

				f, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
				if err != nil {
					log.Fatal(err)
				} 
				defer f.Close()

				_, err = f.WriteString(v.Identity)
				if err != nil {
					log.Fatal(err)
				}
			}

		}

		// Generating key content filename
		// Writing key content 

		if host != "" {
			outputStr += "Host " +  host + "\n"
		}
		if v.User != "" {
			outputStr += "\tUser " + v.User + "\n"
		}
		if v.Identity != "" {
			outputStr += "\tIdentityFile " + f + "\n"
		}
		if v.Hostname != "" {
			outputStr += "\tHostname " + v.Hostname + "\n"
		}
		if v.Port != "" {
			outputStr += "\tPort " + v.Port + "\n"
		}
		if v.Data != "" {
			outputStr += strings.Replace("\t" + v.Data, "\n", "\n\t", -1)
		}
	}

	if outputSshConfigFile == "" {
		fmt.Print(outputStr)
		return
	}

	output, err := os.OpenFile(outputSshConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		panic(err)
	} 

	_, err = output.WriteString(outputStr)
	if err != nil {
		panic(err)
	}
}

// Give a relativePath folder if you want to read keys from this folder instead of absolute paths in config file
// for each IdentityFile: read the real path, or the relative path (inside ssh folder)
// then generate yaml
func FromSSHToYaml(sshConfigPath string, relativeFolderPath string, outputFile string) {
	// read ssh config file

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Could not get home directory", err)
		return
	}



	filename, _ := filepath.Abs(sshConfigPath)
	file,err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	// Reading our file line by line
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	// Parsing
	var sshConfig SshConfig

	currentWorkingHostIndex := 0

	for scanner.Scan() {
		line := scanner.Text()
		//line = strings.Trim(line, " ")
		line = strings.TrimSpace(line)
		// check if line is empty
		if len(line) == 0 {
			continue
		}
		// is this a comment?
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Ignore Include keyword
		if strings.HasPrefix(strings.ToLower(line),"include") {
			fmt.Println("Warning: Inlude keyword is not supported. Ignoring...")
			continue
		}

		// Now, I can process my line
		tokens := strings.Split(line," ")

		if len(tokens) < 2 {
			continue
		}

		switch strings.ToLower(tokens[0]) {
		case "host": 
			currentHostKey := strings.Join(tokens[1:], " ")
			// Create Host 
			currentWorkingHost := Host{}
			currentWorkingHost.Name = currentHostKey
			sshConfig.Hosts = append(sshConfig.Hosts,currentWorkingHost)
			currentWorkingHostIndex = len(sshConfig.Hosts) - 1
		case "user":
			sshConfig.Hosts[currentWorkingHostIndex].User = tokens[1]
		case "identityfile":
			identity := ""
			if relativeFolderPath != "" {
				// get filename only and read this file from given folder
				identityFile, err := filepath.Abs(tokens[1])
				if err != nil {
					panic(err)
				}
				// Note: Only works if you dont have subfolders in your given folder

				f := filepath.Join(relativeFolderPath,filepath.Base(identityFile))
				b,err := ioutil.ReadFile(f) 
				identity = string(b)
				if err != nil {
					panic(err)
				}

			} else {
				idFile := tokens[1]
				if idFile[:5] == "$HOME" {
					idFile = filepath.Join(home, idFile[6:])
				}
				if idFile[:2] == "~/" {
					idFile = filepath.Join(home, idFile[2:])
				}

				b,err := ioutil.ReadFile(idFile)

				if err != nil {
					panic(err)
				}
				identity = string(b)
			}
			sshConfig.Hosts[currentWorkingHostIndex].Identity = identity
		case "hostname":
			sshConfig.Hosts[currentWorkingHostIndex].Hostname = tokens[1]
		case "port":
			sshConfig.Hosts[currentWorkingHostIndex].Port = tokens[1]
		default: 
			d := sshConfig.Hosts[currentWorkingHostIndex].Data
			sshConfig.Hosts[currentWorkingHostIndex].Data = d + line + "\n"
		}
	}

	// y,err := yaml.Marshal(sshConfig)

	encoder := yaml.NewEncoder(os.Stdout)
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			panic(err)
		} 
		encoder = yaml.NewEncoder(f)
	}
	// encoder.SetIndent(2)

	err = encoder.Encode(sshConfig)
	if err != nil {
		panic(err)
	} 
}

