package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/esteth/magic/disney"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

type attraction struct {
}

// Run sets up HTTP hhndlers.
func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/data", handleData)
	appengine.Main()
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	currentTimeStamp := int64(time.Now().Unix())

	parks := []disney.Park{
		disney.NewMagicKingdom(),
		disney.NewEpcot(),
		disney.NewHollywoodStudios(),
		disney.NewAnimalKingdom(),
	}

	for _, park := range parks {
		waitTimes, err := park.FetchWaitTimes(c, currentTimeStamp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, waitTime := range waitTimes {
			err = datastore.RunInTransaction(c, func(c context.Context) error {
				var attraction attraction
				attractionKey := datastore.NewKey(c, "Attraction", waitTime.AttractionID, 0, nil)
				err := datastore.Get(c, attractionKey, &attraction)
				if err == datastore.ErrNoSuchEntity {
					datastore.Put(c, attractionKey, &attraction)
				} else if err != nil {
					return err
				}

				newKey := datastore.NewKey(c, "WaitTime", "", currentTimeStamp, attractionKey)
				_, err = datastore.Put(c, newKey, &waitTime)
				if err != nil {
					return err
				}
				return nil
			}, nil)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, waitTime := range waitTimes {
			fmt.Fprintf(w, "%d - %d minutes\n", waitTime.AttractionID, waitTime.PostedWait)
		}
	}
}

func handleData(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}

	if !u.Admin {
		http.Error(w, "Admin Only", http.StatusUnauthorized)
		return
	}

	attractionKey := datastore.NewKey(c, "Attraction", "Haunted Mansion", 0, nil)
	var waitTimes []disney.WaitTime
	_, err := datastore.NewQuery("WaitTime").Ancestor(attractionKey).GetAll(c, &waitTimes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encodedWaitTimes, err := json.Marshal(waitTimes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(encodedWaitTimes))
}
