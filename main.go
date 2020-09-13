package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	ServiceAccountJSONPath string `long:"service-account" required:"true" description:"Path to Service Account Credential (JSON Format)."`
	Name                   string `long:"name" required:"true" description:"Name of Upload File"`
	ParentFolderID         string `long:"parent" required:"true" description:"Parent Folder ID in Google Drive."`
	ChunkSize              int    `long:"chunk-size" default:"8388608" description:"Chunk Size of Upload. This must be divisible by 256*1024 bytes (Upstream Constraint)."`
}

var opts Options

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.ChunkSize == 0 {
		log.Panicln("--chunk-size must not be zero")
	}
	if (opts.ChunkSize % 256 * 1024) != 0 {
		log.Panicln("--chunk-size must be divisible by 256*1024")
	}

	client := getClientFromJWTConfig(opts.ServiceAccountJSONPath)

	// -- Create Resumable Upload Session
	jsonBytes, err := json.Marshal(map[string]interface{}{
		"name":    opts.Name,
		"parents": []string{opts.ParentFolderID},
	})
	if err != nil {
		panic(err)
	}
	session, err := client.Post("https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable&supportsAllDrives=true&fields=md5Checksum,id,name,mimeType,driveId,size", "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}
	// log.Println(session)
	if session.StatusCode != 200 {
		httpPanic(session)
	}
	destURL := session.Header.Get("Location")

	// -- Upload Loop
	uploadedBytes := 0
	finished := false

	md5Verifier := md5.New() // md5 checksum 検証用
	for {
		// バッファがいっぱいになるかEOFまで読む
		uploadBody := make([]byte, 0, opts.ChunkSize)
		for {
			buf := make([]byte, opts.ChunkSize-len(uploadBody))
			n, err := os.Stdin.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			uploadBody = append(uploadBody, buf[0:n]...)
			finished = err == io.EOF
			if finished || len(uploadBody) >= opts.ChunkSize {
				break
			}
		}
		_, err := md5Verifier.Write(uploadBody)
		if err != nil {
			panic(err)
		}
		n := len(uploadBody)
		log.Printf("%d bytes readed\n", len(uploadBody))
		req, err := http.NewRequest("POST", destURL, bytes.NewBuffer(uploadBody))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Length", strconv.Itoa(n))
		if finished { // サイズ確定
			req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", uploadedBytes, uploadedBytes+n-1, uploadedBytes+n))
		} else {
			req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/*", uploadedBytes, uploadedBytes+n-1))
		}
		uploadRes, err := client.Do(req)
		log.Println(uploadRes.Status)
		if uploadRes.StatusCode == 200 || uploadRes.StatusCode == 201 {
			resStr := readAllString(uploadRes.Body)
			fmt.Println(resStr)
			var resMap map[string]string
			if err = json.Unmarshal([]byte(resStr), &resMap); err != nil {
				panic(err)
			}
			myChecksum := hex.EncodeToString(md5Verifier.Sum([]byte{}))
			remoteChecksum := resMap["md5Checksum"]
			if myChecksum != remoteChecksum {
				log.Panicf("Remote Checksum Invalid (local: %s, remote: %s)\n", myChecksum, remoteChecksum)
			}
			os.Exit(0)
		} else if uploadRes.StatusCode == 308 {
		} else {
			// TODO: check server state and retry correctly
			httpPanic(uploadRes)
		}
		uploadedBytes += n
	}
}
