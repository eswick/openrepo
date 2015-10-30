package main;

import (
  "os"
  "io"
  "io/ioutil"
  "path"
  "fmt"
  "crypto/md5"
  "strconv"
  "bytes"
  "compress/gzip"
  "net/http"
  "strings"
  "time"
)

var config Config;

func getPackageList() string {
  var retVal string;

  files, err := ioutil.ReadDir(config.PackagePath);
  if (err != nil) { panic(err); }

  for _, file := range files {
    if (!strings.HasSuffix(file.Name(), "deb")) {
      continue;
    }
    filePath := path.Join(config.PackagePath, file.Name());
    fileReader, err := os.Open(filePath);
    if (err != nil) { panic(err); }
    defer fileReader.Close();

    control, err := readDebControlFile(fileReader);
    if (err != nil) { panic(err); }

    // Size
    control += "Size: ";
    control += strconv.FormatInt(file.Size(), 10);
    control += "\n";

    // MD5Sum
    fileReader.Seek(0, 0);
    hash := md5.New();
    _, err = io.Copy(hash, fileReader);
    if (err != nil) { panic(err); }

    sum := fmt.Sprintf("%x", hash.Sum(nil));

    control += "MD5sum: ";
    control += sum;
    control += "\n";

    // Filename
    control += "Filename: ";
    control += strings.TrimPrefix(config.HostPath, "/") + "package/" + file.Name();
    control += "\n";

    retVal += control + "\n";
  }

  return retVal;
}

func getGzippedPackageList() []byte {
  var b bytes.Buffer;

  gz := gzip.NewWriter(&b);
  defer gz.Close();

  _, err := gz.Write([]byte(getPackageList()));
  if (err != nil) { panic(err); }

  err = gz.Close();
  if (err != nil) { panic(err); }

  return b.Bytes();
}

func getRelease() string {
  var release bytes.Buffer;

  for _, entry := range config.Release.Entries {
    release.WriteString(entry.Key + ": " + entry.Value + "\n");
  }

  return release.String();
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
  file := strings.TrimPrefix(path.Clean(req.URL.Path), config.HostPath);

  if (file == "Packages") {
    fmt.Fprintf(w, getPackageList());
  } else if (file == "Packages.gz") {
    pkgList := bytes.NewReader(getGzippedPackageList());
    http.ServeContent(w, req, "application/x-gzip", time.Now(), pkgList);
  } else if (file == "Release") {
    fmt.Fprintf(w, getRelease());
  } else {
    http.NotFound(w, req);
  }
}

func handlePackageRequest(w http.ResponseWriter, req *http.Request) {
  fileName := strings.TrimPrefix(path.Clean(req.URL.Path), config.HostPath + "package/");
  filePath := path.Join(config.PackagePath, fileName);

  http.ServeFile(w, req, filePath);
}

func main() {
  config = getConfig();

  packagePathExists, err := exists(config.PackagePath);
  if (err != nil) { panic(err); }

  if (!packagePathExists) {
    err := os.MkdirAll(config.PackagePath, 755);
    if (err != nil) { panic(err); }
  }

  mux := NewRepoMux();
  mux.HandleFunc(config.HostPath, handleRequest);
  mux.HandleFunc(config.HostPath + "package/", handlePackageRequest);

  err = http.ListenAndServe(":80", mux);

  if (err != nil) { panic(err); }

}
