package main;

import (
  "errors"
  "io"
  "strings"
  "compress/gzip"
  "archive/tar"
  "bytes"
  "bufio"
  "github.com/blakesmith/ar"
  "regexp"
)

func parseDebControlFile(control string) map[string]string {
  var m map[string]string = make(map[string]string);

  // Find all valid entries
  regex := regexp.MustCompile(`([\w-]+:\s?.*(?:\s .*)*)`);
  res := regex.FindAllString(control, -1);

  for _, str := range res {
    // Parse key: value pairs
    regex = regexp.MustCompile(`([\w-]+):\s?((?:.*\s?.*)*)`);
    result := regex.FindAllStringSubmatch(str, -1);

    for _, str := range result {
      m[str[1]] = str[2];
    }

  }


  return m;
}

func readDebControlFile(reader io.Reader) (string, error) {
  archiveReader := ar.NewReader(reader);

  for {
    header, err := archiveReader.Next();

    if (err == io.EOF) { break; }
    if (err != nil) { panic(err); }

    if (strings.HasPrefix(header.Name, "control.tar")) {
      var controlReader *tar.Reader;

      if (strings.HasSuffix(header.Name, "gz")) {
        gzipStream, err := gzip.NewReader(archiveReader);
        if (err != nil) { panic(err); }
        controlReader = tar.NewReader(gzipStream);
      } else {
        return "", errors.New("Compression type not supported");
      }

      for {
        header, err := controlReader.Next();

        if (err == io.EOF) { break; }
        if (err != nil) { panic(err); }

        if (strings.HasSuffix(header.Name, "control")) {
          var buffer bytes.Buffer;
          _, err := io.Copy(bufio.NewWriter(&buffer), controlReader);
          if (err != nil) { panic(err); }
          return buffer.String(), nil;
        }
      }
    }
  }

  return "", errors.New("Couldn't find control file in package");
}
