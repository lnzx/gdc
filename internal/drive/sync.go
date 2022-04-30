package drive

import (
	"bufio"
	"context"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

var sas []*string
var n int
var mu sync.RWMutex

func config() {
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

func Sync(dir, suffix, replace, driveId string) {
	config()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case e, ok := <-watcher.Events:
				if !ok {
					return
				}
				switch e.Op {
				case fsnotify.Create:
					name := e.Name
					if strings.HasSuffix(name, suffix) {
						log.Println("-->Create file: ", name)
						final := strings.Replace(name, suffix, replace, 1)
						if os.Rename(name, final) != nil {
							log.Printf("Error: Rename %v -> %v\n", name, final)
							return
						}
						log.Printf("Rename %v -> %v [OK]\n", name, final)
					} else if strings.HasSuffix(name, replace) {
						if upload(name, driveId) == nil {
							log.Printf("<--Delete: %v\n", os.Remove(name))
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func upload(filepath string, driveId string) (err error) {
	log.Println("Upload:", filepath)
	media, err := os.Open(filepath)
	if err != nil {
		log.Println("Open file err", err)
		return
	}
	defer media.Close()
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
	log.Printf("<--Upload [OK]: %s\n", meta.Name)
	return nil
}
