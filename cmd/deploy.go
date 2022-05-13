package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/capeprivacy/cli/progress"
	czip "github.com/capeprivacy/cli/zip"
	"github.com/capeprivacy/go-kit/id"
	"github.com/gosuri/uiprogress"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type DeployRequest struct {
	Data  []byte `json:"data"`
	Name  string `json:"name"`
	Nonce string `json:"nonce"`
}

type DeployResponse struct {
	ID                  string `json:"id"`
	AttestationDocument string `json:"attestation_document"`
}

// runCmd represents the request command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy a function",
	Run:   deploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.PersistentFlags().StringP("token", "t", "", "token to use")
}

func deploy(cmd *cobra.Command, args []string) {
	u, err := cmd.Flags().GetString("url")
	if err != nil {
		log.Errorf("flag not found: %s", err)
	}

	if len(args) != 2 {
		log.Error("expected two arguments, name of function and path to a directory to zip")
	}

	name := args[0]
	functionDir := args[1]

	if len(name) == 0 {
		log.Error("function name cannot be empty")
		return
	}

	file, err := os.Open(functionDir)
	if err != nil {
		log.Errorf("unable to read function directory: %s", err)
		return
	}

	st, err := file.Stat()
	if err != nil {
		log.Errorf("unable to read function directory: %s", err)
		return
	}

	if !st.IsDir() {
		log.Errorf("expected argument %s to be a directory", functionDir)
		return
	}

	_, err = file.Readdirnames(1)
	if err != nil {
		log.Errorf("please pass in a non-empty directory: %s", err)
		return
	}

	err = file.Close()
	if err != nil {
		log.Errorf("something went wrong: %s", err)
		return
	}

	zipRoot := filepath.Base(functionDir)

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	err = filepath.Walk(functionDir, czip.Walker(w, zipRoot))
	if err != nil {
		log.Errorf("zipping directory failed: %s", err)
		return
	}

	// explicitly close now so that the bytes are flushed and
	// available in buf.Bytes() below.
	err = w.Close()
	if err != nil {
		log.Errorf("zipping directory failed: %s", err)
		return
	}

	enclave, err := doStart(u)
	if err != nil {
		log.Errorf("unable to start enclave %s", err)
		return
	}

	fmt.Println("enclave started ...")
	id, err := doDeploy(u, enclave.id, name, buf.Bytes())
	if err != nil {
		log.Errorf("unable to deploy function %s", err)
		return
	}

	fmt.Printf("Successfully deployed function. Function ID: %s\n", id)
}

func doDeploy(url string, id id.ID, name string, data []byte) (string, error) {
	reqData := DeployRequest{
		Name:  name,
		Data:  data,
		Nonce: getNonce(),
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/v1/deploy/%s", url, id)
	buffer := bytes.NewBuffer(body)

	fmt.Println("Uploading zip ...")
	uiprogress.Start()
	bar := uiprogress.AddBar(100)
	bar.AppendCompleted()
	pr := &progress.Reader{Reader: buffer, Size: int64(buffer.Len()), Reporter: func(progress float64) {
		p := int(progress * 100)
		if err := bar.Set(p); err != nil {
			fmt.Println("Upload progress:", p)
		}
	}}
	req, err := http.NewRequest("POST", endpoint, pr)
	if err != nil {
		return "", fmt.Errorf("unable to create request %s", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed %s", err)
	}

	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("bad status code %d", res.StatusCode)
	}

	resData := DeployResponse{}

	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&resData)
	if err != nil {
		return "", fmt.Errorf("unable to decode response %s", err)
	}

	return resData.ID, nil
}
