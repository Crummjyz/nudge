import * as vscode from 'vscode';

const warningRegex = /^::notice file=([^\n]+),line=(\d+),col=(\d+),endLine=(\d+),endColumn=(\d+)::(.*)$/;

export function activate(context: vscode.ExtensionContext) {
	let active = false;

	const diagnosticCollection = vscode.languages.createDiagnosticCollection('nudge');

	function checkWorkspace() {
		const workspaceFolders = vscode.workspace.workspaceFolders
		if (!workspaceFolders) {
			return
		}

		for (const workspaceFolder of workspaceFolders) {
			const workspacePath = workspaceFolder.uri.fsPath;

			const process = require('child_process')
			process.exec(`nudge ${workspacePath} --format=github`, (err: string, stdout: string, stderr: string) => {
				if (err) {
					return
				}

				diagnosticCollection.clear();

				let fileDiagnostics: Map<string, vscode.Diagnostic[]> = new Map();

				for (const warning of stdout.split("\n")) {
					const matches = warning.match(warningRegex);

					if (matches) {
						const filePath = matches[1];
						const message = matches[6];

						const lineNumber = parseInt(matches[2]);
						const columnNumber = parseInt(matches[3]);
						const endLineNumber = parseInt(matches[4]);

						for (let line = lineNumber; line <= endLineNumber; line++) {
							const diagnostic = {
								range: new vscode.Range(
									line - 1,
									columnNumber - 1,
									line,
									0,
								),
								message: message,
								severity: vscode.DiagnosticSeverity.Information,
								code: 'nudge'
							};

							let diagnostics = fileDiagnostics.get(filePath);
							if (diagnostics) {
								diagnostics.push(diagnostic);
							} else {
								fileDiagnostics.set(filePath, [diagnostic]);
							}
						}
					}
				}

				for (const [filePath, diagnostics] of fileDiagnostics.entries()) {
					diagnosticCollection.set(vscode.Uri.file(workspacePath + "/" + filePath), diagnostics);
				}
			});
		}
	}

	vscode.workspace.onDidSaveTextDocument((textDocument) => {
		if (active) {
			const languageId = textDocument.languageId;
			if (languageId === 'go' || languageId === 'rust' || languageId === 'c') {
				checkWorkspace();
			}
		}
	});

	let disposable = vscode.commands.registerCommand('nudge.checkWorkspace', () => {
		active = !active;
		if (active) {
			checkWorkspace();
		} else {
			diagnosticCollection.clear();
		}
	});

	context.subscriptions.push(disposable);
}

export function deactivate() { }
