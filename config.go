package main;

import (
	"os"
	"encoding/xml"
	"io/ioutil"
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

func getConfig() Config {
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
