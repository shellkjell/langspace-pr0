import * as path from 'path';
import { workspace, ExtensionContext } from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
    // The server is implemented in Go and distributed with the CLI
    // We run 'langspace lsp' to start the server
    const serverOptions: ServerOptions = {
        command: 'langspace',
        args: ['lsp'],
        options: { shell: true }
    };

    // Options to control the language client
    const clientOptions: LanguageClientOptions = {
        // Register the server for LangSpace files
        documentSelector: [{ scheme: 'file', language: 'langspace' }],
        synchronize: {
            // Notify the server about file changes to '.ls' files contained in the workspace
            fileEvents: workspace.createFileSystemWatcher('**/*.ls')
        }
    };

    // Create the language client and start the client.
    client = new LanguageClient(
        'langspaceLSP',
        'LangSpace Language Server',
        serverOptions,
        clientOptions
    );

    // Start the client. This will also launch the server
    client.start();
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
