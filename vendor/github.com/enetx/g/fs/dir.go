package fs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/enetx/g"
)

// Dir is a struct representing a directory path.
type Dir struct {
	path g.String // Directory path.
}

// NewDir returns a new Dir instance with the given path.
func NewDir(path g.String) *Dir { return &Dir{path: path} }

// Chown changes the ownership of the directory to the specified UID and GID.
// It uses os.Chown to modify ownership and returns a Result[*Dir] indicating success or failure.
func (d *Dir) Chown(uid, gid int) g.Result[*Dir] {
	err := os.Chown(d.path.Std(), uid, gid)
	if err != nil {
		return g.Err[*Dir](err)
	}

	return g.Ok(d)
}

// Stat retrieves information about the directory represented by the Dir instance.
// It returns a Result[fs.FileInfo] containing details about the directory's metadata.
func (d *Dir) Stat() g.Result[fs.FileInfo] {
	path := d.Path()
	if path.IsErr() {
		return g.Err[fs.FileInfo](path.Err())
	}

	return g.ResultOf(os.Stat(path.Ok().Std()))
}

// Lstat retrieves information about the symbolic link represented by the Dir instance.
// It returns a Result[fs.FileInfo] containing details about the symbolic link's metadata.
// Unlike Stat, Lstat does not follow the link and provides information about the link itself.
func (d *Dir) Lstat() g.Result[fs.FileInfo] {
	path := d.Path()
	if path.IsErr() {
		return g.Err[fs.FileInfo](path.Err())
	}

	return g.ResultOf(os.Lstat(path.Ok().Std()))
}

// IsLink checks if the directory is a symbolic link.
func (d *Dir) IsLink() bool {
	stat := d.Lstat()
	return stat.IsOk() && stat.Ok().Mode()&os.ModeSymlink != 0
}

// CreateTempDir creates a new temporary directory in the specified directory with the
// specified name pattern and returns a Result, which contains a pointer to the Dir
// or an error if the operation fails.
// If no directory is specified, the default directory for temporary directories is used.
// If no name pattern is specified, the default pattern "*" is used.
//
// Parameters:
//
// - args ...String: A variadic parameter specifying the directory and/or name
// pattern for the temporary directory.
//
// Returns:
//
// - *Dir: A pointer to the Dir representing the temporary directory.
//
// Example usage:
//
// tmpdir := fs.CreateTempDir()                     // Creates a temporary directory with default settings
// tmpdirWithDir := fs.CreateTempDir("mydir")       // Creates a temporary directory in "mydir" directory
// tmpdirWithPattern := fs.CreateTempDir("", "tmp") // Creates a temporary directory with "tmp" pattern
func CreateTempDir(args ...g.String) g.Result[*Dir] {
	dir := ""
	pattern := "*"

	if len(args) != 0 {
		if len(args) > 1 {
			pattern = args[1].Std()
		}

		dir = args[0].Std()
	}

	tmpDir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return g.Err[*Dir](err)
	}

	return g.Ok(NewDir(g.String(tmpDir)))
}

// TempDir returns the default directory to use for temporary files.
//
// On Unix systems, it returns $TMPDIR if non-empty, else /tmp.
// On Windows, it uses GetTempPath, returning the first non-empty
// value from %TMP%, %TEMP%, %USERPROFILE%, or the Windows directory.
// On Plan 9, it returns /tmp.
//
// The directory is neither guaranteed to exist nor have accessible
// permissions.
func TempDir() *Dir { return NewDir(g.String(os.TempDir())) }

// Remove attempts to delete the directory and its contents.
// It returns a Result, which contains either the *Dir or an error.
// If the directory does not exist, Remove returns a successful Result with *Dir set.
// Any error that occurs during removal will be of type *PathError.
func (d *Dir) Remove() g.Result[*Dir] {
	if err := os.RemoveAll(d.String().Std()); err != nil {
		return g.Err[*Dir](err)
	}

	return g.Ok(d)
}

// Copy copies the contents of the current directory to the destination directory.
//
// Parameters:
//
// - dest (String): The destination directory where the contents of the current directory should be copied.
//
// - followLinks (optional): A boolean indicating whether to follow symbolic links during the walk.
// If true, symbolic links are followed; otherwise, they are skipped.
//
// Returns:
//
// - Result[*Dir]: A Result type containing either a pointer to a new Dir instance representing the destination directory or an error.
//
// Example usage:
//
//	sourceDir := fs.NewDir("path/to/source")
//	destinationDirResult := sourceDir.Copy("path/to/destination")
//	if destinationDirResult.IsErr() {
//		// Handle error
//	}
//	destinationDir := destinationDirResult.Ok()
func (d *Dir) Copy(dest g.String, followLinks ...bool) g.Result[*Dir] {
	files := g.NewSlice[*File]()

	for r := range d.Walk() {
		if r.IsErr() {
			return g.Err[*Dir](r.Err())
		}
		files.Push(r.Ok())
	}

	root := d.Path()
	if root.IsErr() {
		return g.Err[*Dir](root.Err())
	}

	destRoot := NewDir(dest).Path()
	if destRoot.IsErr() {
		return g.Err[*Dir](destRoot.Err())
	}

	follow := true
	if len(followLinks) > 0 {
		follow = followLinks[0]
	}

	for f := range files.Iter() {
		path := f.Path()
		if path.IsErr() {
			return g.Err[*Dir](path.Err())
		}

		relpath, err := filepath.Rel(root.Ok().Std(), path.Ok().Std())
		if err != nil {
			return g.Err[*Dir](err)
		}

		destpath := g.String(filepath.Join(destRoot.Ok().Std(), relpath))

		// Skip every symlink (file or directory) when not following links;
		// a symlink to a regular file is not an IsDir entry, so the skip must
		// happen before the IsDir branch to avoid dereferencing the link.
		if !follow && f.IsLink() {
			continue
		}

		stat := f.Stat()
		if stat.IsErr() {
			return g.Err[*Dir](stat.Err())
		}

		if stat.Ok().IsDir() {
			if r := NewDir(destpath).CreateAll(stat.Ok().Mode()); r.IsErr() {
				return r
			}

			continue
		}

		if r := f.Copy(destpath, stat.Ok().Mode()); r.IsErr() {
			return g.Err[*Dir](r.Err())
		}
	}

	return g.Ok(NewDir(dest))
}

// Create creates a new directory with the specified mode (optional).
//
// Parameters:
//
// - mode (os.FileMode, optional): The file mode for the new directory.
// If not provided, it defaults to DirDefault (0755).
//
// Returns:
//
// - *Dir: A pointer to the Dir instance on which the method was called.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory")
//	createdDir := dir.Create(0755) // Optional mode argument
func (d *Dir) Create(mode ...os.FileMode) g.Result[*Dir] {
	dmode := os.FileMode(g.DirDefault)
	if len(mode) > 0 {
		dmode = mode[0]
	}
	if err := os.Mkdir(d.path.Std(), dmode); err != nil {
		return g.Err[*Dir](err)
	}

	return g.Ok(d)
}

// Join joins the current directory path with the given path elements, returning the joined path.
//
// Parameters:
//
// - elem (...String): One or more String values representing path elements to
// be joined with the current directory path.
//
// Returns:
//
// - String: The resulting joined path as an String.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory")
//	joinedPath := dir.Join("subdir", "file.txt")
func (d *Dir) Join(elem ...g.String) g.Result[g.String] {
	path := d.Path()
	if path.IsErr() {
		return g.Err[g.String](path.Err())
	}

	parts := make([]string, len(elem)+1)
	parts[0] = path.Ok().Std()
	for i, part := range elem {
		parts[i+1] = part.Std()
	}

	return g.Ok(g.String(filepath.Join(parts...)))
}

// SetPath sets the path of the current directory.
//
// Parameters:
//
// - path (String): The new path to be set for the current directory.
//
// Returns:
//
// - *Dir: A pointer to the updated Dir instance with the new path.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory")
//	dir.SetPath("new/path/to/directory")
func (d *Dir) SetPath(path g.String) *Dir {
	d.path = path
	return d
}

// CreateAll creates all directories along the given path, with the specified mode (optional).
//
// Parameters:
//
// - mode ...os.FileMode (optional): The file mode to be used when creating the directories.
// If not provided, it defaults to the value of DirDefault constant (0755).
//
// Returns:
//
// - *Dir: A pointer to the Dir instance representing the created directories.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory")
//	dir.CreateAll()
//	dir.CreateAll(0755)
func (d *Dir) CreateAll(mode ...os.FileMode) g.Result[*Dir] {
	path := d.Path()
	if path.IsErr() {
		return g.Err[*Dir](path.Err())
	}

	dmode := os.FileMode(g.DirDefault)
	if len(mode) > 0 {
		dmode = mode[0]
	}

	err := os.MkdirAll(path.Ok().Std(), dmode)
	if err != nil {
		return g.Err[*Dir](err)
	}

	return g.Ok(d)
}

// Rename renames the current directory to the new path.
//
// Parameters:
//
// - newpath String: The new path for the directory.
//
// Returns:
//
// - Result[*Dir]: A Result containing a pointer to the Dir instance representing
// the renamed directory, or an error if the rename fails.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory")
//	dir.Rename("path/to/new_directory")
func (d *Dir) Rename(newpath g.String) g.Result[*Dir] {
	if rd := NewDir(g.String(filepath.Dir(filepath.Clean(newpath.Std())))).CreateAll(); rd.IsErr() {
		return rd
	}

	if err := os.Rename(d.path.Std(), newpath.Std()); err != nil {
		return g.Err[*Dir](err)
	}

	return g.Ok(NewDir(newpath))
}

// Path returns the absolute path of the current directory.
//
// Returns:
//
// - Result[String]: The absolute path of the current directory as a String,
// or an error if the path cannot be converted to an absolute path.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory")
//	absPath := dir.Path()
func (d *Dir) Path() g.Result[g.String] {
	path, err := filepath.Abs(d.path.Std())
	if err != nil {
		return g.Err[g.String](err)
	}

	return g.Ok(g.String(path))
}

// Exists checks if the current directory exists.
//
// Returns:
//
// - bool: true if the current directory exists, false otherwise.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory")
//	exists := dir.Exists()
func (d *Dir) Exists() bool {
	path := d.Path()
	if path.IsErr() {
		return false
	}

	info, err := os.Stat(path.Ok().Std())
	return err == nil && info.IsDir()
}

// Read lazily iterates over the content of the current directory and yields a
// File for each entry. Entries are read in bounded batches and retain the
// filesystem order; callers that need sorted output can collect and sort them.
//
// Returns:
//   - SeqResult[*File]: A sequence of Result[*File] instances representing each file and directory
//     in the current directory. It returns an error if reading the directory fails.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory")
//	files := dir.Read()
//	for file := range files {
//	    fmt.Println(file.Ok().Name())
//	}
func (d *Dir) Read() g.SeqResult[*File] {
	return func(yield func(g.Result[*File]) bool) {
		dpath := d.Path()
		if dpath.IsErr() {
			yield(g.Err[*File](dpath.Err()))
			return
		}

		directory, err := os.Open(dpath.Ok().Std())
		if err != nil {
			yield(g.Err[*File](err))
			return
		}

		defer directory.Close()

		const batchSize = 128
		base := dpath.Ok().Std()

		for {
			entries, readErr := directory.ReadDir(batchSize)
			for _, entry := range entries {
				if !yield(g.Ok(NewFile(filepath.Join(base, entry.Name())))) {
					return
				}
			}

			switch readErr {
			case nil:
				continue
			case io.EOF:
				return
			default:
				yield(g.Err[*File](readErr))
				return
			}
		}
	}
}

// Glob iterates over files in the current directory matching a specified pattern and yields File instances for each match.
// This method utilizes a lazy evaluation strategy, processing files as they are needed.
//
// Returns:
//   - SeqResult[*File]: A sequence of Result[*File] instances representing the files that match the
//     provided pattern in the current directory. It returns an error if the glob operation fails.
//
// Example usage:
//
//	dir := fs.NewDir("path/to/directory/*.txt")
//	files := dir.Glob()
//	for file := range files {
//	    fmt.Println(file.Ok().Name())
//	}
func (d *Dir) Glob() g.SeqResult[*File] {
	return func(yield func(g.Result[*File]) bool) {
		matches, err := filepath.Glob(d.path.Std())
		if err != nil {
			yield(g.Err[*File](err))
			return
		}

		for _, match := range matches {
			file := NewFile(g.String(match)).Path()
			if file.IsErr() {
				yield(g.Err[*File](file.Err()))
				return
			}

			if !yield(g.Ok(NewFile(file.Ok()))) {
				return
			}
		}
	}
}

// Walk returns a lazy sequence of all files and directories under the current Dir.
// You can customize inclusion/exclusion using SeqResult methods (Exclude, Filter, etc.).
//
// Example usage:
//
//	NewDir("path/to/dir").
//	  Walk().
//	  Exclude((*File).IsLink).
//	  ForEach(func(r Result[*File]) {
//	      if r.IsOk() {
//	          fmt.Println(r.Ok().Path().Ok().Std())
//	      }
//	  })
func (d *Dir) Walk() g.SeqResult[*File] {
	return func(yield func(g.Result[*File]) bool) {
		stack := g.SliceOf(d)
		stopped := false

		// Track resolved directory paths already scheduled for traversal so a
		// symlink (or hardlinked dir) pointing back into an ancestor does not
		// drive the stack into unbounded recursion.
		visited := g.NewSet[g.String]()
		if root := d.Path(); root.IsOk() {
			visited.Insert(root.Ok())
		}

		for !stack.IsEmpty() && !stopped {
			current := stack.Pop()
			if current.IsNone() {
				break
			}

			current.Some().Read().Range(func(r g.Result[*File]) bool {
				if r.IsErr() {
					if !yield(r) {
						stopped = true
						return false
					}
					return true
				}

				file := r.Ok()
				if !yield(g.Ok(file)) {
					stopped = true
					return false
				}

				// Use Lstat so that a symbolic link to a directory is reported
				// but not descended into; following it (via Stat) is what makes
				// symlink cycles loop forever.
				lstat := file.Lstat()
				if lstat.IsErr() {
					if !yield(g.Err[*File](lstat.Err())) {
						stopped = true
						return false
					}
					return true
				}

				if lstat.Ok().Mode()&os.ModeSymlink != 0 {
					return true
				}

				if lstat.Ok().IsDir() {
					path := file.Path()
					if path.IsErr() {
						if !yield(g.Err[*File](path.Err())) {
							stopped = true
							return false
						}
						return true
					}

					if visited.Contains(path.Ok()) {
						return true
					}

					visited.Insert(path.Ok())
					stack.Push(NewDir(path.Ok()))
				}

				return true
			})

			if stopped {
				return
			}
		}
	}
}

// String returns the String representation of the current directory's path.
func (d *Dir) String() g.String { return d.path }

// Print writes the content of the Dir to the standard output (console)
// and returns the Dir unchanged.
func (d *Dir) Print() *Dir { fmt.Print(d); return d }

// Println writes the content of the Dir to the standard output (console) with a newline
// and returns the Dir unchanged.
func (d *Dir) Println() *Dir { fmt.Println(d); return d }
