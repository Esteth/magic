package magic

import (
	"fmt"
	"net/http"

	"time"

	"github.com/esteth/magic/disney"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type attraction struct {
}

type waitInfo struct {
	PostedWait int  `datastore:",noindex"`
	Operating  bool `datastore:",noindex"`
}

// Run sets up HTTP hhndlers.
func Run() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Admin login forced through app.yaml
	// u := user.Current(c)
	// if u == nil {
	// 	url, err := user.LoginURL(c, r.URL.String())
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	w.Header().Set("Location", url)
	// 	w.WriteHeader(http.StatusFound)
	// 	return
	// }

	// if !u.Admin {
	// 	http.Error(w, "Admin Only", http.StatusUnauthorized)
	// 	return
	// }

	currentTimeStamp := int64(time.Now().Unix())

	waitTimes, err := disney.NewMagicKingdom().FetchWaitTimes(c)
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
			waitInfo := waitInfo{
				Operating:  waitTime.Operating,
				PostedWait: waitTime.PostedWait,
			}
			_, err = datastore.Put(c, newKey, &waitInfo)
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
