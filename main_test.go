package main

import (
	"encoding/json"
	"os"
	"testing"

	flags "github.com/jessevdk/go-flags"
	"github.com/nerdalize/nerd/command"
)

//Documented can be implemented by command we want to have documented
type Documented interface {
	Usage() string
	Description() string
	Synopsis() string
	Options() *flags.Parser
}

func TestDocGeneration(t *testing.T) {
	cli := create()

	type opt struct {
		LongName     string   `json:"long_name"`
		ShortName    string   `json:"short_name,omitempty"`
		Description  string   `json:"description"`
		DefaultValue []string `json:"default_value"`
		Choices      []string `json:"choices"`
	}

	type entry struct {
		Usage       string           `json:"usage"`
		Synopsis    string           `json:"synopsis"`
		Description string           `json:"description"`
		Options     map[string][]opt `json:"options"`
	}

	type docs struct {
		GlobalOptions []opt             `json:"global_options"`
		Commands      map[string]*entry `json:"commands"`
	}

	d := docs{
		Commands: map[string]*entry{},
	}

	for name, cmdFn := range cli.Commands {
		if name == "project place" || name == "project expel" || name == "workload work" {
			continue
		}
		cmd, err := cmdFn()
		if err != nil {
			t.Fatalf("failed to create command for documentation purposes: %v", err)
		}

		var (
			ok  bool
			doc Documented
		)

		if doc, ok = cmd.(Documented); !ok {
			t.Logf("command '%s' doesn't implement documented interface, skipping", name)
			continue
		}

		opts := map[string][]opt{}
		d.Commands[name] = &entry{
			Usage:       doc.Usage(),
			Synopsis:    doc.Synopsis(),
			Description: doc.Description(),
			Options:     opts,
		}

		for _, g := range doc.Options().Groups() {

			gopts := []opt{}
			for _, o := range g.Options() {
				op := opt{
					Description:  o.Description,
					LongName:     o.LongName,
					DefaultValue: o.Default,
					Choices:      o.Choices,
				}

				if o.ShortName != rune(0) {
					op.ShortName = string(o.ShortName)
				}

				gopts = append(gopts, op)
			}

			if g.LongDescription != command.SharedOptionsGroup {
				opts[g.LongDescription] = gopts
			} else {
				d.GlobalOptions = gopts
			}
		}

	}

	f, err := os.Create("spec.json")
	if err != nil {
		t.Fatalf("failed to write docs: %v", err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(d)
	if err != nil {
		t.Fatalf("failed to encode: %+v", err)
	}

}
