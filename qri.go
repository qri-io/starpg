package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/qri-io/dataset"

	"github.com/libp2p/go-libp2p-crypto"
	"github.com/qri-io/cafs"
	"github.com/qri-io/qri/actions"
	"github.com/qri-io/qri/config"
	"github.com/qri-io/qri/p2p"
	"github.com/qri-io/qri/repo"
	"github.com/qri-io/qri/repo/profile"
)

var (
	node     *p2p.QriNode
	nodeLock sync.Mutex
	testPk   = []byte(`CAASpgkwggSiAgEAAoIBAQC/7Q7fILQ8hc9g07a4HAiDKE4FahzL2eO8OlB1K99Ad4L1zc2dCg+gDVuGwdbOC29IngMA7O3UXijycckOSChgFyW3PafXoBF8Zg9MRBDIBo0lXRhW4TrVytm4Etzp4pQMyTeRYyWR8e2hGXeHArXM1R/A/SjzZUbjJYHhgvEE4OZy7WpcYcW6K3qqBGOU5GDMPuCcJWac2NgXzw6JeNsZuTimfVCJHupqG/dLPMnBOypR22dO7yJIaQ3d0PFLxiDG84X9YupF914RzJlopfdcuipI+6gFAgBw3vi6gbECEzcohjKf/4nqBOEvCDD6SXfl5F/MxoHurbGBYB2CJp+FAgMBAAECggEAaVOxe6Y5A5XzrxHBDtzjlwcBels3nm/fWScvjH4dMQXlavwcwPgKhy2NczDhr4X69oEw6Msd4hQiqJrlWd8juUg6vIsrl1wS/JAOCS65fuyJfV3Pw64rWbTPMwO3FOvxj+rFghZFQgjg/i45uHA2UUkM+h504M5Nzs6Arr/rgV7uPGR5e5OBw3lfiS9ZaA7QZiOq7sMy1L0qD49YO1ojqWu3b7UaMaBQx1Dty7b5IVOSYG+Y3U/dLjhTj4Hg1VtCHWRm3nMOE9cVpMJRhRzKhkq6gnZmni8obz2BBDF02X34oQLcHC/Wn8F3E8RiBjZDI66g+iZeCCUXvYz0vxWAQQKBgQDEJu6flyHPvyBPAC4EOxZAw0zh6SF/r8VgjbKO3n/8d+kZJeVmYnbsLodIEEyXQnr35o2CLqhCvR2kstsRSfRz79nMIt6aPWuwYkXNHQGE8rnCxxyJmxV4S63GczLk7SIn4KmqPlCI08AU0TXJS3zwh7O6e6kBljjPt1mnMgvr3QKBgQD6fAkdI0FRZSXwzygx4uSg47Co6X6ESZ9FDf6ph63lvSK5/eue/ugX6p/olMYq5CHXbLpgM4EJYdRfrH6pwqtBwUJhlh1xI6C48nonnw+oh8YPlFCDLxNG4tq6JVo071qH6CFXCIank3ThZeW5a3ZSe5pBZ8h4bUZ9H8pJL4C7yQKBgFb8SN/+/qCJSoOeOcnohhLMSSD56MAeK7KIxAF1jF5isr1TP+rqiYBtldKQX9bIRY3/8QslM7r88NNj+aAuIrjzSausXvkZedMrkXbHgS/7EAPflrkzTA8fyH10AsLgoj/68mKr5bz34nuY13hgAJUOKNbvFeC9RI5g6eIqYH0FAoGAVqFTXZp12rrK1nAvDKHWRLa6wJCQyxvTU8S1UNi2EgDJ492oAgNTLgJdb8kUiH0CH0lhZCgr9py5IKW94OSM6l72oF2UrS6PRafHC7D9b2IV5Al9lwFO/3MyBrMocapeeyaTcVBnkclz4Qim3OwHrhtFjF1ifhP9DwVRpuIg+dECgYANwlHxLe//tr6BM31PUUrOxP5Y/cj+ydxqM/z6papZFkK6Mvi/vMQQNQkh95GH9zqyC5Z/yLxur4ry1eNYty/9FnuZRAkEmlUSZ/DobhU0Pmj8Hep6JsTuMutref6vCk2n02jc9qYmJuD7iXkdXDSawbEG6f5C4MUkJ38z1t1OjA==`)
	privKey  crypto.PrivKey

	testPeerProfile = &profile.Profile{
		Peername: "peer",
		ID:       "QmZePf5LeXow3RW5U1AgEiNbW46YnRGhZ7HPvm1UmPFPwt",
	}
)

func init() {
	data, err := base64.StdEncoding.DecodeString(string(testPk))
	if err != nil {
		panic(err)
	}
	testPk = data

	privKey, err = crypto.UnmarshalPrivateKey(testPk)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling private key: %s", err.Error()))
	}
	testPeerProfile.PrivKey = privKey

	mr, err := repo.NewMemRepo(testPeerProfile, cafs.NewMapstore(), profile.NewMemStore(), nil)
	if err != nil {
		panic(err)
	}
	node, err = p2p.NewQriNode(mr, config.DefaultP2PForTesting())
	if err != nil {
		panic(err)
	}

}

// ExecQriTransformHandler executes a qri transform in memory
func ExecQriTransformHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// use a mutex b/c we want this node to only correspond with one request at
	// a time
	nodeLock.Lock()
	defer nodeLock.Unlock()

	node.LocalStreams = p2p.IOStreams{Out: w, ErrOut: w}

	f, err := ioutil.TempFile("", "exec_skylark")
	if err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	defer os.Remove(f.Name())
	if _, err := io.Copy(f, r.Body); err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	dsp := &dataset.DatasetPod{
		Structure: &dataset.StructurePod{
			Format: "json",
			Schema: map[string]interface{}{"type": "array"},
		},
		Transform: &dataset.TransformPod{
			ScriptPath: f.Name(),
		},
		Meta: &dataset.Meta{},
	}

	if str := r.FormValue("secrets"); str != "" {
		dsp.Transform.Secrets, err = parseSecrets(strings.Split(str, ",")...)
		if err != nil {
			log.Error(err.Error())
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}

	if str := r.FormValue("config"); str != "" {
		cfg := map[string]interface{}{}
		sec, err := parseSecrets(strings.Split(str, ",")...)
		if err != nil {
			log.Error(err.Error())
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		for key, val := range sec {
			cfg[key] = val
		}
		dsp.Transform.Config = cfg
	}

	ds, body, secrets, err := actions.NewDataset(dsp)
	if err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	ref, err := actions.CreateDataset(node, "test_dataset", ds, body, secrets, false)
	if err != nil {
		err = fmt.Errorf(strings.Replace(err.Error(), f.Name(), "line", 1))
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := actions.ReadDataset(node.Repo, &ref); err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	_, bodyBytes, err := actions.LookupBody(node, ref.Path, dataset.JSONDataFormat, nil, 0, 0, true)
	if err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	ref.Dataset.Body = json.RawMessage(bodyBytes)

	data, err := json.MarshalIndent(ref.Dataset, "", "\t")
	if err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Write(data)

	if err := actions.DeleteDataset(node, &ref); err != nil {
		log.Error(err.Error())
	}
}

// parseSecrets turns a key,value sequence into a map[string]string
func parseSecrets(secrets ...string) (map[string]string, error) {
	if len(secrets)%2 != 0 {
		return nil, fmt.Errorf("expected even number of (key,value) pairs for secrets")
	}
	s := map[string]string{}
	for i := 0; i < len(secrets); i = i + 2 {
		s[secrets[i]] = secrets[i+1]
	}
	return s, nil
}
