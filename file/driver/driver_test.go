package driver

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"math/rand"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"os"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/gluster/glusterd2/pkg/api"
// 	"github.com/gluster/glusterd2/pkg/restclient"
// 	"github.com/kubernetes-csi/csi-test/pkg/sanity"
// 	"github.com/sirupsen/logrus"
// )

// func init() {
// 	rand.Seed(time.Now().UnixNano())
// }

// func TestDriverSuite(t *testing.T) {
// 	socket := "/tmp/csi.sock"
// 	endpoint := "unix://" + socket
// 	if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
// 		t.Fatalf("failed to remove unix domain socket file %s, error: %s", socket, err)
// 	}
// 	//id := uuid.Parse("02dfdd19-e01e-46ec-a887-97b309a7dd2f")
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		switch r.Method {
// 		case "GET":
// 			if strings.Contains(r.URL.String(), "/v1/peers") {
// 				var resp api.PeerListResp

// 				resp[0] = api.PeerGetResp{
// 					Name: "dhcp43-209.lab.eng.blr.redhat.com",
// 					PeerAddresses: []string{
// 						"10.70.43.209:24008"},
// 					ClientAddresses: []string{
// 						"127.0.0.1:24007",
// 						"10.70.43.209:24007"},
// 					Online: true,
// 					PID:    24935,
// 					Metadata: map[string]string{
// 						"_zone": "02dfdd19-e01e-46ec-a887-97b309a7dd2f",
// 					},
// 				}
// 				resp = append(resp, api.PeerGetResp{
// 					Name: "dhcp43-209.lab.eng.blr.redhat.com",
// 					PeerAddresses: []string{
// 						"10.70.43.209:24008"},
// 					ClientAddresses: []string{
// 						"127.0.0.1:24007",
// 						"10.70.43.209:24007"},
// 					Online: true,
// 					PID:    24935,
// 					Metadata: map[string]string{
// 						"_zone": "02dfdd19-e01e-46ec-a887-97b309a7dd2f",
// 					},
// 				})
// 				err := json.NewEncoder(w).Encode(&resp)
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 			} else {
// 				if strings.Contains(r.URL.String(), "/v1/volume") {
// 					var resp api.VolumeListResp

// 					resp[0] = api.VolumeGetResp{
// 						//ID:       id,
// 						Name:     "test1",
// 						Metadata: map[string]string{"a": "b"},
// 					}

// 					resp = append(resp, api.VolumeGetResp{
// 						//ID:       id,
// 						Name:     "test1",
// 						Metadata: map[string]string{"a": "b"},
// 					})
// 					err := json.NewEncoder(w).Encode(&resp)
// 					if err != nil {
// 						t.Fatal(err)
// 					}
// 				}
// 			}

// 		case "DELETE":
// 		}

// 	}))

// 	defer ts.Close()

// 	url, _ := url.Parse(ts.URL)
// 	doClient := restclient.New(url.String(), "", "", "", false)
// 	d := Driver{
// 		endpoint: endpoint,
// 		client:   doClient,
// 		logger:   logrus.New(),
// 	}
// 	defer d.Stop()

// 	go d.Run()

// 	mntDir, err := ioutil.TempDir("", "mnt")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(mntDir)

// 	mntStageDir, err := ioutil.TempDir("", "mnt-stage")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(mntStageDir)

// 	cfg := &sanity.Config{
// 		StagingPath: mntStageDir,
// 		TargetPath:  mntDir,
// 		Address:     endpoint,
// 	}

// 	sanity.Test(t, cfg)
// }
