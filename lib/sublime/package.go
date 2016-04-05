// Copyright 2016 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package sublime

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/limetext/lime-backend/lib"
	"github.com/limetext/lime-backend/lib/keys"
	"github.com/limetext/lime-backend/lib/log"
	"github.com/limetext/lime-backend/lib/packages"
	_ "github.com/limetext/lime-backend/lib/sublime/api"
	"github.com/limetext/text"
)

// A sublime package
type pkg struct {
	dir  string
	name string
	text.HasSettings
	keys.HasKeyBindings
	platformSettings *text.HasSettings
	defaultSettings  *text.HasSettings
	defaultKB        *keys.HasKeyBindings
	plugins          map[string]*plugin
	// TODO: themes, snippets, etc more info on iss#71
}

func newPKG(dir string) packages.Package {
	return &pkg{
		dir:              dir,
		name:             path.Base(dir),
		platformSettings: new(text.HasSettings),
		defaultSettings:  new(text.HasSettings),
		defaultKB:        new(keys.HasKeyBindings),
		plugins:          make(map[string]*plugin),
	}
}

func (p *pkg) Load() {
	log.Debug("Loading package %s", p.Name())
	p.loadKeyBindings()
	p.loadSettings()
	p.loadPlugins()
}

func (p *pkg) Path() string {
	return p.dir
}

func (p *pkg) Name() string {
	return p.name
}

// TODO: how we should watch the package and the files containing?
func (p *pkg) FileCreated(name string) {}

func (p *pkg) loadPlugins() {
	log.Fine("Loading %s plugins", p.Name())
	fis, err := ioutil.ReadDir(p.Path())
	if err != nil {
		log.Warn("Error on reading directory %s, %s", p.Path(), err)
		return
	}
	for _, fi := range fis {
		if isPlugin(fi.Name()) {
			p.loadPlugin(path.Join(p.Path(), fi.Name()))
		}
	}
}

func (p *pkg) loadPlugin(fn string) {
	if _, exist := p.plugins[fn]; exist {
		return
	}

	pl := newPlugin(fn)
	pl.Load()

	p.plugins[fn] = pl.(*plugin)
}

func (p *pkg) loadKeyBindings() {
	log.Fine("Loading %s keybindings", p.Name())
	ed := backend.GetEditor()
	tmp := ed.KeyBindings().Parent()

	ed.KeyBindings().SetParent(p)
	p.KeyBindings().SetParent(p.defaultKB)
	p.defaultKB.KeyBindings().SetParent(tmp)

	pt := path.Join(p.Path(), "Default.sublime-keymap")
	log.Finest("Loading %s", pt)
	packages.LoadJSON(pt, p.defaultKB.KeyBindings())

	pt = path.Join(p.Path(), "Default ("+ed.Plat()+").sublime-keymap")
	log.Finest("Loading %s", pt)
	packages.LoadJSON(pt, p.KeyBindings())
}

func (p *pkg) loadSettings() {
	log.Fine("Loading %s settings", p.Name())
	ed := backend.GetEditor()
	tmp := ed.Settings().Parent()

	ed.Settings().SetParent(p)
	p.Settings().SetParent(p.platformSettings)
	p.platformSettings.Settings().SetParent(p.defaultSettings)
	p.defaultSettings.Settings().SetParent(tmp)

	pt := path.Join(p.Path(), "Preferences.sublime-settings")
	log.Finest("Loading %s", pt)
	packages.LoadJSON(pt, p.defaultSettings.Settings())

	pt = path.Join(p.Path(), "Preferences ("+ed.Plat()+").sublime-settings")
	log.Finest("Loading %s", pt)
	packages.LoadJSON(pt, p.platformSettings.Settings())

	pt = path.Join(ed.PackagesPath("user"), "Preferences.sublime-settings")
	log.Finest("Loading %s", pt)
	packages.LoadJSON(pt, p.Settings())
}

// Any directory in sublime is a package
func isPKG(dir string) bool {
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return false
	}
	return true
}

var packageRecord *packages.Record = &packages.Record{isPKG, newPKG}

func init() {
	packages.Register(packageRecord)
}