package drive

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

const (
	uploadChunkSize = 64 * 1024 * 1024
)

var service *drive.Service

func InitService(ts oauth2.TokenSource) {
	var err error
	service, err = drive.NewService(context.Background(), option.WithTokenSource(ts))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func CreateDrive(names []string, count int, group string, user string) {
	for i, name := range names {
		if count > 1 {
			for j := 1; j <= count; j++ {
				doCreateDrive(i, fmt.Sprintf("%s-%d", names[0], j), group, user)
			}
		} else {
			doCreateDrive(i, name, group, user)
		}
	}
}

func doCreateDrive(index int, name, group, user string) {
	if d, err := service.Drives.Create(name, &drive.Drive{
		Name: name,
	}).Fields("id").Do(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%d Create drive id: %s name: %s [OK]\n", index, d.Id, name)
		addDriveGroup(d.Id, group)
		addDriveUser(d.Id, user)
	}
}

func addDriveGroup(driveId, group string) {
	if group != "" {
		if _, err := service.Permissions.Create(driveId, &drive.Permission{
			EmailAddress: group,
			Role:         "organizer", // owner organizer fileOrganizer writer commenter reader
			Type:         "group",     // user group domain anyone
		}).Fields().SupportsAllDrives(true).Do(); err != nil {
			fmt.Println("add drive group error", err)
		}
	}
}

func addDriveUser(driveId, user string) {
	if user != "" {
		if _, err := service.Permissions.Create(driveId, &drive.Permission{
			EmailAddress: user,
			Role:         "organizer", // owner organizer fileOrganizer writer commenter reader
			Type:         "user",      // user group domain anyone
		}).Fields().SupportsAllDrives(true).Do(); err != nil {
			fmt.Println("add drive user error", err)
		}
	}
}

func List(driveId string) {
	if driveId == "" {
		listDrives()
	} else {
		listFiles(driveId)
	}
}

func listDrives() {
	fmt.Println("list drives")
	if list, err := service.Drives.List().Fields("drives/id", "drives/name").Do(); err != nil {
		fmt.Println(err)
	} else {
		for i, v := range list.Drives {
			fmt.Printf("%d id: %s name: %s\n", i, v.Id, v.Name)
		}
	}
}

func listFiles(driveId string) {
	fmt.Println("List drive's files:", driveId)
	if list, err := service.Files.List().Fields().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Corpora("drive").
		Spaces("drive").
		Q("trashed=false").
		PageSize(1000). //Default: 100
		Fields("nextPageToken,files(id,name,mimeType,size)").
		DriveId(driveId).Do(); err != nil {
		fmt.Println(err)
	} else {
		for i, file := range list.Files {
			fmt.Printf("%d id: %s name: %s size: %d\n", i, file.Id, file.Name, file.Size)
		}
	}
}

func DeleteDrive(driveIds []string, force bool) {
	for i, driveId := range driveIds {
		err := service.Drives.Delete(driveId).Fields().Do()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%d delete drive: %s force: %v [OK]\n", i, driveId, force)
		}
	}
}

func Cat(fileId string, ranges string, count int, quiet bool) {
	for i := 0; i < count; i++ {
		doCat(fileId, ranges, quiet)
	}
}

func doCat(fileId string, ranges string, quiet bool) {
	call := service.Files.Get(fileId).Fields()
	if ranges != "" {
		call.Header().Set("Range", "bytes="+ranges)
	}
	start := time.Now()
	if res, err := call.Download(); err != nil {
		log.Panicln(err)
	} else {
		defer res.Body.Close()
		if !quiet {
			reader := bufio.NewReaderSize(res.Body, googleapi.MinUploadChunkSize)
			reader.WriteTo(os.Stdout)
		}
	}
	fmt.Printf("\nCat file: %s range: %s time:%s \n", fileId, ranges, time.Since(start))
}

func Copy(filepath string, driveId string) error {
	media, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Open file err", filepath, err)
		return err
	}
	defer media.Close()
	stat, err := media.Stat()
	if err != nil {
		fmt.Println("File stat err", filepath, err)
		return err
	}
	meta := &drive.File{
		Name: stat.Name(),
		//FileExtension: path.Ext(filepath),
		//FullFileExtension: "",
		Parents: []string{driveId},
	}
	reader := bufio.NewReaderSize(media, uploadChunkSize)
	file, err := service.Files.Create(meta).SupportsAllDrives(true).Fields("id").Media(reader).Do()
	if err != nil {
		fmt.Println("Upload file err", err)
		return err
	}
	fmt.Printf("Upload file id: %s name: %s [OK]\n", file.Id, stat.Name())
	return nil
}

func CopyRemote(fileId string, driveId string) {
	src, err := service.Files.Get(fileId).Fields("name").SupportsAllDrives(true).Do()
	if err != nil {
		fmt.Println("CopyRemote get error", err)
		return
	}
	_, err = service.Files.Copy(fileId, &drive.File{
		Name:    src.Name,
		Parents: []string{driveId},
	}).Fields().SupportsAllDrives(true).Do()
	if err != nil {
		fmt.Println("CopyRemote error", err)
		return
	} else {
		fmt.Println("CopyRemote [OK]", src.Name)
	}
}

func Move(filepath string, driveId string) error {
	err := Copy(filepath, driveId)
	if err == nil {
		fmt.Println("remove local file", os.Remove(filepath))
	}
	return err
}

func Remove(fileIds []string) {
	for i, id := range fileIds {
		if err := service.Files.Delete(id).SupportsAllDrives(true).Fields().Do(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%d Delete file: %s [OK]\n", i, id)
		}
	}
}
