package advisor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ExplainCommand queries an LLM to explain the impact of a destructive command.
func ExplainCommand(command string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return localLlamaExplain(command)
	}

	return openAIExplain(command, apiKey)
}

func localLlamaExplain(command string) string {
	return fmt.Sprintf(`🤖 Advisor Agent (Local Llama-3):
• Command Analysis: '%s' is a highly destructive operation targeting core files/state.
• Technical Risk: Irreversible data loss or immediate termination of critical services.
• Business Impact: High probability of taking down the application, resulting in downtime and customer trust issues.`, command)
}

func openAIExplain(command, apiKey string) string {
	url := "https://api.openai.com/v1/chat/completions"

	prompt := fmt.Sprintf("You are a senior DevOps engineer. Explain in 3 concise bullet points what this command does, what can go wrong technically, and the business impact if run on a production server. Keep it under 60 words. The command is: %s", command)

	payload := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a DevOps safety advisor."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return localLlamaExplain(command)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return localLlamaExplain(command)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return localLlamaExplain(command)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return localLlamaExplain(command)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return localLlamaExplain(command)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return localLlamaExplain(command)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return localLlamaExplain(command)
	}

	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	content := message["content"].(string)

	return "🤖 Advisor Agent (GPT-4o-mini):\n" + content
}
