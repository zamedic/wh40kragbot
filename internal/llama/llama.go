package llama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"
)

type Parse struct {
	textDirectory  string
	imageDirectory string
	llamaApiKey    string
}

func NewLlamaParse(textDirectory, imageDirectory, llamaApiKey string) *Parse {
	return &Parse{
		textDirectory:  textDirectory,
		imageDirectory: imageDirectory,
		llamaApiKey:    llamaApiKey,
	}
}

func (l *Parse) Parse(ctx context.Context, filePath string) error {
	response, err := l.uploadFile(ctx, filePath)
	if err != nil {
		return err
	}
	l.monitorFile(ctx, response, filePath)
	return nil
}

func (l *Parse) Resume(ctx context.Context, jobId, fileName string) {
	s := &StatusResponse{
		Id: jobId,
	}
	l.monitorFile(ctx, s, fileName)
}

func (l *Parse) monitorFile(ctx context.Context, response *StatusResponse, filePath string) {
	for {
		errorCount := 0
		status, err := getJobStatus(response.Id, l.llamaApiKey)
		if err != nil {
			errorCount++
			if errorCount > 5 {
				log.Fatalf("Error getting job status: %v", err)
			}
			time.Sleep(5 * time.Second)
		}

		switch status.Status {
		case "SUCCESS", "PARTIAL_SUCCESS":
			l.readFile(ctx, response, l.textDirectory, filePath)
			return
		case "ERROR":
			log.Printf("Job %s failed with error: %s", response.Id, status.ErrorMessage)
			return
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

func (l *Parse) readFile(ctx context.Context, response *StatusResponse, textDir string, doc string) {
	result, err := l.getJobResultText(response.Id, "raw/text")
	if err != nil {
		log.Fatalf("Error getting job result: %v", err)
	}
	err = os.WriteFile(filepath.Join(l.textDirectory, doc+".txt"), []byte(result), 0644)
	if err != nil {
		log.Fatalf("Error writing result to file: %v", err)
	}

	result, err = l.getJobResultText(response.Id, "raw/markdown")
	if err != nil {
		log.Fatalf("Error getting job result: %v", err)
	}
	err = os.WriteFile(filepath.Join(l.textDirectory, doc+".md"), []byte(result), 0644)
	if err != nil {
		log.Fatalf("Error writing result to file: %v", err)
	}

	result, err = l.getJobResultText(response.Id, "json")
	if err != nil {
		log.Fatalf("Error getting job result: %v", err)
	}
	err = os.WriteFile(filepath.Join(l.textDirectory, doc+".json"), []byte(result), 0644)
	if err != nil {
		log.Fatalf("Error writing result to file: %v", err)
	}

	l.images(ctx, response, result, doc)
}

func (l *Parse) images(ctx context.Context, response *StatusResponse, result string, doc string) {
	ls := &LlamaParse{}
	err := json.Unmarshal([]byte(result), ls)
	if err != nil {
		log.Fatalf("Error unmarshalling json: %v", err)
	}
	for _, page := range ls.Pages {
		for _, img := range page.Images {
			l.downloadImage(ctx, response.Id, doc, img)
		}
	}
}

func (l *Parse) downloadImage(ctx context.Context, id string, name string, img LlamaImage) {
	url := fmt.Sprintf("https://api.cloud.llamaindex.ai/api/parsing/job/%s/result/image/%s", id, img.Name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.llamaApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(b))
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading image file: %v", err)
	}
	err = os.WriteFile(filepath.Join(l.imageDirectory, name, img.Name), b, 0644)
	if err != nil {
		log.Panicf("Error saving image file: %v", err)
	}
}

// uploadFile uploads a file to the specified URL using a POST request with multipart/form-data
func (l *Parse) uploadFile(ctx context.Context, filePath string) (*StatusResponse, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err = writer.WriteField("take_screenshot", "true")
	if err != nil {
		return nil, err
	}
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	_, err = part.Write(b)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.cloud.llamaindex.ai/api/parsing/upload", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.llamaApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to upload file: %s", resp.Status)
	}
	response := &StatusResponse{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// getJobStatus retrieves the status of a job from the LlamaIndex API
func getJobStatus(jobID string, apiKey string) (*StatusResponse, error) {
	url := fmt.Sprintf("https://api.cloud.llamaindex.ai/api/parsing/job/%s", jobID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get job status: %s", resp.Status)
	}

	response := &StatusResponse{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	return response, nil

}

// getJobResultMarkdown retrieves the result of a job in markdown format from the LlamaIndex API
func (l *Parse) getJobResultText(jobID string, resultType string) (string, error) {
	url := fmt.Sprintf("https://api.cloud.llamaindex.ai/api/parsing/job/%s/result/%s", jobID, resultType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.llamaApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get job result in markdown: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

type StatusResponse struct {
	Id           string `json:"id"`
	Status       string `json:"status"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
