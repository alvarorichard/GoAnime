# Auto-Update Functionality

GoAnime now includes built-in auto-update functionality that allows users to
update to the latest version without manually downloading and installing.

## Features

- **Automatic Update Detection**: Checks GitHub releases for newer versions
- **Interactive Update Process**: Beautiful interactive prompts using Charm's
  `huh` library
- **Cross-Platform Support**: Works on Windows, macOS, and Linux
- **Safe Updates**: Creates backups and handles rollback if updates fail
- **Release Notes**: Option to view release notes before updating
- **Progress Indication**: Shows download progress and status

## Usage

### Check for Updates

```bash
goanime --update
```

This command will:

1. Check GitHub releases for the latest version
2. Compare with your current version
3. If an update is available, show an interactive menu with options:
   - **Yes, update now**: Downloads and installs the update
   - **No, maybe later**: Skips the update
   - **Show release notes**: Displays the changelog for the new version

### Interactive Update Menu

When an update is available, you'll see a beautiful menu like this:

```text
┃ Update Available!
┃ A new version of GoAnime is available.
┃
┃ Current: v1.0.0
┃ Latest:  v1.1.0 (4.0 MB)
┃
┃ Would you like to update now?
┃ > Yes, update now
┃   No, maybe later
┃   Show release notes
```

## How It Works

### Update Detection

1. **Version Comparison**: Compares your current version with the latest GitHub release
2. **Semantic Versioning**: Properly handles version comparison (e.g., 1.0.0 vs 1.1.0)
3. **Platform Detection**: Automatically selects the correct binary for your platform

### Download Process

1. **Asset Selection**: Automatically finds the right binary for your OS/architecture
2. **Safe Download**: Downloads to a temporary location first
3. **Progress Feedback**: Shows download status and file size

### Installation Process

1. **Backup Creation**: Creates a backup of your current executable
2. **Safe Replacement**: Uses multiple strategies to replace the executable:
   - Atomic rename (preferred)
   - Copy and replace (fallback for cross-device issues)
   - Gradual replacement for files in use
3. **Permission Setting**: Ensures the new executable has correct permissions
4. **Cleanup**: Removes temporary files and old backups

### Platform-Specific Handling

#### Linux/macOS

- Handles "text file busy" errors when executable is in use
- Uses atomic operations when possible
- Falls back to copy+rename for cross-filesystem operations

#### Windows

- Handles file locking by renaming current executable first
- Proper cleanup of old executables
- Handles Windows-specific path issues

## Error Handling

The auto-updater includes comprehensive error handling:

- **Network Issues**: Graceful handling of connection problems
- **Permission Issues**: Clear error messages for permission problems
- **File System Issues**: Fallback strategies for various file system limitations
- **Corrupted Downloads**: Validation and retry mechanisms
- **Rollback Support**: Automatic restoration if update fails

## Manual Fallback

If automatic update fails, the system provides:

- Clear error messages explaining what went wrong
- Link to the GitHub releases page for manual download
- Instructions for manual installation

## Security Considerations

- **HTTPS Only**: All downloads use secure HTTPS connections
- **GitHub Releases**: Only downloads from official GitHub releases
- **File Validation**: Verifies download integrity where possible
- **Safe Defaults**: Conservative approach to file operations

## Examples

### Successful Update

```bash
$ goanime --update
INFO  GoAnime  : Checking for updates...
INFO  GoAnime  : Update available: v1.0.0 → v1.1.0
# Interactive menu appears
INFO  GoAnime  : Downloading GoAnime v1.1.0...
INFO  GoAnime  : Downloading update...
INFO  GoAnime  : Successfully updated to GoAnime v1.1.0!
INFO  GoAnime  : The update has been installed. Please restart GoAnime to use
                   the new version.
```

### No Updates Available

```bash
$ goanime --update
INFO  GoAnime  : Checking for updates...
INFO  GoAnime  : You're already running the latest version of GoAnime (v1.1.0)!
```

### Update Declined

```bash
$ goanime --update
INFO  GoAnime  : Checking for updates...
INFO  GoAnime  : Update available: v1.0.0 → v1.1.0
# User selects "No, maybe later"
INFO  GoAnime  : Update skipped. You can update later by running: goanime --update
```

## Integration with Help System

The auto-update feature is fully integrated with GoAnime's help system:

```bash
goanime --help
```

Shows the `--update` flag in the options section with a clear description.

## Future Enhancements

Potential future improvements:

- Automatic update checks on startup (optional)
- Update channels (stable, beta, nightly)
- Delta updates for smaller downloads
- Digital signature verification
- Update scheduling
