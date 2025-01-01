
    Application Core Layer (Auth/Session Management)
Manages the overall flow and holds user credentials (phone, API ID, API Hash).
Responsible for session token storage, two-factor authentication (2FA), etc.

    User Input / CLI  
If needed, prompts user for verification code (when 2FA or sign-in is required).

    Peer Resolver      
Given a username or chat ID, resolves it to the actual internal peer

    Message Fetcher
Queries Telegram’s history for the target chat in batches (offset-based) or in real-time (updates).
Passes each message to the Media Extractor.

    Media Extractor                                          
Analyzes each message for attachments                     
Returns file info to File Downloader

    File Downloader
Downloads attached from Telegram servers;
provides local path/direct write
tgc Downloader.Download()

    Local Storage Mgr
Creates /tpm/*folders*, renames, saves final files,
