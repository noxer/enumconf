package enumconf

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Options contains the enumeration options for config files
type Options struct {
	appName          string
	configName       string
	configNameInPath string
	includeMissing   bool
	fs               fs.FS
}

// Create a new config file enumerator for the provided app name. The app name must be a valid path segment.
func New(appName string) *Options {
	return &Options{
		appName:          appName,
		configName:       appName + ".conf",
		configNameInPath: "." + appName,
		fs:               os.DirFS(root),
	}
}

// ConfigName overrides the default config file name of appName.conf
func (o *Options) ConfigName(configName string) *Options {
	o.configName = configName
	return o
}

// ConfigNameInPath overrides the default in-path config file name of .appName
func (o *Options) ConfigNameInPath(configName string) *Options {
	o.configNameInPath = configName
	return o
}

// IncludeMissing defines if missing files should still be present in the config file list
func (o *Options) IncludeMissing(include bool) *Options {
	o.includeMissing = include
	return o
}

// FS overrides the default file system
func (o *Options) FS(fs fs.FS) *Options {
	o.fs = fs
	return o
}

// Enumerate generates a list of config files present in the system. It is ordered system config
// files, user config files, path config from root to the current working directory.
func (o *Options) Enumerate() []string {
	var files []string
	files = o.appendSystem(files)
	files = o.appendUser(files)
	files = o.appendPath(files)
	return files
}

// EnumerateSystem generates a list of system config files. May be nil or empty.
func (o *Options) EnumerateSystem() []string {
	return o.appendSystem(nil)
}

func (o *Options) appendSystem(files []string) []string {
	for _, s := range systemConfigs {
		p := filepath.Join(s, o.appName, o.configName)
		files = o.appendIfFile(files, p)
	}

	return files
}

// EnumerateUser generates a list of user config files. May be nil or empty.
func (o *Options) EnumerateUser() []string {
	return o.appendUser(nil)
}

func (o *Options) appendUser(files []string) []string {
	d, err := os.UserConfigDir()
	if err != nil {
		return files
	}

	p := filepath.Join(d, o.appName, o.configName)
	return o.appendIfFile(files, p)
}

// EnumeratePath generates a list of config files in the current working dir or it's parents. May be nil or empty.
func (o *Options) EnumeratePath() []string {
	return o.appendPath(nil)
}

func (o *Options) appendPath(files []string) []string {
	path, err := os.Getwd()
	if err != nil {
		return files
	}

	return o.appendPathFromRoot(files, path)
}

func (o *Options) appendPathFromRoot(files []string, path string) []string {
	parent := filepath.Dir(path)
	if parent != path {
		files = o.appendPathFromRoot(files, parent)
	}

	p := filepath.Join(path, o.configNameInPath)
	return o.appendIfFile(files, p)
}

func (o *Options) appendIfFile(files []string, name string) []string {
	if !o.includeMissing {
		info, err := fs.Stat(o.fs, name)
		if err != nil || info.IsDir() {
			return files
		}
	}

	return append(files, name)
}
