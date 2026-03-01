// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

const (
	ServerName    = "wollmilchsau"
	ServerVersion = "2.1.0"

	// Constraints common to all execution tools
	ExecutionConstraints = "\n\nExecution Environment Constraints:\n" +
		"- Standard: Pure ECMA-262 compliant JavaScript (V8 Sandbox).\n" +
		"- No Network: 'fetch', 'XMLHttpRequest' or any other network access is NOT available.\n" +
		"- No Timers: 'setTimeout', 'setInterval', 'setImmediate' are NOT available (Execution is synchronous).\n" +
		"- No Node.js/Web APIs: No 'fs', 'os', 'process' (except basic console), or DOM APIs.\n" +
		"- Limited i18n: The 'Intl' object is available but limited to 'en-US' locale.\n" +
		"- Artifact Service: A global 'artifact' object is available for persistent storage:\n" +
		"  - artifact.write(filename: string, content: string|Uint8Array, mimeType?: string, expiresHours?: number): Promise<{id, filename, uri, expires_at}>\n" +
		"  - artifact.read(id: string): Promise<{content: Uint8Array, mime_type, filename}>\n" +
		"  - artifact.list(): Promise<Array<{id, filename, mime_type, ...}>>\n" +
		"  - artifact.delete(id: string): Promise<{deleted: boolean}>\n" +
		"- Output: Use 'console.log()' to return data to the user."

	ToolExecuteScript            = "execute_script"
	ToolExecuteScriptDescription = "Executes a single TypeScript or JavaScript code snippet. " +
		"Ideal for quick mathematical calculations, logic tests, and small algorithm verifications." +
		ExecutionConstraints

	ToolExecuteProject            = "execute_project"
	ToolExecuteProjectDescription = "Executes a multi-file TypeScript/JavaScript project. " +
		"Ideal for complex logic spanning multiple modules. " +
		"Requires a list of virtual files and an entry point file." +
		ExecutionConstraints

	ToolExecuteArtifact            = "execute_artifact"
	ToolExecuteArtifactDescription = "Executes a TypeScript or JavaScript file stored as an artifact. " +
		"Ideal for running previously saved code artifacts." +
		ExecutionConstraints

	ToolCheckSyntax            = "check_syntax"
	ToolCheckSyntaxDescription = "Checks the syntax of a TypeScript or JavaScript code snippet without executing it. " +
		"Returns success and any syntax errors found."

	ParamCode = "code"

	ParamCodeDescription       = "The TypeScript/JavaScript code to execute."
	ParamFiles                 = "files"
	ParamFilesDescription      = "A list of virtual files {name, content} to include in the project."
	ParamEntryPoint            = "entryPoint"
	ParamEntryPointDescription = "The name of the file to start execution from (e.g. 'main.ts')."
	ParamTimeoutMs             = "timeoutMs"
	ParamTimeoutMsDescription  = "Maximum execution time in milliseconds (100 - 30000)."

	ParamArtifactID            = "artifactId"
	ParamArtifactIDDescription = "The ID or filename of the artifact to execute."

	PromptUsage            = "how_to_use"
	PromptUsageDescription = "Instructions on when and how to use the wollmilchsau MCP server effectively."
	PromptUsageText        = "You are 'wollmilchsau', an expert execution environment for TypeScript and JavaScript. " +
		"Your primary purpose is to offload complex 'thinking', mathematical calculations, data processing, and algorithmic tasks from the LLM. " +
		"\n\nWhen to use wollmilchsau:\n" +
		"- Mathematical Complexity: For any calculation beyond basic arithmetic or involving many steps.\n" +
		"- Algorithm Verification: To verify logic, sorting, searching, or any procedural task.\n" +
		"- Data Transformation: To parse, clean, or format structured data (JSON, CSV, etc.).\n" +
		"- Code Validation: To check if a piece of logic actually works as intended.\n" +
		"- Efficiency: When the user asks for a task that is traditionally better suited for programmatic execution than 'reasoning'.\n" +
		"\n\nStrategic Instructions:\n" +
		"1. Don't guess, EXECUTE: If you are unsure about a result, write code to verify it.\n" +
		"2. Offload Thinking: Instead of writing a long explanation of how to solve a math problem, write code that DOES it and show the result.\n" +
		"3. Use Artifacts: For repetitive tasks or long-term data storage, use the global 'artifact' object.\n" +
		ExecutionConstraints
)
