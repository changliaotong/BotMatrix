package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type PluginManifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Entry       string   `json:"entry"`
	Actions     []string `json:"actions"`
	Events      []string `json:"events"`
	Intents     []any    `json:"intents"`
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
	fmt.Println("  bm-cli help")
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
		Actions:     []string{"send_message"},
		Events:      []string{"on_message"},
		Intents:     []any{},
	}

	switch strings.ToLower(*lang) {
	case "go":
		manifest.Entry = "./main"
		createGoTemplate(pluginName, manifest)
	case "python":
		manifest.Entry = "python main.py"
		createPythonTemplate(pluginName, manifest)
	case "csharp":
		manifest.Entry = "dotnet run"
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

func createGoTemplate(dir string, m PluginManifest) {
	content := `package main

import (
	"BotMatrix/sdk"
	"fmt"
)

func main() {
	p := sdk.NewPlugin()

	p.OnMessage(func(ctx *sdk.Context) error {
		ctx.Reply(fmt.Sprintf("Hello! You said: %s", ctx.Event.Payload["text"]))
		return nil
	})

	p.Run()
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
