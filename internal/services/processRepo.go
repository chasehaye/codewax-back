package services

import (
    "os"
    "os/exec"
    "fmt"
	"path/filepath"
	"strings"
	"bytes"
	"encoding/binary"
	"net/http"
	"encoding/json"


    "gorm.io/gorm"
    "codewax/internal/models"
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
    ".env":            true,
    ".gitignore":      true,
    ".dockerignore":   true,
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






func embeddingToBytes(embedding []float32) []byte {
    buf := new(bytes.Buffer)
    for _, v := range embedding {
        binary.Write(buf, binary.LittleEndian, v)
    }
    return buf.Bytes()
}

func embedChunks(chunks []FileChunk) ([][]float32, error) {
    const batchSize = 128
    var allEmbeddings [][]float32

    client := &http.Client{}

    for i := 0; i < len(chunks); i += batchSize {
        end := min(i+batchSize, len(chunks))
        batch := chunks[i:end]

        texts := make([]string, len(batch))
        for j, chunk := range batch {
            texts[j] = chunk.Content
        }

        reqBody, err := json.Marshal(map[string]any{
            "model": "voyage-code-3",
            "input": texts,
        })
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request: %w", err)
        }

        req, err := http.NewRequest("POST", "https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(reqBody))
        if err != nil {
            return nil, fmt.Errorf("failed to create request: %w", err)
        }
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", "Bearer "+os.Getenv("VOYAGE_API_KEY"))

        resp, err := client.Do(req)
        if err != nil {
            return nil, fmt.Errorf("voyage request failed: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            return nil, fmt.Errorf("voyage returned status %d", resp.StatusCode)
        }

        var result struct {
            Data []struct {
                Embedding []float32 `json:"embedding"`
            } `json:"data"`
        }

        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            return nil, fmt.Errorf("failed to decode voyage response: %w", err)
        }

        for _, d := range result.Data {
            allEmbeddings = append(allEmbeddings, d.Embedding)
        }
    }

    return allEmbeddings, nil
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
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to clone repo: %w", err)
    }
    return nil
}

func ProcessRepository(db *gorm.DB, repo models.Repository) {

    db.Model(&repo).Update("status", "processing")


    tempDir, err := os.MkdirTemp("", "codewax-repo-*")
    if err != nil {
        db.Model(&repo).Update("status", "failed")
        return
    }
    defer os.RemoveAll(tempDir)


    if err := cloneRepo(repo.GithubURL, repo.Branch, tempDir); err != nil {
        db.Model(&repo).Update("status", "failed")
        return
    }

    chunks, err := walkAndChunk(tempDir, repo.ID)
	if err != nil {
		db.Model(&repo).Update("status", "failed")
		return
	}

    embeddings, err := embedChunks(chunks)
    if err != nil {
        db.Model(&repo).Update("status", "failed")
        return
    }

	for i, chunk := range chunks {
        if err := db.Create(&models.Chunk{
            RepositoryID: repo.ID,
            FilePath:     chunk.FilePath,
            StartLine:    chunk.StartLine,
            EndLine:      chunk.EndLine,
            Content:      chunk.Content,
            Embedding:    embeddingToBytes(embeddings[i]),
        }).Error; err != nil {
            db.Model(&repo).Update("status", "failed")
            return
        }
    }

    db.Model(&repo).Update("status", "ready")
}