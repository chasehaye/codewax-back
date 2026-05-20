package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"io"

	"codewax/internal/models"
	"gorm.io/gorm"
)

const MaxChunks = 10

func EmbedQuery(query string) ([]float32, error) {
	client := &http.Client{}
	embeddings, _, err := doFetchEmbeddingBatch(client, []string{query})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

func searchChunks(db *gorm.DB, repoID uint, embedding []float32) ([]models.Chunk, error) {
	strs := make([]string, len(embedding))
	for i, v := range embedding {
		strs[i] = fmt.Sprintf("%f", v)
	}
	vectorStr := "[" + strings.Join(strs, ",") + "]"

	var chunks []models.Chunk
	err := db.Raw(`
		SELECT * FROM chunks
		WHERE repository_id = ?
		ORDER BY embedding <=> ?::vector
		LIMIT ?
	`, repoID, vectorStr, MaxChunks).Scan(&chunks).Error
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	return chunks, nil
}

func buildPrompt(chunks []models.Chunk, question string) string {
	var sb strings.Builder

	sb.WriteString("You are a helpful code assistant. Use the following code snippets from the repository to answer the user's question.\n\n")
	sb.WriteString("Relevant code:\n\n")

	for _, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("// %s (lines %d-%d)\n", chunk.FilePath, chunk.StartLine, chunk.EndLine))
		sb.WriteString(chunk.Content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Question: ")
	sb.WriteString(question)

	return sb.String()
}

func streamClaude(prompt string, onChunk func(string)) (string, error) {
	reqBody, err := json.Marshal(map[string]any{
		"model":      "claude-haiku-4-5-20251001",
		"max_tokens": 1024,
		"stream":     true,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("claude returned status %d: %s", resp.StatusCode, string(body))
	}

	var fullResponse strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			onChunk(event.Delta.Text)
			fullResponse.WriteString(event.Delta.Text)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading stream: %w", err)
	}

	return fullResponse.String(), nil
}

func ProcessMessage(db *gorm.DB, conversation models.Conversation, content string, onChunk func(string)) (string, error) {
	embedding, err := EmbedQuery(content)
	if err != nil {
		return "", fmt.Errorf("failed to embed query: %w", err)
	}

	if err := db.Preload("Repositories").First(&conversation, conversation.ID).Error; err != nil {
		return "", fmt.Errorf("failed to load conversation repositories: %w", err)
	}

	if len(conversation.Repositories) == 0 {
		return "", fmt.Errorf("no repositories attached to this conversation")
	}

	var allChunks []models.Chunk
	for _, repo := range conversation.Repositories {
		chunks, err := searchChunks(db, repo.ID, embedding)
		if err != nil {
			return "", fmt.Errorf("failed to search chunks for repo %d: %w", repo.ID, err)
		}
		allChunks = append(allChunks, chunks...)
	}

	prompt := buildPrompt(allChunks, content)

	fullResponse, err := streamClaude(prompt, onChunk)
	if err != nil {
		return "", fmt.Errorf("failed to stream claude response: %w", err)
	}

	return fullResponse, nil
}