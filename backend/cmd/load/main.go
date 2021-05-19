package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
)

var baseState = `{
    "ws": null,
    "connected": false,
    "errorMessage": "",
    "roomId": "%s",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ]
}`

func createRoom() (*cookiejar.Jar, string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, "", err
	}

	rId := uuid.New().String()

	b, err := json.Marshal(
		map[string]string{"id": rId, "password": "test"},
	)
	if err != nil {
		return nil, "", err
	}

	r := &http.Request{
		Method: "POST",
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(bytes.NewReader(b)),
		URL: &url.URL{
			Scheme: "https",
			Host:   "localhost",
			Path:   "/api/v1/registration/create",
		},
	}

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, "", fmt.Errorf(
			"creating room failed with status code %d",
			resp.StatusCode,
		)
	}

	p := resp.Header.Get("Content-Location")
	rId = p[strings.LastIndex(p, "?")+len("?roomId="):]

	return jar, rId, nil
}

func main() {
	var (
		endNumRooms       uint64 = 1000
		numRooms          uint64 = 0
		numClientsPerRoom        = 10
		numRoomsPerEpoch         = 20
		epochTicker              = time.NewTicker(5 * time.Second)
		progressTicker           = time.NewTicker(1 * time.Second)
	)

	go func() {
		var secs int

		for {
			<-progressTicker.C

			n := atomic.LoadUint64(&numRooms)
			fmt.Printf("snapshot at %ds - running: %d rooms (%d clients) \n", secs, n, numClientsPerRoom*int(n))
			secs++
		}
	}()

	for {
		<-epochTicker.C

		for i := 0; i < numRoomsPerEpoch; i++ {
			atomic.AddUint64(&numRooms, 1)
			n := atomic.LoadUint64(&numRooms)

			if n > endNumRooms {
				fmt.Println("end num rooms reached")
				return
			}

			go func() {
				jar, roomId, err := createRoom()
				if err != nil {
					fmt.Println("err creating room: ", err)
					return
				}

				for j := 0; j < numClientsPerRoom; j++ {
					go func() {
						dialer := websocket.Dialer{
							Jar: jar,
							TLSClientConfig: &tls.Config{
								InsecureSkipVerify: true,
							},
						}
						c, _, err := dialer.Dial(
							fmt.Sprintf(
								"%s%s%s",
								"wss://localhost",
								"/api/v1/retrospectives/",
								roomId),
							nil,
						)
						if err != nil {
							fmt.Println("err connecting to room: ", err)
							return
						}

						go func() {
							var stateToSend data.State

							baseState := fmt.Sprintf(baseState, roomId)

							if err := json.Unmarshal(
								[]byte(baseState),
								&stateToSend,
							); err != nil {
								panic(fmt.Sprintf("err unmarshaling state: %s", err))
							}

							writeTicker := time.NewTicker(10 * time.Second)

							for {
								<-writeTicker.C

								if err := c.WriteJSON(stateToSend); err != nil {
									fmt.Println("err writing: ", err)
									return
								}
							}
						}()

						go func() {
							var stateToReceive data.State

							for {
								if err := c.ReadJSON(
									&stateToReceive,
								); err != nil {
									fmt.Println("err reading: ", err)
									return
								}
							}
						}()
					}()
				}
			}()
		}
	}
}
