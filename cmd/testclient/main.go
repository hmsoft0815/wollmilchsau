// Copyright (c) 2026 Michael Lechner. All rights reserved.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/hmsoft0815/wollmilchsau/internal/server"
)

type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// das ist KI generierter Code für den test.. ich bin gespannt ob es funktioniert
func main() {
	fmt.Println("🚀 wollmilchsau Structured API Test Suite")

	// 1. Build
	_ = exec.Command("go", "build", "-o", "wollmilchsau", "./cmd/main.go").Run()
	defer os.Remove("wollmilchsau")

	// 2. Start
	cmd := exec.Command("./wollmilchsau")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill() }()

	reader := bufio.NewReader(stdout)
	id := 1

	runTest := func(name, tool string, args map[string]any) {
		fmt.Printf("\n--- TEST: %s (%s) ---\n", name, tool)
		sendRequest(stdin, "tools/call", map[string]any{"name": tool, "arguments": args}, id)
		id++

		line, _ := reader.ReadBytes('\n')
		var resp JSONRPCResponse
		_ = json.Unmarshal(line, &resp)

		if resp.Error != nil {
			fmt.Printf("❌ Error: %v\n", resp.Error)
			return
		}

		resultMap, _ := resp.Result.(map[string]any)
		contentList, _ := resultMap["content"].([]any)
		fmt.Printf("📦 Received %d content blocks\n", len(contentList))
		for _, c := range contentList {
			fmt.Printf("   ↳ %s\n", c.(map[string]any)["text"].(string))
		}
	}

	// Case 1: execute_script
	runTest("Simple Script", server.ToolExecuteScript, map[string]any{
		server.ParamCode: "console.log('Result:', Math.sqrt(16));",
	})

	// Case 2: execute_project
	runTest("Multi-file Project", server.ToolExecuteProject, map[string]any{
		server.ParamFiles: []map[string]any{
			{"name": "lib.ts", "content": "export const hello = () => 'Hello from Lib';"},
			{"name": "main.ts", "content": "import { hello } from './lib'; console.log(hello());"},
		},
		server.ParamEntryPoint: "main.ts",
	})

	// Case 2b: Complex Project (Subdirs + JSON)
	runTest("Complex Project (Subdirs/JSON)", server.ToolExecuteProject, map[string]any{
		server.ParamFiles: []map[string]any{
			{"name": "data/config.json", "content": `{ "appName": "Wollmilchsau Test", "version": 1 }`},
			{"name": "utils/math.ts", "content": "export const double = (n: number) => n * 2;"},
			{"name": "main.ts", "content": `
				import { double } from './utils/math';
				import config from './data/config.json';
				console.log('App:', config.appName);
				console.log('Calculation:', double(21));
			`},
		},
		server.ParamEntryPoint: "main.ts",
	})

	// Case 3: Timeout
	runTest("Timeout", server.ToolExecuteScript, map[string]any{
		server.ParamCode:      "while(true){}",
		server.ParamTimeoutMs: 500,
	})

	// Case 4: Out of Memory (OOM)
	runTest("Out of Memory", server.ToolExecuteScript, map[string]any{
		server.ParamCode:      "const list = []; for(let i=0; i<100000; i++) { list.push(new BigUint64Array(1024 * 1024)); }",
		server.ParamTimeoutMs: 30000,
	})

	// Case 5: i18n / Intl
	runTest("i18n / Intl", server.ToolExecuteScript, map[string]any{
		server.ParamCode: "const d = new Date(2026, 1, 26); console.log(new Intl.DateTimeFormat('en-US').format(d));",
	})

	// Case 6: Primes (First 40)
	runTest("Primes (First 40)", server.ToolExecuteScript, map[string]any{
		server.ParamCode: `
			function isPrime(n) {
				if (n < 2) return false;
				for (let i = 2; i <= Math.sqrt(n); i++) {
					if (n % i === 0) return false;
				}
				return true;
			}
			const primes = [];
			let num = 2;
			while (primes.length < 40) {
				if (isPrime(num)) primes.push(num);
				num++;
			}
			console.log(primes.join(', '));
		`,
	})

	// Case Polyfills: performance, crypto, atob/btoa, Buffer
	runTest("Polyfills Check", server.ToolExecuteScript, map[string]any{
		server.ParamCode: `
			console.log('Performance.now:', performance.now());
			
			const rand = new Uint8Array(4);
			crypto.getRandomValues(rand);
			console.log('Crypto Random:', Array.from(rand).join(', '));
			
			const b64 = btoa('Hello');
			console.log('Base64 Encode (btoa):', b64);
			console.log('Base64 Decode (atob):', atob(b64));
			
			const buf = Buffer.from('V29sbG1pbGNoc2F1', 'base64');
			console.log('Buffer from Base64:', new TextDecoder().decode(buf));
		`,
	})

	// Case 7: Syntax Check (Valid)
	runTest("Syntax Check (Valid)", server.ToolCheckSyntax, map[string]any{
		server.ParamCode: "const x: number = 42;",
	})

	// Case 8: Syntax Check (Invalid)
	runTest("Syntax Check (Invalid)", server.ToolCheckSyntax, map[string]any{
		server.ParamCode: "const x: = ;",
	})

	fmt.Println("\n🏁 All tests completed.")
}

func sendRequest(w io.Writer, method string, params any, id int) {
	req := JSONRPCRequest{JSONRPC: "2.0", ID: id, Method: method, Params: params}
	b, _ := json.Marshal(req)
	_, _ = w.Write(append(b, '\n'))
}
