package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"io"

	"codewax/internal/models"
	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

const (
	ChunkSize    = 100
	ChunkOverlap = 20
)

type FileChunk struct {
	FilePath  string
	StartLine int
	EndLine   int
	Content   string
}

var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".next":        true,
	"dist":         true,
	"build":        true,
	"__pycache__":  true,
}

var ignoredFiles = map[string]bool{
	".env":          true,
	".gitignore":    true,
	".dockerignore": true,
	"swagger.json":  true,
	"swagger.yaml":  true,
}

var ignoredExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".ico":  true,
	".pdf":  true,
	".zip":  true,
	".tar":  true,
	".gz":   true,
	".exe":  true,
	".bin":  true,
	".lock": true,
	".sum":  true,
	".mod":  true,
}

func fetchEmbeddingBatch(client *http.Client, texts []string) ([][]float32, int, error) {
	const maxRetries = 5
	backoff := time.Second

	for attempt := range maxRetries {
		embeddings, tokens, err := doFetchEmbeddingBatch(client, texts)
		if err == nil {
			return embeddings, tokens, nil
		}

		if !strings.Contains(err.Error(), "status 429") {
			return nil, 0, err
		}

		if attempt == maxRetries-1 {
			return nil, 0, fmt.Errorf("exceeded max retries: %w", err)
		}

		log.Printf("rate limited by voyage, retrying in %s (attempt %d/%d)", backoff, attempt+1, maxRetries)
		time.Sleep(backoff)
		backoff *= 2
	}

	return nil, 0, fmt.Errorf("unreachable")
}

func doFetchEmbeddingBatch(client *http.Client, texts []string) ([][]float32, int, error) {
	reqBody, err := json.Marshal(map[string]any{
		"model": "voyage-code-3",
		"input": texts,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("VOYAGE_API_KEY"))

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("voyage request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("voyage returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode voyage response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = d.Embedding
	}
	return embeddings, result.Usage.TotalTokens, nil
}

func embedChunks(chunks []FileChunk) ([]FileChunk, [][]float32, int, error) {
    var filtered []FileChunk
    for _, chunk := range chunks {
        if strings.TrimSpace(chunk.Content) != "" {
            filtered = append(filtered, chunk)
        }
    }

    const batchSize = 10
    var allEmbeddings [][]float32
    totalTokens := 0
    client := &http.Client{}

    for i := 0; i < len(filtered); i += batchSize {
        end := min(i+batchSize, len(filtered))

        texts := make([]string, end-i)
        for j, chunk := range filtered[i:end] {
            texts[j] = chunk.Content
        }

        batchEmbeddings, tokens, err := fetchEmbeddingBatch(client, texts)
        if err != nil {
            return nil, nil, 0, fmt.Errorf("batch %d failed: %w", i/batchSize, err)
        }
        allEmbeddings = append(allEmbeddings, batchEmbeddings...)
        totalTokens += tokens
    }

    return filtered, allEmbeddings, totalTokens, nil
}

func chunkFile(filePath string, content string) []FileChunk {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	var chunks []FileChunk

	start := 0
	for start < totalLines {
		end := min(start+ChunkSize, totalLines)

		chunkContent := strings.Join(lines[start:end], "\n")

		chunks = append(chunks, FileChunk{
			FilePath:  filePath,
			StartLine: start + 1,
			EndLine:   end,
			Content:   chunkContent,
		})

		start += ChunkSize - ChunkOverlap

		if totalLines-start < ChunkOverlap {
			break
		}
	}

	return chunks
}

func walkAndChunk(tempDir string, repoID uint) ([]FileChunk, error) {
	var chunks []FileChunk

	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if ignoredDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if ignoredFiles[info.Name()] {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ignoredExtensions[ext] {
			return nil
		}

		if info.Size() > 500*1024 {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return nil
		}

		fileChunks := chunkFile(relPath, string(content))
		chunks = append(chunks, fileChunks...)

		return nil
	})

	return chunks, err
}

func cloneRepo(url string, branch string, destDir string) error {
	cmd := exec.Command("git", "clone", "--branch", branch, "--single-branch", url, destDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repo: %w — stderr: %s", err, stderr.String())
	}
	return nil
}

func ProcessRepository(db *gorm.DB, repo models.Repository) {

	db.Model(&repo).Update("status", "processing")


	exceeded, err := ExceedsMonthlyBudget(db)
	if err != nil {
		log.Printf("[repo %d] failed to check budget: %v", repo.ID, err)
		db.Model(&repo).Update("status", "failed")
		return
	}
	if exceeded {
		log.Printf("[repo %d] monthly token budget exceeded, skipping", repo.ID)
		db.Model(&repo).Update("status", "budget_exceeded")
		return
	}

	tempDir, err := os.MkdirTemp("", "codewax-repo-*")
	if err != nil {
		log.Printf("[repo %d] failed to create temp dir: %v", repo.ID, err)
		db.Model(&repo).Update("status", "failed")
		return
	}
	defer os.RemoveAll(tempDir)


	if err := cloneRepo(repo.GithubURL, repo.Branch, tempDir); err != nil {
		log.Printf("[repo %d] clone failed: %v", repo.ID, err)
		db.Model(&repo).Update("status", "failed")
		return
	}

	chunks, err := walkAndChunk(tempDir, repo.ID)
	if err != nil {
		log.Printf("[repo %d] walk failed: %v", repo.ID, err)
		db.Model(&repo).Update("status", "failed")
		return
	}

	chunks, embeddings, totalTokens, err := embedChunks(chunks)
	if err != nil {
		log.Printf("[repo %d] embedding failed: %v", repo.ID, err)
		db.Model(&repo).Update("status", "failed")
		return
	}

	log.Printf("[repo %d] embedding success: %d embeddings, %d tokens used", repo.ID, len(embeddings), totalTokens)
	if err := RecordTokenUsage(db, repo.UserID, repo.ID, totalTokens); err != nil {
		log.Printf("[repo %d] failed to record token usage: %v", repo.ID, err)
	}

	dbChunks := make([]models.Chunk, len(chunks))
	for i, chunk := range chunks {
		dbChunks[i] = models.Chunk{
			RepositoryID: repo.ID,
			FilePath:     chunk.FilePath,
			StartLine:    chunk.StartLine,
			EndLine:      chunk.EndLine,
			Content:      chunk.Content,
			Embedding:    pgvector.NewVector(embeddings[i]),
		}
	}

	tx := db.Begin()
	if err := tx.Create(&dbChunks).Error; err != nil {
		tx.Rollback()
		log.Printf("[repo %d] failed to insert chunks: %v", repo.ID, err)
		db.Model(&repo).Update("status", "failed")
		return
	}
	tx.Commit()

	db.Model(&repo).Update("status", "ready")
	log.Printf("[repo %d] processing complete", repo.ID)
}