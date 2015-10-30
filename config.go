package main;

import (
	"os"
	"encoding/xml"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
)

type ReleaseEntry struct {
	Key string `xml:"key,attr"`;
	Value string `xml:",chardata"`;
}

type Release struct {
	Entries []ReleaseEntry `xml:"ReleaseEntry"`;
}

type Config struct {
  HostPath string;
  PackagePath string;
	Release Release;
}

func createConfig() {

	defaultConfigPath, err := filepath.Abs(filepath.Dir(os.Args[0]));

	if (err != nil) { panic(err); }

	in, err := os.Open(path.Join(defaultConfigPath, "config_default.xml"));
	if (err != nil) { panic(err); }
	defer in.Close();

	out, err := os.Create("/etc/openrepo/config.xml");
	defer out.Close();

	_, err = io.Copy(out, in);
	cerr := out.Close();

	if (err != nil) { panic(err); }
	if (cerr != nil) { panic(err); }
}

func getConfig() Config {
	configDirExists, err := exists("/etc/openrepo");
	if (err != nil) { panic(err); }

	if (!configDirExists) {
		os.Mkdir("/etc/openrepo", 777);
	}

	configExists, err := exists("/etc/openrepo/config.xml");
	if (err != nil) { panic(err); }

	if (!configExists) {
		createConfig();
	}

	reader, err := os.Open("/etc/openrepo/config.xml");
	if (err != nil) { panic(err); }
	defer reader.Close();

	var config Config;

	buffer, err := ioutil.ReadAll(reader);
	if (err != nil) { panic(err); }

	err = xml.Unmarshal(buffer, &config);
	if (err != nil) { panic(err); }

	return config;
}
