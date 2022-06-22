package drive

import (
	"bufio"
	"bytes"
	"context"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var sas []*string
var n int
var mu sync.RWMutex

const (
	PLOT = ".plot"
	GZ   = ".gz"
)

var uploads = make(map[string]int)

var headSvc *drive.Service

func initHead() {
	var err error
	headSvc, err = drive.NewService(context.Background(), option.WithCredentialsFile("head.json"))
	if err != nil {
		log.Fatalln("Error: new head service", err)
	}
}

func initSync() {
	initHead()

	dir := "sa"
	f, err := os.Stat(dir)
	if err != nil {
		log.Fatal(err)
	}
	if !f.IsDir() {
		log.Fatal("sa not a folder.")
	}
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, x := range fs {
		path := dir + "/" + x.Name()
		sas = append(sas, &path)
	}
	l := len(sas)
	if l == 0 {
		log.Fatal("Error: cannot found sa file.")
	}
	mu = sync.RWMutex{}
	log.Println("Sync init sa size: ", l)
}

func next() *string {
	mu.Lock()
	sa := sas[n]
	n++
	if n >= len(sas) {
		n = 0
	}
	mu.Unlock()
	return sa
}

func Sync(dir, driveId string, t time.Duration, parentId string) {
	initSync()

	defer func() {
		if err := recover(); err != nil {
			log.Println("Error:", err)
			err = nil
		}
	}()

	log.Println("Sync dir:", dir, "time:", t)
	readDir(dir, driveId, parentId)

	ticker := time.NewTicker(t)
	for {
		<-ticker.C
		readDir(dir, driveId, parentId)
	}
}

func readDir(dir, driveId, parentId string) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println("read dir error:", err)
		return
	}
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	log.Println("->scan dir")
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		filename := f.Name()
		if strings.HasSuffix(filename, PLOT) {
			newname := mixFilename(filename)
			if err = os.Rename(dir+filename, dir+newname); err != nil {
				log.Println("Rename error:", err)
				continue
			}
			log.Println("Rename:", filename, "->", newname)
			go uploadTask(dir, newname, driveId, parentId)
		} else if strings.HasSuffix(filename, GZ) {
			go uploadTask(dir, filename, driveId, parentId)
		}
	}
}

func uploadTask(dir, filename, driveId string, parentId string) {
	if _, ok := uploads[filename]; ok {
		return
	} else {
		uploads[filename] = 1
	}
	log.Println("Upload:", filename)
	filepath := dir + filename
	media, err := os.Open(filepath)
	if err != nil {
		log.Println("Open file err", err)
		return
	}

	head := make([]byte, kib64)
	_, err = media.Read(head)
	if err != nil {
		return
	}
	uploadHead(head, filename, parentId)

	sa := next()
	log.Println("use sa: ", *sa)
	svc, err := drive.NewService(context.Background(), option.WithCredentialsFile(*sa))
	if err != nil {
		log.Println("Error: new drive service", err)
		return
	}
	reader := bufio.NewReaderSize(media, uploadChunkSize)
	_, err = svc.Files.Create(&drive.File{
		Name:    filename,
		Parents: []string{driveId},
	}).SupportsAllDrives(true).Fields().Media(reader).Do()
	if err != nil {
		log.Println("Upload err", err)
		return
	}
	media.Close()
	log.Printf("<--Upload body [OK]: %s\n", filename)

	if err = os.Remove(filepath); err != nil {
		log.Printf("<--Remove [ERROR]: %s\n", media.Name())
	}
}

func uploadHead(head []byte, filename string, parentId string) {
	reader := bytes.NewReader(head)
	_, err := headSvc.Files.Create(&drive.File{
		Name:    filename,
		Parents: []string{parentId},
	}).Fields().Media(reader).Do()
	if err != nil {
		log.Println("Upload head err", err)
		return
	}
	log.Printf("<--Upload head [OK]: %s\n", filename)
}

func mixFilename(filename string) string {
	n := strings.LastIndex(filename, "-")
	if n != -1 {
		filename = filename[n+1:]
	}
	filename = strings.Replace(filename, PLOT, GZ, 1)
	return filename
}
