package main

import (
	"BotMatrix/common/plugin/generator"
	"archive/zip"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type EventMessage struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	CorrelationId string `json:"correlation_id,omitempty"`
	Payload       any    `json:"payload"`
}

type Action struct {
	Type          string         `json:"type"`
	Target        string         `json:"target,omitempty"`
	TargetID      string         `json:"target_id,omitempty"`
	Text          string         `json:"text,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
}

type ResponseMessage struct {
	ID      string   `json:"id"`
	OK      bool     `json:"ok"`
	Actions []Action `json:"actions"`
}

type Intent struct {
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
	Regex    string   `json:"regex"`
}

type PluginManifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	EntryPoint  string   `json:"entry_point"`
	RunOn       []string `json:"run_on"`
	Permissions []string `json:"permissions"`
	Events      []string `json:"events"`
	Intents     []Intent `json:"intents"`
	MaxRestarts int      `json:"max_restarts"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "init":
		handleInit()
	case "pack":
		handlePack()
	case "debug":
		handleDebug()
	case "test":
		handleTest()
	case "gen":
		handleGen()
	case "run":
		handleRun()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("BotMatrix Plugin CLI (bm-cli)")
	fmt.Println("Usage:")
	fmt.Println("  bm-cli init <name> --lang <go|python|csharp>")
	fmt.Println("  bm-cli pack <dir> [--out <filename>]")
	fmt.Println("  bm-cli debug [dir]   Launch plugin in interactive debug mode")
	fmt.Println("  bm-cli test [dir]    Run automated tests defined in tests.json")
	fmt.Println("  bm-cli gen <prompt> [--lang go|python]  Generate a plugin using natural language")
	fmt.Println("  bm-cli run [dir]     Run plugin in a sandbox with a Web-based simulator")
	fmt.Println("  bm-cli help")
}

func handleRun() {
	dir := "."
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}

	manifestPath := filepath.Join(dir, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Printf("Error: Cannot find plugin.json in %s\n", dir)
		return
	}

	var m PluginManifest
	json.Unmarshal(data, &m)

	fmt.Printf("Starting Sandbox for %s...\n", m.Name)

	// 1. Start Plugin Process
	parts := strings.Fields(m.EntryPoint)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting plugin: %v\n", err)
		return
	}
	defer cmd.Process.Kill()

	// 2. Setup Bridge & Web Server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, simulatorHTML, m.Name)
	})

	// Simple WebSocket-like bridge using Server-Sent Events or Long Polling
	// For simplicity in a single-file Go CLI, we'll use a simple POST/GET queue
	inputChan := make(chan string, 10)
	outputChan := make(chan string, 10)

	// Plugin stdout -> outputChan
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			outputChan <- scanner.Text()
		}
	}()

	// inputChan -> Plugin stdin
	go func() {
		for msg := range inputChan {
			fmt.Fprintln(stdin, msg)
		}
	}()

	http.HandleFunc("/api/send", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		inputChan <- string(body)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/receive", func(w http.ResponseWriter, r *http.Request) {
		select {
		case msg := <-outputChan:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, msg)
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusNoContent)
		}
	})

	port := "8080"
	fmt.Printf("\n🚀 Simulator is running at http://localhost:%s\n", port)
	fmt.Println("Open this URL in your browser to interact with your plugin without any code!")

	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	cmd.Wait()
}

const simulatorHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>BotMatrix Plugin Simulator - %s</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 20px auto; background: #f4f4f9; }
        #chat { height: 400px; border: 1px solid #ccc; overflow-y: scroll; padding: 10px; background: white; margin-bottom: 10px; border-radius: 8px; }
        .msg { margin: 5px 0; padding: 8px; border-radius: 5px; max-width: 80%%; }
        .user { background: #e3f2fd; align-self: flex-end; margin-left: auto; }
        .bot { background: #f5f5f5; }
        .system { color: #888; font-size: 0.8em; text-align: center; }
        #input-area { display: flex; gap: 10px; }
        input { flex: 1; padding: 10px; border: 1px solid #ccc; border-radius: 4px; }
        button { padding: 10px 20px; background: #2196f3; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #1976d2; }
    </style>
</head>
<body>
    <h2>🤖 BotMatrix Simulator: %s</h2>
    <div id="chat">
        <div class="system">Simulator connected. Type 'ping' or any message to test.</div>
    </div>
    <div id="input-area">
        <input type="text" id="msgInput" placeholder="Type a message..." onkeypress="if(event.key==='Enter') send()">
        <button onclick="send()">Send</button>
    </div>

    <script>
        const chat = document.getElementById('chat');
        const input = document.getElementById('msgInput');

        function append(text, type) {
            const div = document.createElement('div');
            div.className = 'msg ' + type;
            div.innerText = text;
            chat.appendChild(div);
            chat.scrollTop = chat.scrollHeight;
        }

        async function send() {
            const text = input.value;
            if(!text) return;
            append(text, 'user');
            input.value = '';

            const event = {
                id: 'sim_' + Date.now(),
                type: 'event',
                name: 'on_message',
                payload: { from: 'Tester', text: text }
            };

            await fetch('/api/send', { method: 'POST', body: JSON.stringify(event) });
        }

        async function poll() {
            try {
                const resp = await fetch('/api/receive');
                if(resp.status === 200) {
                    const data = await resp.json();
                    if(data.actions) {
                        data.actions.forEach(a => {
                            if(a.type === 'send_text' || a.type === 'send_message') {
                                append(a.text || a.payload.text, 'bot');
                            } else {
                                append('[Action: ' + a.type + ']', 'system');
                            }
                        });
                    }
                }
            } catch(e) {}
            setTimeout(poll, 100);
        }
        poll();
    </script>
</body>
</html>
`

func handleGen() {
	genCmd := flag.NewFlagSet("gen", flag.ExitOnError)
	lang := genCmd.String("lang", "python", "Language for the generated plugin (go, python)")
	instantRun := genCmd.Bool("run", false, "Immediately run the plugin in simulator after generation")

	if len(os.Args) < 3 {
		fmt.Println("Error: Missing prompt")
		fmt.Println("Usage: bm-cli gen \"a plugin that echoes messages\" [--lang python] [--run]")
		return
	}

	// The prompt is the first argument after "gen", but flags can be anywhere
	// We'll join all non-flag arguments as the prompt
	var promptParts []string
	for i := 2; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			genCmd.Parse(os.Args[i:])
			break
		}
		promptParts = append(promptParts, os.Args[i])
	}
	prompt := strings.Join(promptParts, " ")

	apiKey := os.Getenv("BM_AI_KEY")
	baseURL := os.Getenv("BM_AI_URL")
	model := os.Getenv("BM_AI_MODEL")

	if apiKey == "" {
		fmt.Println("Error: BM_AI_KEY environment variable not set")
		return
	}
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	if model == "" {
		model = "deepseek-chat"
	}

	fmt.Printf("Generating %s plugin for: \"%s\"...\n", *lang, prompt)

	result, err := generator.GeneratePlugin(prompt, *lang)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	dir, err := generator.SavePlugin(result, ".")
	if err != nil {
		fmt.Printf("Error saving plugin: %v\n", err)
		return
	}

	fmt.Printf("\n✨ Plugin generated successfully in: %s\n", dir)
	fmt.Println("To get started:")
	fmt.Printf("  1. cd %s\n", dir)
	if *lang == "go" {
		fmt.Printf("  2. go mod init %s\n  3. go build -o %s . \n", result.Manifest["id"], result.Manifest["name"].(string)+".exe")
	} else if *lang == "python" {
		fmt.Println("  2. (Optional) Create a virtual environment: python -m venv venv")
		fmt.Println("  3. (Optional) Install dependencies: pip install botmatrix-sdk")
		fmt.Printf("  4. Run/Debug: bm-cli run .\n")
	}

	if *instantRun {
		fmt.Println("\n🚀 Instant Run requested. Starting simulator...")
		// We need to update os.Args to simulate "bm-cli run <dir>"
		os.Args = []string{"bm-cli", "run", dir}
		handleRun()
	}
}

type TestCase struct {
	Name   string         `json:"name"`
	Input  TestInput      `json:"input"`
	Expect []ActionExpect `json:"expect"`
}

type TestInput struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

type ActionExpect struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func handleTest() {
	dir := "."
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}

	// 1. Read manifest
	manifestPath := filepath.Join(dir, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Printf("Error: Cannot find plugin.json in %s\n", dir)
		return
	}
	var m PluginManifest
	json.Unmarshal(data, &m)

	// 2. Read test cases
	testPath := filepath.Join(dir, "tests.json")
	testData, err := os.ReadFile(testPath)
	if err != nil {
		fmt.Printf("Error: Cannot find tests.json in %s\n", dir)
		return
	}
	var cases []TestCase
	if err := json.Unmarshal(testData, &cases); err != nil {
		fmt.Printf("Error: Invalid tests.json: %v\n", err)
		return
	}

	fmt.Printf("Running %d test cases for %s...\n\n", len(cases), m.Name)

	passed := 0
	for i, tc := range cases {
		fmt.Printf("[%d/%d] Testing: %s... ", i+1, len(cases), tc.Name)

		success := runSingleTest(dir, m, tc)
		if success {
			fmt.Println("✅ PASS")
			passed++
		} else {
			fmt.Println("❌ FAIL")
		}
	}

	fmt.Printf("\nTest Summary: %d/%d passed\n", passed, len(cases))
	if passed < len(cases) {
		os.Exit(1)
	}
}

func runSingleTest(dir string, m PluginManifest, tc TestCase) bool {
	parts := strings.Fields(m.EntryPoint)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr // Pass through logs

	if err := cmd.Start(); err != nil {
		return false
	}
	defer cmd.Process.Kill()

	encoder := json.NewEncoder(stdin)
	decoder := json.NewDecoder(stdout)

	// Send input
	event := EventMessage{
		ID:      "test_ev",
		Type:    "event",
		Name:    tc.Input.Type,
		Payload: tc.Input.Payload,
	}
	encoder.Encode(event)

	// Read response with timeout
	respChan := make(chan ResponseMessage)
	errChan := make(chan error)

	go func() {
		var resp ResponseMessage
		if err := decoder.Decode(&resp); err != nil {
			errChan <- err
		} else {
			respChan <- resp
		}
	}()

	select {
	case resp := <-respChan:
		// Verify actions
		if len(resp.Actions) != len(tc.Expect) {
			return false
		}
		for i, expected := range tc.Expect {
			actual := resp.Actions[i]
			if actual.Type != expected.Type {
				return false
			}
			if expected.Text != "" && actual.Text != expected.Text {
				return false
			}
		}
		return true
	case <-errChan:
		return false
	case <-time.After(2 * time.Second):
		return false
	}
}

func handlePack() {
	packCmd := flag.NewFlagSet("pack", flag.ExitOnError)
	outFlag := packCmd.String("out", "", "Output filename (default: <id>_<version>.bmpk)")

	if len(os.Args) < 3 {
		fmt.Println("Error: Missing plugin directory")
		printUsage()
		return
	}

	dir := os.Args[2]
	packCmd.Parse(os.Args[3:])

	// 1. Read manifest
	manifestPath := filepath.Join(dir, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Printf("Error: Cannot find plugin.json in %s\n", dir)
		return
	}

	var m PluginManifest
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Printf("Error: Invalid plugin.json: %v\n", err)
		return
	}

	// 2. Determine output name
	outName := *outFlag
	if outName == "" {
		outName = fmt.Sprintf("%s_%s.bmpk", m.ID, m.Version)
	}
	if !strings.HasSuffix(outName, ".bmpk") {
		outName += ".bmpk"
	}

	fmt.Printf("Packing plugin '%s' (v%s) into %s...\n", m.Name, m.Version, outName)

	// 3. Create zip (bmpk)
	outFile, err := os.Create(outName)
	if err != nil {
		fmt.Printf("Error: Cannot create output file: %v\n", err)
		return
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Skip existing .bmpk files and hidden files
		if strings.HasSuffix(relPath, ".bmpk") || strings.HasPrefix(relPath, ".") {
			return nil
		}

		fmt.Printf("  + Adding %s\n", relPath)

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		w, err := zw.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, f)
		return err
	})

	if err != nil {
		fmt.Printf("Error during packing: %v\n", err)
		return
	}

	fmt.Printf("\nSuccess! Created %s\n", outName)
}

func handleInit() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	lang := initCmd.String("lang", "go", "Language for the plugin (go, python, csharp)")

	if len(os.Args) < 3 {
		fmt.Println("Error: Missing plugin name")
		printUsage()
		return
	}

	pluginName := os.Args[2]
	initCmd.Parse(os.Args[3:])

	fmt.Printf("Initializing plugin '%s' in %s...\n", pluginName, *lang)

	err := os.MkdirAll(pluginName, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	manifest := PluginManifest{
		ID:          fmt.Sprintf("com.botmatrix.%s", strings.ToLower(pluginName)),
		Name:        pluginName,
		Description: "A new BotMatrix plugin",
		Version:     "1.0.0",
		Author:      "Developer",
		RunOn:       []string{"worker"},
		Permissions: []string{"send_msg", "call_skill"},
		Events:      []string{"on_message"},
		Intents: []Intent{
			{
				Name:     "hello",
				Keywords: []string{"hello", "hi"},
			},
		},
		MaxRestarts: 5,
	}

	switch strings.ToLower(*lang) {
	case "go":
		manifest.EntryPoint = "./main"
		createGoTemplate(pluginName, manifest)
	case "python":
		manifest.EntryPoint = "python main.py"
		createPythonTemplate(pluginName, manifest)
	case "csharp":
		manifest.EntryPoint = "dotnet run"
		createCSharpTemplate(pluginName, manifest)
	default:
		fmt.Printf("Unsupported language: %s\n", *lang)
		return
	}

	// Write manifest
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	os.WriteFile(filepath.Join(pluginName, "plugin.json"), manifestData, 0644)

	fmt.Printf("\nSuccess! Plugin '%s' created in ./%s\n", pluginName, pluginName)
}

func handleDebug() {
	dir := "."
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}

	manifestPath := filepath.Join(dir, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Printf("Error: Cannot find plugin.json in %s\n", dir)
		return
	}

	var m PluginManifest
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Printf("Error: Invalid plugin.json: %v\n", err)
		return
	}

	fmt.Printf("Starting debug session for %s (%s)...\n", m.Name, m.ID)

	// Prepare command
	parts := strings.Fields(m.EntryPoint)
	if len(parts) == 0 {
		fmt.Println("Error: Empty entry_point")
		return
	}

	cmdName := parts[0]
	cmdArgs := parts[1:]

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = dir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Error: Failed to open stdin: %v\n", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error: Failed to open stdout: %v\n", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error: Failed to open stderr: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error: Failed to start plugin: %v\n", err)
		return
	}

	fmt.Println("Plugin started. Use 'msg <text>' to send a message, 'event <name> <json>' for custom events, or 'exit' to quit.")

	// Stdout reader
	go func() {
		decoder := json.NewDecoder(stdout)
		for {
			var resp ResponseMessage
			if err := decoder.Decode(&resp); err != nil {
				if err != io.EOF {
					// fmt.Printf("\n[Plugin STDOUT Error] %v\n", err)
				}
				return
			}
			fmt.Printf("\n[Plugin Response] ID: %s, OK: %v\n", resp.ID, resp.OK)
			for _, action := range resp.Actions {
				fmt.Printf("  - Action: %s, Text: %s, Payload: %v\n", action.Type, action.Text, action.Payload)
			}
			fmt.Print("> ")
		}
	}()

	// Stderr reader
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("\n[Plugin LOG] %s\n> ", scanner.Text())
		}
	}()

	// Input REPL
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		if input == "exit" {
			cmd.Process.Kill()
			break
		}

		if strings.HasPrefix(input, "msg ") {
			text := input[4:]
			event := EventMessage{
				ID:   fmt.Sprintf("dbg_%d", time.Now().UnixNano()),
				Type: "event",
				Name: "on_message",
				Payload: map[string]any{
					"from": "debug_user",
					"text": text,
				},
			}
			encoder.Encode(event)
			continue
		}

		if strings.HasPrefix(input, "event ") {
			parts := strings.SplitN(input[6:], " ", 2)
			if len(parts) < 2 {
				fmt.Println("Usage: event <name> <json_payload>")
				continue
			}
			var payload any
			if err := json.Unmarshal([]byte(parts[1]), &payload); err != nil {
				fmt.Printf("Invalid JSON: %v\n", err)
				continue
			}
			event := EventMessage{
				ID:      fmt.Sprintf("dbg_%d", time.Now().UnixNano()),
				Type:    "event",
				Name:    parts[0],
				Payload: payload,
			}
			encoder.Encode(event)
			continue
		}

		if input != "" {
			fmt.Println("Unknown command. Available: 'msg <text>', 'event <name> <payload>', 'exit'")
		}
	}

	cmd.Wait()
	fmt.Println("Debug session ended.")
}

func createGoTemplate(dir string, m PluginManifest) {
	content := `package main

import (
	"BotMatrix/sdk"
	"fmt"
	"log"
)

func main() {
	p := sdk.NewPlugin()

	// Handle standard messages
	p.OnMessage(func(ctx *sdk.Context) error {
		log.Printf("Received message from %s: %s", ctx.Event.Payload["from"], ctx.Event.Payload["text"])
		
		if ctx.Event.Payload["text"] == "ping" {
			ctx.Reply("pong!")
		}
		return nil
	})

	// Handle intents defined in plugin.json
	p.OnIntent("hello", func(ctx *sdk.Context) error {
		ctx.Reply(fmt.Sprintf("Hello! I am %s v%s", "` + m.Name + `", "` + m.Version + `"))
		return nil
	})

	// Example: Call a skill from another plugin
	// p.CallSkill("com.example.otherplugin", "some_skill", map[string]any{"param1": "value1"})

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
`
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(content), 0644)
	fmt.Println("- Created main.go")
	fmt.Println("- Please run 'go mod init " + m.ID + "' and add sdk dependency.")
}

func createPythonTemplate(dir string, m PluginManifest) {
	content := `import asyncio
from botmatrix import BotMatrixPlugin, Context

plugin = BotMatrixPlugin()

@plugin.on_message()
async def handle_message(ctx: Context):
    await ctx.reply(f"Hello from Python! You said: {ctx.text}")

if __name__ == "__main__":
    asyncio.run(plugin.run())
`
	os.WriteFile(filepath.Join(dir, "main.py"), []byte(content), 0644)
	fmt.Println("- Created main.py")
}

func createCSharpTemplate(dir string, m PluginManifest) {
	content := `using System;
using System.Threading.Tasks;
using BotMatrix.SDK;

class Program
{
    static async Task Main(string[] args)
    {
        var plugin = new Plugin();

        plugin.OnMessage(async (ctx) => {
            ctx.Reply($"Hello from C#! You said: {ctx.Event.Payload["text"]}");
        });

        await plugin.RunAsync();
    }
}
`
	os.WriteFile(filepath.Join(dir, "Program.cs"), []byte(content), 0644)
	fmt.Println("- Created Program.cs")
}
