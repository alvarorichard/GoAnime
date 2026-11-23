// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// New is the recommended way to return a new yt-dlp command builder. Once all
// flags are set, you can call [Run] to invoke yt-dlp with the necessary args, or
// the independent execution method (e.g. [Version]).
func New() *Command {
	cmd := &Command{
		env:           make(map[string]string),
		flagConfig:    &FlagConfig{},
		cancelMaxWait: 1 * time.Second,
	}
	return cmd
}

type Command struct {
	mu                   sync.RWMutex
	executable           string
	directory            string
	env                  map[string]string
	flagConfig           *FlagConfig
	separateProcessGroup bool
	cancelMaxWait        time.Duration
	disableEnvVarInherit bool

	progress *progressHandler
	log      LogCallbackFunc
}

// Clone returns a copy of the command, with all flags, env vars, executable,
// working directory, etc copied over.
func (c *Command) Clone() *Command {
	c.mu.RLock()
	cc := &Command{
		executable:           c.executable,
		directory:            c.directory,
		env:                  make(map[string]string, len(c.env)),
		flagConfig:           c.flagConfig.Clone(),
		separateProcessGroup: c.separateProcessGroup,
		cancelMaxWait:        c.cancelMaxWait,
		disableEnvVarInherit: c.disableEnvVarInherit,
		progress:             c.progress,
		log:                  c.log,
	}
	maps.Copy(cc.env, c.env)
	c.mu.RUnlock()
	return cc
}

// GetFlagConfig returns a copy of the flag config.
func (c *Command) GetFlagConfig() *FlagConfig {
	c.mu.RLock()
	cc := c.flagConfig.Clone()
	c.mu.RUnlock()
	return cc
}

// SetFlagConfig sets the flag config for the command, overriding ALL previously
// set flags. If nil is provided, a new empty flag config will be used.
func (c *Command) SetFlagConfig(flagConfig *FlagConfig) *Command {
	if flagConfig == nil {
		flagConfig = &FlagConfig{}
	}
	c.mu.Lock()
	c.flagConfig = flagConfig.Clone()
	c.mu.Unlock()
	return c
}

// SetExecutable sets the executable path to yt-dlp for the command.
func (c *Command) SetExecutable(path string) *Command {
	c.mu.Lock()
	c.executable = path
	c.mu.Unlock()

	return c
}

// SetWorkDir sets the working directory for the command. Defaults to current working
// directory.
func (c *Command) SetWorkDir(path string) *Command {
	c.mu.Lock()
	c.directory = path
	c.mu.Unlock()

	return c
}

// SetEnvVar sets an environment variable for the command. If value is empty, it will
// be removed. If the key is "PATH", it will be merged with the parent process's PATH,
// (where explicit values provided here will take precedence). See also [SetEnvVarInherit].
func (c *Command) SetEnvVar(key, value string) *Command {
	c.mu.Lock()
	if value == "" {
		delete(c.env, key)
	} else {
		c.env[key] = value
	}
	c.mu.Unlock()

	return c
}

// SetSeparateProcessGroup sets whether yt-dlp should run in a separate
// process group. Download managers should enable this so Ctrl-C / signals
// aimed at the app do not also hit yt-dlp; cancel via [context.Context]
// instead.
//
// On context cancel, [os/exec.Cmd] sends os.Interrupt first (SIGINT on
// Unix), then os.Kill after [Command.SetCancelMaxWait]. yt-dlp handles
// SIGINT by stopping ffmpeg children. SIGKILL of yt-dlp alone does not
// kill ffmpeg; those children keep running. On Windows, Interrupt
// (CTRL_BREAK) only works with a new process group, so managers should
// enable this or cancel may TerminateProcess immediately and leak ffmpeg.
//
// Only supported on Windows and Unix-like systems.
func (c *Command) SetSeparateProcessGroup(value bool) *Command {
	c.mu.Lock()
	c.separateProcessGroup = value
	c.mu.Unlock()

	return c
}

// SetCancelMaxWait sets how long to wait after context cancel before
// yt-dlp is SIGKILL'd. It is assigned to [os/exec.Cmd.WaitDelay].
// Defaults to 1 second.
//
// A positive delay sends os.Interrupt first so yt-dlp can tear down
// ffmpeg children, then Kill. Zero WaitDelay Kill's immediately and
// ffmpeg typically keeps running. Download managers should keep a
// positive delay (and [Command.SetSeparateProcessGroup] on Windows; see
// that method).
func (c *Command) SetCancelMaxWait(value time.Duration) *Command {
	c.mu.Lock()
	c.cancelMaxWait = value
	c.mu.Unlock()
	return c
}

// SetEnvVarInherit sets whether the command should inherit environment variables
// from the parent process. If enabled, the command will inherit the parent process's
// environment variables, and any environment variables set with [SetEnvVar] will be
// merged with the parent process's environment variables. If disabled, the command
// will only use the environment variables set with [SetEnvVar] (with the exception
// of PATH). Explicitly set env vars with [SetEnvVar] will always take precedence
// over the parent process's environment variables.
func (c *Command) SetEnvVarInherit(enabled bool) *Command {
	c.mu.Lock()
	c.disableEnvVarInherit = !enabled
	c.mu.Unlock()
	return c
}

func (c *Command) hasJSONFlag() bool {
	if slices.Contains(c.flagConfig.VerbositySimulation.Print, "%(j)") && len(c.flagConfig.VerbositySimulation.Print) == 1 {
		return true
	}
	if v := c.flagConfig.VerbositySimulation.PrintJSON; v != nil && *v {
		return true
	}
	if v := c.flagConfig.VerbositySimulation.DumpJSON; v != nil && *v {
		return true
	}
	if v := c.flagConfig.VerbositySimulation.DumpSingleJSON; v != nil && *v {
		return true
	}
	return false
}

// toMap converts a slice of environment variables to a map. Handles Windows
// environment variables that start with '='.
func toMap(env []string) map[string]string {
	r := make(map[string]string, len(env))
	for _, e := range env {
		if runtime.GOOS == "windows" {
			// On Windows, env vars can start with "=".
			prefix := false
			if e != "" && e[0] == '=' {
				e = e[1:]
				prefix = true
			}

			key, value, ok := strings.Cut(e, "=")
			if prefix {
				key = "=" + key
			}
			if ok {
				r[key] = value
			}
			continue
		}

		if key, value, ok := strings.Cut(e, "="); ok {
			r[key] = value
		}
	}
	return r
}

// BuildCommand builds the command to be executed. args passed here are any additional
// arguments to be passed to yt-dlp (commonly URLs or similar). This should not be used
// directly unless you want to reference the arguments passed to yt-dlp.
func (c *Command) BuildCommand(ctx context.Context, args ...string) *exec.Cmd {
	var cmdArgs []string

	if bunResolveCache.Load() != nil {
		c.mu.Lock()
		if len(c.flagConfig.General.JsRuntimes) == 0 && c.flagConfig.General.NoJsRuntimes == nil {
			// Explicitly disable other options, and enable bun since they installed bun
			// through our install cache.
			c.flagConfig.General.NoJsRuntimes = new(true)
			c.flagConfig.General.JsRuntimes = append(c.flagConfig.General.JsRuntimes, "bun")
		}
		c.mu.Unlock()
	}

	c.mu.RLock()
	for _, f := range c.flagConfig.ToFlags() {
		cmdArgs = append(cmdArgs, f.Raw()...)
	}

	cmdArgs = append(cmdArgs, args...) // URLs or similar.

	var name string
	var err error

	name = c.executable

	if name == "" {
		if r := ytdlpResolveCache.Load(); r != nil {
			name = r.Executable
		} else {
			if config, _ := getBinaryConfig(ytdlpBinConfigs); config != nil {
				r, err = resolveExecutable(ctx, false, false, config.binaries)
				if err == nil {
					name = r.Executable
				}
			}
		}
	}

	env := map[string]string{}
	if !c.disableEnvVarInherit {
		env = toMap(os.Environ())
	}

	// Merge in the command's environment variables, accounting for things like
	// PATH, which should be merged with the previously provided PATH.
	for k, v := range c.env {
		switch k {
		case "PATH":
			if env["PATH"] != "" {
				paths := filepath.SplitList(env["PATH"])
				cpaths := filepath.SplitList(v)
				// Append parent process paths to the end of our custom provided
				// paths, only if they are not already in the PATH.
				seen := make(map[string]struct{}, len(cpaths)+len(paths))
				for _, p := range cpaths {
					seen[p] = struct{}{}
				}
				for _, p := range paths {
					if _, ok := seen[p]; !ok {
						seen[p] = struct{}{}
						cpaths = append(cpaths, p)
					}
				}
				env["PATH"] = strings.Join(cpaths, string(filepath.ListSeparator))
			} else {
				env["PATH"] = v
			}
		default:
			env[k] = v
		}
	}

	cmd := exec.CommandContext(ctx, name, cmdArgs...)

	// Interrupt first (see SetCancelMaxWait), then Kill so ffmpeg can die.
	cmd.WaitDelay = c.cancelMaxWait

	// Add cache directory to $PATH, which would cover ffmpeg, ffprobe, etc.
	cacheDir, err := GetCacheDir()
	if err == nil {
		env["PATH"] = strings.Join(append([]string{cacheDir}, filepath.SplitList(env["PATH"])...), string(filepath.ListSeparator))
	}

	if err != nil {
		cmd.Err = err // Hijack the existing command to return the error from resolveExecutable.
	}

	if c.directory != "" {
		cmd.Dir = c.directory
	}

	cmd.Env = make([]string, 0, len(env))
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	c.mu.RUnlock()

	return cmd
}

// runWithResult runs the provided command, collects stdout/stderr, massages the
// result into a Result struct, and returns it (with error wrapping).
func (c *Command) runWithResult(ctx context.Context, cmd *exec.Cmd) (*Result, error) {
	if cmd.Err != nil {
		return wrapError(nil, cmd.Err)
	}

	c.mu.RLock()
	progress := c.progress
	logFn := c.log
	separateProcessGroup := c.separateProcessGroup
	hasJSONFlag := c.hasJSONFlag()
	c.mu.RUnlock()

	var sink *logSink
	if logFn != nil {
		sink = &logSink{fn: logFn}
	}
	stdout := &timestampWriter{pipe: PipeStdout, checkJSON: hasJSONFlag, progress: progress, log: sink}
	stderr := &timestampWriter{pipe: PipeStderr, checkJSON: hasJSONFlag, log: sink}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	applySyscall(cmd, separateProcessGroup)

	debug(
		ctx, "running command",
		"path", cmd.Path,
		"args", cmd.Args,
		"env", cmd.Env,
		"dir", cmd.Dir,
	)

	err := cmd.Run()

	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	result := &Result{
		Executable: cmd.Path,
		Args:       cmd.Args[1:],
		ExitCode:   exitCode,
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		OutputLogs: stdout.mergeResults(stderr),
	}

	return wrapError(result, err)
}

// Run invokes yt-dlp with the provided arguments (and any flags previously set),
// and returns the results (stdout/stderr, exit code, etc). args should be the
// URLs that would normally be passed in to yt-dlp. To download from previously
// extracted info without re-hitting the site, use [Command.RunWithInfo].
func (c *Command) Run(ctx context.Context, args ...string) (*Result, error) {
	return c.run(ctx, nil, args...)
}

func (c *Command) run(ctx context.Context, stdin io.Reader, args ...string) (*Result, error) {
	c.mu.RLock()
	err := c.flagConfig.Validate()
	c.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	cmd := c.BuildCommand(ctx, args...)
	if stdin != nil {
		cmd.Stdin = stdin
	}
	return c.runWithResult(ctx, cmd)
}

// ExtractInfo runs yt-dlp with --skip-download --dump-json on a clone of
// c (the caller's flags are not modified) and returns parsed extractor
// JSON.
func (c *Command) ExtractInfo(ctx context.Context, urls ...string) ([]*ExtractedInfo, *Result, error) {
	result, err := c.Clone().SkipDownload().DumpJSON().Run(ctx, urls...)
	if err != nil {
		return nil, result, err
	}
	info, err := result.GetExtractedInfo()
	return info, result, err
}

// RunWithInfo writes info to yt-dlp via --load-info-json - (stdin) so the
// download does not re-extract from the site. c is not modified. A single
// item is written as that object; multiple items are wrapped as a playlist.
func (c *Command) RunWithInfo(ctx context.Context, info []*ExtractedInfo) (*Result, error) {
	var payload any
	switch len(info) {
	case 0:
		return nil, errors.New("no extracted info")
	case 1:
		payload = info[0]
	default:
		payload = &ExtractedInfo{Type: ExtractedTypePlaylist, Entries: info}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode load-info-json: %w", err)
	}

	cc := c.Clone().LoadInfoJSON("-")
	return cc.run(ctx, bytes.NewReader(data))
}

type Flag struct {
	ID             string `json:"id"`              // Unique ID to ensure boolean flags are not duplicated.
	Flag           string `json:"flag"`            // Actual flag, e.g. "--version".
	AllowsMultiple bool   `json:"allows_multiple"` // If the flag allows multiple values.
	Args           []any  `json:"args"`            // Optional args. If nil, it's a boolean flag.
}

func (f *Flag) Raw() (args []string) {
	args = make([]string, 1, len(f.Args)+1)
	args[0] = f.Flag
	if f.Args == nil {
		return args
	}

	for _, arg := range f.Args {
		if arg == nil {
			continue
		}

		switch arg := arg.(type) {
		case string:
			args = append(args, arg)
		case int:
			args = append(args, strconv.Itoa(arg))
		case int64:
			args = append(args, strconv.FormatInt(arg, 10))
		case float64:
			args = append(args, strconv.FormatFloat(arg, 'g', -1, 64))
		case bool:
			args = append(args, strconv.FormatBool(arg))
		default:
			panic(fmt.Sprintf("unsupported arg type for flag: %T", arg))
		}
	}

	return args
}

type Flags []*Flag

// Sort sorts the flags, ensuring some special flags are always ordered correctly.
func (f Flags) Sort() {
	slices.SortFunc(f, func(a, b *Flag) int {
		switch {
		case strings.HasPrefix(a.Flag, "--no") && !strings.HasPrefix(b.Flag, "--no"):
			return -1
		case strings.HasPrefix(b.Flag, "--no") && !strings.HasPrefix(a.Flag, "--no"):
			return 1
		default:
			return strings.Compare(a.ID, b.ID)
		}
	})
}

func (f Flags) FindByID(id string) (flags Flags) {
	for _, flag := range f {
		if flag.ID == id {
			flags = append(flags, flag)
		}
	}
	return flags
}

func (f Flags) Duplicates() (duplicates Flags) {
	seen := make(map[string]int, len(f))
	for _, flag := range f {
		if flag.AllowsMultiple {
			continue
		}
		seen[flag.ID]++
	}

	for _, flag := range f {
		if !flag.AllowsMultiple && seen[flag.ID] > 1 {
			duplicates = append(duplicates, flag)
		}
	}
	return duplicates
}
