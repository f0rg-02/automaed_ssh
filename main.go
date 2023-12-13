package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

/*
Written by: f0rg/Alex
Mastodon is: https://infosec.exchange/@alex_02
Buymeacoffee: https://www.buymeacoffee.com/alex_f0rg

A lot of this code I either borrowed from SO and the libraries or I
reused from other programs that I've written. The whole SO and library
codes is because I would've written the exact same thing and couldn't be
bothered to write the same code by hand. Feel free to edit it as is and reuse
the code. I don't really care too much about getting credit since we are programmers
and we copy each other's codes. I do wish that you don't explicity take full credit
as being your own and if you straight up copy and paste some of this code for lets say
your programming homework, that is on you and teachers always know when someone
is cheating and copying code.
*/

type Config struct {
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
	User     string `yaml:"username"`
	Slurp    int    `yaml:"sleep"` // Don't ask. I am sleep deprived.
	Key_Path string `yaml:"key_path"`
	Shell    string `yaml:"shell"`

	Commands []string `yaml:"commands"`

	APT_Update   bool     `yaml:"apt_update"`
	APT_Upgrade  bool     `yaml:"apt_upgrade"`
	APT_Packages []string `yaml:"apt_packages"`

	Upload bool `yaml:"upload_files"`

	// Thanks. I hate it.
	Files struct {
		File []struct {
			Source     string `yaml:"source"`
			Destinatin string `yaml:"destination"`
		} `yaml:"file"`
	} `yaml:"files"`
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: \n", os.Args[0])
	fmt.Println()
	flag.PrintDefaults()
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	var password string // To store the password variable to be used with sudo if triggered.
	var file *string

	file = flag.String("f", "", "Specify config yaml file")

	flag.Parse()

	if !isFlagPassed("f") {
		usage()
		os.Exit(1)
	}

	config := ChkYaml(file)

	// Start new ssh connection with private key.
	auth, err := goph.Key(config.Key_Path, "")
	if err != nil {
		log.Fatal(err)
	}

	client, err := goph.New(config.User, config.Server, auth)
	if err != nil {
		log.Fatal(err)
	}

	// Defer closing the network connection.
	defer client.Close()

	// Check if upload was specified
	if config.Upload {

		for _, file := range config.Files.File {
			// Upload files
			fmt.Printf("Uploading file: %s to destination %s on server %s\n", file.Source, file.Destinatin, config.Server)
			err := client.Upload(file.Source, file.Destinatin) // goph does have upload and download in the library

			if err != nil { // error was returned
				log.Fatal(err) // Print error and exit
			} else {
				fmt.Println("Done uploading")

			}
		}

	} else {
		fmt.Println("User specified upload to false")
	}

	// I usually do something like sudo apt update && sudo apt upgrade --assume-yes
	// But for this, it should be specified seperately in the yaml
	if config.APT_Update { // Check if true or false in config.yaml.
		// If true, apt update
		// Get password
		fmt.Println("Updating")
		if password == "" { // Checks if the variable is empty or not.
			password, err = credentials()
			if err != nil {
				log.Fatal(err)
			}

		}

		// Execute your command.
		// Library doesn't support sudo properly so had to do some stupid string foo.
		// Looks sloppy to my eyes as of this writing, but it works well and to help
		// give a layer of security, the functin credentials() uses term which doesn't echo.
		// Sudo -S helps prevent from the command including the echo from showing up in history.
		_, err := client.Run("echo " + password + "| sudo -S apt update")

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Updating compeleted")
	} else {
		fmt.Println("User specified update to false")
	}

	if config.APT_Upgrade { // Check if true or false in config.yaml.
		// If true, apt upgrade
		fmt.Println("Upgrading")
		if password == "" { // Checks if the variable is empty or not.
			password, err = credentials()
			if err != nil {
				log.Fatal(err)
			}

		}

		// Execute your command.
		// Library doesn't support sudo properly so had to do some stupid string foo.
		// Looks sloppy to my eyes as of this writing, but it works well and to help
		// give a layer of security, the functin credentials() uses term which doesn't echo.
		// Sudo -S helps prevent from the command including the echo from showing up in history.
		_, err := client.Run("echo " + password + "| sudo -S apt upgrade --assume-yes")

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Upgrading completed")
	} else {
		fmt.Println("User specified upgrade to false")
	}

	if len(config.APT_Packages) != 0 { // Probably not needed.
		fmt.Println("Installing packages")
		if password == "" { // Checks if the variable is empty or not.
			password, err = credentials()
			if err != nil {
				log.Fatal(err)
			}

		}

		for i := range config.APT_Packages {

			install := config.APT_Packages[i]
			fmt.Println("\nInstalling: ", install)

			// Execute your command.
			// Library doesn't support sudo properly so had to do some stupid string foo.
			// Looks sloppy to my eyes as of this writing, but it works well and to help
			// give a layer of security, the functin credentials() uses term which doesn't echo.
			// Sudo -S helps prevent from the command including the echo from showing up in history.
			_, err := client.Run("echo " + password + "| sudo -S apt install --assume-yes " + install)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Done installing: ", install)
		}
	} else {
		fmt.Println("No packages were set to install.")
	}

	if len(config.Commands) != 0 { // Probably not needed.
		for i := range config.Commands {
			command := config.Commands[i]

			// For the sake of readability adding a bunch of dashes makes it look nicer to my eyes.
			// Please do not request ASCII art unless it is actually good and not obnoxiously large.
			if i == 0 {
				fmt.Printf("\n-------------------------------------\n")
			}
			fmt.Printf("Executing command: %s\n", command)
			// Execute your command.
			out, err := client.Run(command)

			if err != nil {
				log.Fatal(err)
			}

			// Get your output as []byte.
			fmt.Printf("-------------------------------------\n")
			fmt.Printf("Output of command: \n-------------------------------------\n%s", string(out))
			fmt.Printf("\n-------------------------------------\n")
			if config.Slurp != 0 {
				time.Sleep(time.Duration(config.Slurp) * time.Second)
			}
		}
	} else {
		fmt.Println("No commands were specified")
	}
}

// Taken from here: https://stackoverflow.com/a/32768479

func credentials() (string, error) {
	fmt.Print("Enter Password: ")                              // Ask for password
	bytePassword, err := term.ReadPassword(int(syscall.Stdin)) // Use term to read file. Check for any errors.
	if err != nil {
		return "", err
	}

	password := string(bytePassword) // Convert the bytes to a string so we can return the password.
	return strings.TrimSpace(password + "\n"), nil
}

func ChkYaml(file *string) Config {
	var config Config
	_, err := os.Stat(*file) // check if config exists
	if err == nil {          // If exists, read file
		data, err := os.ReadFile(*file)
		if err != nil { // Check for any errors reading the file
			log.Fatal(err)
		}

		if err := yaml.Unmarshal(data, &config); err != nil { // Try to unmarshal. Check for any errors.
			log.Fatal(err)
		}
	}
	return config

}
