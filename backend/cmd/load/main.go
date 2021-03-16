package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

	roomId := uuid.New().String()

	body, err := json.Marshal(
		map[string]string{"id": roomId, "password": "test"},
	)
	if err != nil {
		return nil, "", err
	}

	req := &http.Request{
		Method: "POST",
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(bytes.NewReader(body)),
		URL: &url.URL{
			Scheme: "https",
			Host:   "localhost",
			Path:   "/api/v1/registration/create",
		},
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	client := &http.Client{
		Jar:       jar,
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	roomPath := resp.Header.Get("Content-Location")
	log.Println(resp.StatusCode, resp.Status)
	roomId = roomPath[strings.LastIndex(roomPath, "?")+len("?roomId="):]

	return jar, roomId, nil

}

type duration struct {
	kind   string
	length float64
}

func main() {
	endNum := 10000
	newRoomTicker := time.NewTicker(5 * time.Second)
	numRoomsTicker := time.NewTicker(1 * time.Second)
	done := make(chan struct{})
	var numRooms uint64 = 0

	go func() {
		var secs uint
		for {
			<-numRoomsTicker.C
			n := atomic.LoadUint64(&numRooms)
			fmt.Printf("snapshot at %ds - running: %d rooms\n", secs, n)
			secs++
		}
	}()

	for {
		<-newRoomTicker.C

		for k := 0; k < 20; k++ {
			atomic.AddUint64(&numRooms, 1)
			n := atomic.LoadUint64(&numRooms)

			if n > uint64(endNum) {
				<-done
			}

			go func() {
				jar, roomId, err := createRoom()
				if err != nil {
					log.Println("err creating room:", err)
					return
				}

				for i := 0; i < 10; i++ {
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
							panic(err)
						}

						go func() {
							var stateToSend data.State

							baseState := fmt.Sprintf(baseState, roomId)

							if err := json.Unmarshal(
								[]byte(baseState),
								&stateToSend,
							); err != nil {
								panic(err)
							}

							writeTicker := time.NewTicker(30 * time.Second)

							for {
								<-writeTicker.C
								if err := c.WriteJSON(stateToSend); err != nil {
									log.Println(err)
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
									log.Println(err)
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
