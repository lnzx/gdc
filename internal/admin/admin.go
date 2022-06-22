package admin

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
	"log"
	"os"
	"strings"
)

var service *admin.Service

func InitService(ts oauth2.TokenSource) {
	var err error
	service, err = admin.NewService(context.Background(), option.WithTokenSource(ts))
	if err != nil {
		fmt.Println("Unable to create admin service", err)
		os.Exit(1)
	}
}

func ListGroups(domain string) {
	list, err := service.Groups.List().Domain(domain).Do()
	if err != nil {
		log.Fatalln(err)
	}
	for i, group := range list.Groups {
		fmt.Println(i, group.Id, group.Email)
	}
}

func CreateGroup(email string) {
	group := &admin.Group{
		AdminCreated: false,
		Email:        email,
	}
	_, err := service.Groups.Insert(group).Do()
	if err != nil {
		fmt.Println(err)
	}
}

func AddGroupMember(group string, user string, filepath string) {
	if user != "" {
		doAddGroupMember(group, user)
		return
	}
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		user = strings.TrimSpace(scanner.Text())
		if user == "" {
			continue
		}
		doAddGroupMember(group, user)
	}
}

func doAddGroupMember(group string, user string) error {
	_, err := service.Members.Insert(group, &admin.Member{Email: user}).Fields().Do()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Add user: %s [OK]\n", user)
	}
	return err
}
