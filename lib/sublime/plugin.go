// Copyright 2016 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package sublime

import (
	"path/filepath"

	"github.com/limetext/gopy/lib"
	"github.com/limetext/lime-backend/lib"
	"github.com/limetext/lime-backend/lib/log"
	"github.com/limetext/lime-backend/lib/packages"
	_ "github.com/limetext/lime-backend/lib/sublime/api"
)

// Sublime plugin which is a single python file
type plugin struct {
	path string
	name string
}

func newPlugin(fn string) packages.Package {
	return &plugin{path: fn}
}

// TODO: implement unload
func (p *plugin) Load() {
	// in case error ocured on running onInit function
	// TODO: it's better to unregister plugin record but then we need to
	// unregister package record to
	if module == nil {
		return
	}

	dir, file := filepath.Split(p.Path())
	p.name = file
	name := filepath.Base(dir) + "." + file[:len(file)-3]
	s, err := py.NewUnicode(name)
	if err != nil {
		log.Warn(err)
		return
	}
	defer s.Decref()

	log.Debug("Loading plugin %s", name)
	l := py.NewLock()
	defer l.Unlock()
	if r, err := module.Base().CallMethodObjArgs("reload_plugin", s); err != nil {
		log.Warn(err)
		return
	} else if r != nil {
		r.Decref()
	}
}

func (p *plugin) Name() string {
	return p.name
}

func (p *plugin) Path() string {
	return p.path
}

func (p *plugin) FileChanged(name string) {
	p.Load()
}

func isPlugin(filename string) bool {
	return filepath.Ext(filename) == ".py"
}

var (
	pluginRecord *packages.Record = &packages.Record{isPlugin, newPlugin}

	module *py.Module
)

func onInit() {
	l := py.NewLock()
	defer l.Unlock()
	var err error
	if module, err = py.Import("sublime_plugin"); err != nil {
		log.Error("Error importing sublime_plugin: %s", err)
	}
}

func onPackagesPathAdd(p string) {
	l := py.NewLock()
	defer l.Unlock()
	py.AddToPath(p)
}

func init() {
	backend.OnInit.Add(onInit)
	backend.OnPackagesPathAdd.Add(onPackagesPathAdd)
	packages.Register(pluginRecord)
}