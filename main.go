package main

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	cfgPath := "workspace.yaml"
	if len(os.Args) > 2 {
		cfgPath = os.Args[1]
	}
	spaces, err := ReadConfig(cfgPath)
	if err != nil {
		log.Fatalln(err)
	}
	for _, space := range spaces {
		ignore := make(map[string]struct{}, len(space.Ignore))
		for _, v := range space.Ignore {
			ignore[v] = struct{}{}
		}
		using := make(map[string]struct{}, len(space.Using))
		for _, v := range space.Using {
			using[v] = struct{}{}
		}

		ents, err := os.ReadDir(space.DstRoot)
		if err != nil {
			log.Printf("Err, read dst root(%s) err: %s\n", space.DstRoot, err.Error())
			return
		}
		exists := make(map[string]struct{}, len(space.Using))
		for _, e := range ents {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if _, ok := ignore[e.Name()]; ok {
				continue
			}
			if _, ok := using[e.Name()]; !ok {
				del := filepath.Join(space.DstRoot, e.Name())
				if err = os.Remove(del); err != nil {
					log.Printf("Err, remove path(%s) err: %s\n", del, err.Error())
					return
				}
			} else {
				exists[e.Name()] = struct{}{}
			}
		}

		for _, name := range space.Using {
			if _, ok := exists[name]; ok {
				continue
			}
			src := filepath.Join(space.SrcRoot, name)

			_, err := os.Stat(src)
			if err != nil {
				log.Printf("Err, Stat (%s) err: %s\n", src, err.Error())
				return
			}

			dst := filepath.Join(space.DstRoot, name)
			err = os.Symlink(src, dst)
			if err != nil {
				log.Printf("Err, Symlink (%s, %s) err: %s\n", src, dst, err.Error())
				return
			}
		}
		log.Printf("%s success", space.Name)
	}
}



type Space struct {
	Name    string   `yaml:"name"`
	Using   []string `yaml:"using"`
	DstRoot string   `yaml:"dst_root"`
	SrcRoot string   `yaml:"src_root"`
	Ignore  []string `yaml:"ignore"`
}

func ReadConfig(path string) ([]*Space, error) {
	c, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	type T struct {
		Spaces []*Space
	}
	v := &T{}
	err = yaml.Unmarshal(c, &v)
	return v.Spaces, err
}
