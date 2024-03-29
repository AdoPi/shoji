package main

import (
	"os"
	shoji "github.com/adopi/shoji/convert"
	"github.com/alecthomas/kong"
)

type CLI struct {
	Convert Convert `cmd:"" help:"Convert files."`
}

type Convert struct {
	YamlCommand YamlCommand `cmd:"" name:"yaml" help:"Convert back a yaml file into ssh config and a folder containing all private keys."`
	SshCommand SshCommand `cmd:"" name:"ssh" help:" Convert ssh config file and ssh folder containing private keys into a single YAML file."`
}

// Yaml into SSH
type YamlCommand struct {
	Input   string `arg:"" help:"Input file, needs to be a Yaml file, previously generated with this program." type:"path"`
	K      string `short:"k" default:"./ssh" optional:"" name:"keys-directory" help:"SSH directory which will be containing all private keys. Default is ./ssh" type:"path"`
	Output string `short:"o" optional:"" name:"output" help:"SSH config output file." type:"path"`
	Unsecure bool `short:"u" name:"unsecure" help:"Unsecurely prints to stdout (useful for pipes)."`
}

// SSH into yaml
type SshCommand struct {
	Input   string `arg:"" help:"Input file, needs to be a SSH Config file" type:"path"`
	K      string `short:"k" optional:"" name:"keys-directory" help:"Directory where to read SSH private keys. Read paths from ssh config file by default." type:"path"`
	Output string `short:"o" optional:"" name:"output" help:"Yaml output file. Default is stdout (NOT SAFE)." type:"path"`
	Unsecure bool `short:"u" name:"unsecure" help:"Unsecurely prints to stdout (useful for pipes)."`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
	kong.Name("shoji"),
	kong.Description("Shoji, parse SSH config file, read SSH keys and convert everything into a single YAML file."),
	kong.UsageOnError(),
)
err := run(ctx, &cli)
ctx.FatalIfErrorf(err)
}

func run(ctx *kong.Context, cli *CLI) error {
	switch {
	case ctx.Command() == "convert yaml <input>":
		output := ""
		if cli.Convert.YamlCommand.Unsecure == false {
			output = cli.Convert.YamlCommand.Output
			if output == "" {
				os.Stderr.WriteString("Use --unsecure (-u) to unsafely print secrets to stdout or --output (-o) file.")
				os.Exit(1)
			}
		} else {
			if cli.Convert.YamlCommand.Output != "" {
				os.Stderr.WriteString("You can't mix --unsecure (-u) and --output (-o) options.")
				os.Exit(2)
			}
		}
		shoji.FromYamlToSSH(cli.Convert.YamlCommand.Input, cli.Convert.YamlCommand.K, output)
	case ctx.Command() == "convert ssh <input>":
		output := ""
		if cli.Convert.SshCommand.Unsecure == false {
			output = cli.Convert.SshCommand.Output
			if output == "" {
				os.Stderr.WriteString("Use --unsecure (-u) to unsafely print secrets to stdout or --output (-o) file")
				os.Exit(1)
			}
		} else {
			if cli.Convert.SshCommand.Output != "" {
				os.Stderr.WriteString("You can't mix --unsecure (-u) and --output (-o) options.")
				os.Exit(2)
			}
		}
		shoji.FromSSHToYaml(cli.Convert.SshCommand.Input,cli.Convert.SshCommand.K, output)
	}
	return nil
}
