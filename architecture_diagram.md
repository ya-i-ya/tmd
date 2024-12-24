                                         ┌────────────────────────────────────┐
                                         │           External World           │
                                         │                                    │
                                         │          ┌─────────────┐          │
                                         │          │ Telegram API │          │
                                         │          └───────┬─────┘          │
                                         │                  │                │
                                         │                  │                │
                                         └──────────────────┼────────────────┘
                                                            │
                                                            │
                                                   (MTProto/HTTPS)
                                                            │
                                                  ┌─────────┴────────┐
                                                  │     TDLib /       │
                                                  │ Telegram Client    │ 
                                                  │  Library Binding   │
                                                  └───────┬───────────┘
                                                          │
                                                          │
                                             ┌────────────┼─────────────┐
                                             │ Application Core Layer    │
                                             │                           │
                                             │   ┌─────────────────────┐ │
                                             │   │ Authentication/      │
                                             │   │ Session Manager      │
                                             │   └───────┬─────────────┘
                                             │           │
                                             │           │   Stores persistent
                                             │           │   auth sessions/tokens
                                             │           │
                                        ┌─────┴─────┐    │
                                        │ User Input │    │
                                        │ Interface  │    │
                                        └─────┬─────┘    │
                                              │           │
                          ┌────────────────────┘           │
                          │                                │
                          ▼                                ▼
                ┌────────────────┐               ┌──────────────────┐
                │ Peer Resolver  │               │ Message Fetcher   │
                │                │               │                  │
                │ Resolves chat  │<--------------┘  Iterates over    │
                │ usernames/IDs  │  Provides         conversation      │
                │ into internal  │  chat_id/peer    history           │
                │ peer objects   │                                   │
                └───────┬────────┘                                   │
                        │                                            │
                        │                                            │
                        │                                            ▼
                ┌────────┴────────┐                         ┌────────────────┐
                │ Media Extractor  │                         │ File Downloader│
                │                  │                         │                │
                │  Parses messages │                         │ Downloads files│
                │  for attachments │<------------------------│ (documents,    │
                │ (photos/docs/etc)│  Provides file info     │ photos, etc.)  │
                │                  │                         │ from Telegram  │
                └───────┬─────────┘                         │ servers         │
                        │                                    └───────┬────────┘
                        │                                            │
                        │                                            │ Saves to
                        │                                            │ filesystem
                        ▼                                            │
              ┌───────────────────────┐                             │
              │ Local Storage Manager │<-----------------------------┘
              │                       │
              │   Responsible for     │
              │   creating directories│
              │   & saving files      │
              │   locally (e.g.       │
              │   ./downloads/)       │
              └───────────────────────┘


Key Flows:
1. The app starts and sets up the Telegram client via TDLib bindings.
2. The Authentication/Session Manager ensures that the user is authorized:
   - If not authenticated, prompt user (via User Input Interface) for phone/code/password.
   - Store session info locally for subsequent runs.
3. After authentication, the Peer Resolver converts a known username/chat reference into an internal peer object.
4. The Message Fetcher requests conversation history in batches from Telegram.
5. For each message, the Media Extractor checks if it contains any downloadable files.
6. If files are found, the File Downloader retrieves them from Telegram, often requiring a file reference and associated metadata.
7. The Local Storage Manager saves these files to disk in an organized manner.
8. The process continues until all messages (or the requested range) have been processed.

This architecture separates concerns:
- Networking & Telegram API complexity: handled by TDLib or other Go client libraries.
- Authentication and session persistence: isolated in their own component.
- Peer resolving, message fetching, and file downloading: each in dedicated modules.
- User interaction (e.g. console input): centralized.
- File saving: handled by a dedicated component, making it easy to switch storage backends if desired.
