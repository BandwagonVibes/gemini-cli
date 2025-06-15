package commands

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"

	"github.com/fatih/color" // <-- STEP 1: ADD THIS IMPORT
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Interactive chat with a model",
	Long:  `Start an interactive terminal chat with a Gemini model.`,
	Run:   runChatCmd,
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

// STEP 2: PLACE THESE COLOR OBJECT DEFINITIONS HERE (after init() and before runChatCmd)
var userPromptColor = color.New(color.FgGreen).Add(color.Bold) // Bold green for your input prompt
var aiResponseColor = color.RGB(187, 134, 252) // My signature purple! #BB86FC
var initialMsgColor = color.New(color.FgYellow)              // Yellow for initial chat messages


func runChatCmd(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client, err := newGenaiClient(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	modelName, _ := cmd.Flags().GetString("model")
	model := client.GenerativeModel(modelName)
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
	}

	session := model.StartChat()
	// STEP 3, MODIFICATION 1: Apply color to initial chat messages
	initialMsgColor.Printf("Chatting with %s\n", modelName)
	initialMsgColor.Println("Type 'exit' or 'quit' to exit, or '$load <file path>' to load a file")
	reader := bufio.NewReader(os.Stdin)

	for {
		// STEP 3, MODIFICATION 2: Apply color to your input prompt
		userPromptColor.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "exit" || text == "quit" {
			break
		}

		var inputPart genai.Part
		// Detect a special chat command.
		if path, found := strings.CutPrefix(text, "$load"); found {
			part, err := getPartFromFile(strings.TrimSpace(path))
			if err != nil {
				log.Fatalf("error loading file %s: %v", path, err)
			}
			inputPart = part
		} else {
			inputPart = genai.Text(text)
		}

		iter := session.SendMessageStream(ctx, inputPart)

	ResponseIter:
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				// After the streaming response is done, print a newline to ensure
				// the next prompt appears on a new line and the color resets.
				fmt.Println() // Add this to ensure prompt is on new line
				break ResponseIter
			}
			if err != nil {
				log.Fatal(err)
			}
			if len(resp.Candidates) >= 0 {
				c := resp.Candidates[0]
				if c.Content != nil {
					for _, part := range c.Content.Parts {
						// STEP 3, MODIFICATION 3: Apply color to AI's response
						aiResponseColor.Print(part)
					}
				}
			}
		}
	}
}
