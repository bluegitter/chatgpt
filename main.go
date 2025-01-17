package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/chzyer/readline"
)

type payload struct {
	Text string `json:"text"`
}

type response struct {
	EOF   bool   `json:"eof"`
	Error string `json:"error"`
	Text  string `json:"text"`
}

func doJson(client gpt3.Client, r io.Reader, w io.Writer) error {
	enc := json.NewEncoder(w)
	dec := json.NewDecoder(r)
	for {
		var p payload
		err := dec.Decode(&p)
		if err != nil {
			return err
		}
		err = client.CompletionStreamWithEngine(
			context.Background(),
			gpt3.TextDavinci003Engine,
			gpt3.CompletionRequest{
				Prompt: []string{
					p.Text,
				},
				MaxTokens:   gpt3.IntPtr(4000),
				Temperature: gpt3.Float32Ptr(0),
			}, func(resp *gpt3.CompletionResponse) {
				enc.Encode(response{EOF: false, Text: resp.Choices[0].Text})
			},
		)
		if err != nil {
			err = enc.Encode(response{Error: err.Error()})
			if err != nil {
				return err
			}
			continue
		}
		err = enc.Encode(response{EOF: true})
		if err != nil {
			return err
		}
	}
}

func main() {
	var j bool
	flag.BoolVar(&j, "json", false, "json input/output")
	flag.Parse()

	apiKey := os.Getenv("CHATGPT_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing CHATGPT_API KEY")
	}
	options := gpt3.WithTimeout(time.Duration(600 * time.Second))
	client := gpt3.NewClient(apiKey, options)

	if j {
		log.Fatal(doJson(client, os.Stdin, os.Stdout))
	}

	rl, err := readline.New("> ")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer rl.Close()

	for {
	    text, err1 := rl.Readline()
        if err1 != nil {
            fmt.Println("Error:", err1)
            break
        }

        // 如果输入为空或长度为零，则不做任何处理
        if len(text) == 0 {
            continue
        }

		// text := scanner.Text()
		err := client.CompletionStreamWithEngine(
			context.Background(),
			gpt3.TextDavinci003Engine,
			gpt3.CompletionRequest{
				Prompt: []string{
					text,
				},
				MaxTokens:   gpt3.IntPtr(4000),
				Temperature: gpt3.Float32Ptr(0.7),
			}, func(resp *gpt3.CompletionResponse) {
				fmt.Print(resp.Choices[0].Text)
			})
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println()
	}
}
