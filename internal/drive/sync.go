package drive

import (
	"bufio"
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
	PREFIX = "plot-k32-"
	PLOT   = ".plot"
	GZ     = ".gz"
)

var uploads = make(map[string]int)

func init() {
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
	log.Println("Init sa size: ", l)
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

func Sync(dir, driveId string, t time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Error:", err)
			err = nil
		}
	}()

	log.Println("Sync dir:", dir, "time:", t)
	readDir(dir, driveId)

	ticker := time.NewTicker(t)
	for {
		<-ticker.C
		readDir(dir, driveId)
	}
}

func readDir(dir, driveId string) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println("read dir error:", err)
		return
	}
	log.Println("->read dir")
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		filename := f.Name()
		if strings.HasSuffix(filename, PLOT) {
			newname := strings.Replace(filename, PREFIX, "", 1)
			newname = strings.Replace(newname, PLOT, GZ, 1)
			if err = os.Rename(dir+"/"+filename, dir+"/"+newname); err != nil {
				log.Println("Rename error:", err)
				continue
			}
			log.Println("Rename:", filename, "->", newname)
			go uploadTask(dir+"/"+newname, driveId)
		} else if strings.HasSuffix(filename, GZ) {
			go uploadTask(dir+"/"+filename, driveId)
		}
	}
}

func uploadTask(filepath string, driveId string) {
	if _, ok := uploads[filepath]; ok {
		return
	} else {
		uploads[filepath] = 1
	}
	log.Println("Upload:", filepath)
	media, err := os.Open(filepath)
	if err != nil {
		log.Println("Open file err", err)
		return
	}
	stat, err := media.Stat()
	if err != nil {
		log.Println("File stat err", err)
		return
	}
	sa := next()
	log.Println("use sa: ", *sa)
	service, err := drive.NewService(context.Background(), option.WithCredentialsFile(*sa))
	if err != nil {
		log.Println("Error: new drive service", err)
		return
	}
	meta := &drive.File{
		Name:    stat.Name(),
		Parents: []string{driveId},
	}
	reader := bufio.NewReaderSize(media, uploadChunkSize)
	_, err = service.Files.Create(meta).SupportsAllDrives(true).Fields().Media(reader).Do()
	if err != nil {
		log.Println("Upload err", err)
		return
	}
	media.Close()
	log.Printf("<--Upload [OK]: %s\n", meta.Name)

	if err = os.Remove(filepath); err != nil {
		log.Printf("<--Remove [ERROR]: %s\n", media.Name())
	}
}
