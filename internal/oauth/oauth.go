package oauth

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"log"
	"os"
)

func InitTokenSource(sa string, subject string) oauth2.TokenSource {
	key, err := ioutil.ReadFile(sa)
	if err != nil {
		log.Fatalln("sa file not found", err)
	}
	creds, err := google.CredentialsFromJSONWithParams(context.Background(), key, google.CredentialsParams{
		Scopes:  []string{drive.DriveScope, admin.AdminDirectoryGroupScope},
		Subject: subject,
	})
	if err != nil {
		log.Fatalln("Credentials from JSON", err)
	}
	tokenFile := "token.json"
	ts, err := tokenFromFile(tokenFile, creds.TokenSource)
	if err != nil {
		ts = getTokenFromWeb(tokenFile, creds.TokenSource)
	}
	return ts
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(file string, ts oauth2.TokenSource) oauth2.TokenSource {
	t, err := ts.Token()
	if err != nil {
		log.Fatalln("Unable to retrieve token from web", err)
	}
	//fmt.Printf("token from web \nAccessToken: %s \nRefreshToken: %s\n", t.AccessToken, t.RefreshToken)
	saveToken(file, t)
	return oauth2.ReuseTokenSource(t, ts)
}

// Retrieves a token from a local file.
func tokenFromFile(file string, ts oauth2.TokenSource) (oauth2.TokenSource, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("token from file\n: AccessToken: %s \n RefreshToken: %s\n", t.AccessToken, t.RefreshToken)
	// TODO Don't wrap a reuseTokenSource in itself ?
	return oauth2.ReuseTokenSource(t, ts), err
}

// Saves a token to a file path.
func saveToken(file string, token *oauth2.Token) {
	//fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
