package main

import (
	"fmt"
	"os"
	"reflect"

	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/go-etcd/etcd"
	flags "github.com/jessevdk/go-flags"
	"github.com/mickep76/iodatafmt"
)

// Import data.
func Import(client *etcd.Client, dir string, d map[string]interface{}) error {
	for k, v := range d {
		if reflect.ValueOf(v).Kind() == reflect.Map {
			fmt.Printf("%s/%s/\n", dir, k)
			if _, err := client.CreateDir(dir+"/"+k, 0); err != nil {
				log.Fatal(err.Error())
			}
			Import(client, dir+"/"+k, v.(map[string]interface{}))
		} else {
			fmt.Printf("%s/%s: %s\n", dir, k, v)
			if _, err := client.Set(dir+"/"+k, v.(string), 0); err != nil {
				log.Fatal(err.Error())
			}
		}
	}

	return nil
}

func main() {
	// Set log options.
	log.SetOutput(os.Stderr)
	log.SetLevel(log.WarnLevel)

	// Options.
	var opts struct {
		Verbose  bool    `short:"v" long:"verbose" description:"Verbose"`
		Version  bool    `long:"version" description:"Version"`
		Format   string  `short:"f" long:"format" description:"Data serialization format YAML, TOML or JSON" default:"JSON"`
		Input    *string `short:"i" long:"input" description:"Input file (STDOUT)"`
		EtcdNode *string `short:"n" long:"etcd-node" description:"Etcd Node"`
		EtcdPort int     `short:"p" long:"etcd-port" description:"Etcd Port" default:"2379"`
		EtcdDir  string  `short:"d" long:"etcd-dir" description:"Etcd Dir" default:"/"`
	}

	// Parse options.
	if _, err := flags.Parse(&opts); err != nil {
		ferr := err.(*flags.Error)
		if ferr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Fatal(err.Error())
		}
	}

	// Print version.
	if opts.Version {
		fmt.Printf("etcd-export %s\n", Version)
		os.Exit(0)
	}

	// Set verbose.
	if opts.Verbose {
		log.SetLevel(log.InfoLevel)
	}

	// Validate input.
	if opts.EtcdNode == nil {
		log.Fatalf("You need to specify Etcd host.")
	}

	// Get data format.
	f, err := iodatafmt.Format(opts.Format)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Setup Etcd client.
	node := []string{fmt.Sprintf("http://%v:%v", *opts.EtcdNode, opts.EtcdPort)}
	client := etcd.NewClient(node)

	// Import data.
	d, err := iodatafmt.Load(*opts.Input, f)
	if err != nil {
		log.Fatal(err.Error())
	}

	Import(client, "", d)
}