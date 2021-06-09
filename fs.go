package glogp

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func gzipLogFile(filename string) error {
	rawfile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer rawfile.Close()

	info, _ := rawfile.Stat()
	var size int64 = info.Size()
	rawbytes := make([]byte, size)

	buffer := bufio.NewReader(rawfile)
	_, err = buffer.Read(rawbytes)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	writer.Write(rawbytes)
	writer.Close()

	err = ioutil.WriteFile(filename+".gz", buf.Bytes(), info.Mode())
	if err != nil {
		return err
	}
	return nil
}

func mkdirForce(dir string) error {
	var err error
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func rotateLogArchives(fpath string) error {
	dir := filepath.Dir(fpath)
	srcFilename := filepath.Base(fpath)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for i := len(files) - 1; i >= 0; i-- {
		spFilename := strings.Split(files[i].Name(), srcFilename)
		if len(spFilename) != 2 {
			continue
		}
		spFilename[1] = strings.TrimPrefix(spFilename[1], ".")
		spFilename[1] = strings.TrimSuffix(spFilename[1], ".gz")
		fileNum, err := strconv.Atoi(spFilename[1])
		if err != nil {
			continue
		}
		curArchFpath := fmt.Sprintf("%s/%s.%d.gz", dir, srcFilename, fileNum)
		if fileNum >= LogArchivesMax {
			if err = os.Remove(curArchFpath); err != nil {
				return err
			}
		} else {
			newArchFpath := fmt.Sprintf("%s/%s.%d.gz", dir, srcFilename, fileNum+1)
			if err = os.Rename(curArchFpath, newArchFpath); err != nil {
				return err
			}
		}
	}

	return nil
}
