// Package fs provides chainable, Result-based filesystem operations (File, Dir),
// including lazy SeqResult iterators over file lines and chunks.
package fs

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/enetx/g"
	"github.com/enetx/g/internal/filelock"
	"github.com/enetx/g/internal/mimesniff"
)

// errChunkSize is the shared error yielded by File.Chunks and File.ChunksRaw
// when the requested chunk size is not positive.
var errChunkSize = errors.New("chunk size must be > 0")

// File is a struct that represents a file.
type File struct {
	file  *os.File // Underlying os.File.
	name  g.String // File name.
	guard bool     // Guard indicates whether the file is protected against concurrent access.
}

type fileReader struct {
	owner *File
	file  *os.File
}

func (r fileReader) Read(p []byte) (int, error) { return r.file.Read(p) }
func (r fileReader) Close() error {
	if r.owner.file == r.file {
		return r.owner.Close()
	}

	return r.file.Close()
}

// NewFile returns a new File instance with the given name.
func NewFile[T ~string](name T) *File { return &File{name: g.String(name)} }

// Lines returns a new iterator instance that can be used to read the file
// line by line.
//
// Example usage:
//
//	// Open a new file with the specified name "text.txt"
//	fs.NewFile("text.txt").
//		Lines().                 // Read the file line by line
//		Skip(3).                 // Skip the first 3 lines
//		Exclude(f.IsZero).       // Exclude empty lines
//		Dedup().                 // Remove consecutive duplicate lines
//		Map(g.String.Upper).     // Convert each line to uppercase
//		ForEach(func(s g.Result[g.String]) { s.Ok().Print() }) // For each line, print it
//
//	// Output:
//	// UPPERCASED_LINE4
//	// UPPERCASED_LINE5
//	// UPPERCASED_LINE6
func (f *File) Lines() g.SeqResult[g.String] {
	return func(yield func(g.Result[g.String]) bool) {
		if f.file == nil {
			if r := f.Open(); r.IsErr() {
				yield(g.Err[g.String](r.Err()))
				return
			}
		}

		defer f.Close()

		scanner := bufio.NewScanner(f.file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			if !yield(g.Ok(g.String(scanner.Text()))) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield(g.Err[g.String](err))
			return
		}
	}
}

// LinesRaw returns a new iterator instance that reads the file line by line,
// yielding each line as a Bytes slice (raw []byte).
//
// This version avoids intermediate string allocations by working directly with byte slices.
// The returned Bytes are copies of the scanner buffer and are safe to retain.
//
// Returns:
//
// - SeqResult[Bytes]: An iterator over raw byte lines from the file.
//
// Example usage:
//
//	fs.NewFile("text.txt").
//		LinesRaw().                // Read raw byte lines
//		Filter(func(b g.Bytes) bool {
//			return len(b) > 0
//		}).
//		ForEach(func(line g.Result[g.Bytes]) {
//			line.Ok().Print()
//		})
//
// Output:
// LINE_1
// LINE_2
// ...
//
// Note: Each line is copied before yielding to avoid scanner buffer reuse issues.
func (f *File) LinesRaw() g.SeqResult[g.Bytes] {
	return func(yield func(g.Result[g.Bytes]) bool) {
		if f.file == nil {
			if r := f.Open(); r.IsErr() {
				yield(g.Err[g.Bytes](r.Err()))
				return
			}
		}

		defer f.Close()

		scanner := bufio.NewScanner(f.file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			line := make(g.Bytes, len(scanner.Bytes()))
			copy(line, scanner.Bytes())
			if !yield(g.Ok(line)) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield(g.Err[g.Bytes](err))
			return
		}
	}
}

// Chunks returns a new iterator instance that can be used to read the file
// in fixed-size chunks of the specified size in bytes.
//
// Parameters:
//
// - size (int): The size of each chunk in bytes.
//
// Example usage:
//
//	// Open a new file with the specified name "text.txt"
//	fs.NewFile("text.txt").
//		Chunks(100).              // Read the file in chunks of 100 bytes
//		Map(g.String.Upper).      // Convert each chunk to uppercase
//		ForEach(func(s g.Result[g.String]) { s.Ok().Print() }) // For each chunk, print it
//
//	// Output:
//	// UPPERCASED_CHUNK1
//	// UPPERCASED_CHUNK2
//	// UPPERCASED_CHUNK3
func (f *File) Chunks(size g.Int) g.SeqResult[g.String] {
	return func(yield func(g.Result[g.String]) bool) {
		if size.Lte(0) {
			yield(g.Err[g.String](errChunkSize))
			return
		}

		if f.file == nil {
			if r := f.Open(); r.IsErr() {
				yield(g.Err[g.String](r.Err()))
				return
			}
		}

		defer f.Close()

		buffer := make([]byte, size)

		for {
			n, err := f.file.Read(buffer)
			if err != nil && err != io.EOF {
				yield(g.Err[g.String](err))
				return
			}

			if n == 0 {
				break
			}

			if !yield(g.Ok(g.String(buffer[:n]))) {
				return
			}
		}
	}
}

// ChunksRaw returns a new iterator instance that reads the file in fixed-size
// chunks of bytes, yielding each chunk as a Bytes slice.
//
// This method avoids intermediate string allocations and operates directly on byte slices.
// Each chunk is copied from the underlying buffer to make it safe for downstream use.
//
// Parameters:
//
// - size (Int): The size of each chunk in bytes. Must be > 0.
//
// Returns:
//
// - SeqResult[Bytes]: An iterator over raw byte chunks from the file.
//
// Example usage:
//
//	fs.NewFile("text.txt").
//		ChunksRaw(128).            // Read raw 128-byte chunks
//		ForEach(func(chunk g.Result[g.Bytes]) {
//			chunk.Ok().Print()
//		})
//
// Output:
// RAW_CHUNK_1
// RAW_CHUNK_2
// ...
//
// Note: Each chunk is copied from the buffer to ensure memory safety.
func (f *File) ChunksRaw(size g.Int) g.SeqResult[g.Bytes] {
	return func(yield func(g.Result[g.Bytes]) bool) {
		if size.Lte(0) {
			yield(g.Err[g.Bytes](errChunkSize))
			return
		}

		if f.file == nil {
			if r := f.Open(); r.IsErr() {
				yield(g.Err[g.Bytes](r.Err()))
				return
			}
		}

		defer f.Close()

		buf := make([]byte, size)

		for {
			n, err := f.file.Read(buf)
			if err != nil && err != io.EOF {
				yield(g.Err[g.Bytes](err))
				return
			}

			if n == 0 {
				break
			}

			chunk := make(g.Bytes, n)
			copy(chunk, buf[:n])

			if !yield(g.Ok(chunk)) {
				return
			}
		}
	}
}

// Append appends the given content to the file, with the specified mode (optional).
// If no FileMode is provided, the default FileMode (0644) is used.
// Don't forget to close the file!
func (f *File) Append(content g.String, mode ...os.FileMode) g.Result[*File] {
	if f.file == nil {
		if r := f.createAll(); r.IsErr() {
			return r
		}

		fmode := os.FileMode(g.FileDefault)
		if len(mode) > 0 {
			fmode = mode[0]
		}

		if r := f.OpenFile(os.O_APPEND|os.O_CREATE|os.O_WRONLY, fmode); r.IsErr() {
			return r
		}
	}

	if _, err := f.file.WriteString(content.Std()); err != nil {
		return g.Err[*File](err)
	}

	return g.Ok(f)
}

// Chmod changes the mode of the file.
func (f *File) Chmod(mode os.FileMode) g.Result[*File] {
	var err error
	if f.file != nil {
		err = f.file.Chmod(mode)
	} else {
		err = os.Chmod(f.name.Std(), mode)
	}

	if err != nil {
		return g.Err[*File](err)
	}

	return g.Ok(f)
}

// Chown changes the owner of the file.
func (f *File) Chown(uid, gid int) g.Result[*File] {
	var err error
	if f.file != nil {
		err = f.file.Chown(uid, gid)
	} else {
		err = os.Chown(f.name.Std(), uid, gid)
	}

	if err != nil {
		return g.Err[*File](err)
	}

	return g.Ok(f)
}

// Seek sets the file offset for the next Read or Write operation. The offset
// is specified by the 'offset' parameter, and the 'whence' parameter determines
// the reference point for the offset.
//
// The 'offset' parameter specifies the new offset in bytes relative to the
// reference point determined by 'whence'. If 'whence' is set to io.SeekStart,
// io.SeekCurrent, or io.SeekEnd, the offset is relative to the start of the file,
// the current offset, or the end of the file, respectively.
//
// If the file is not open, this method will attempt to open it. If the open
// operation fails, an error is returned.
//
// If the Seek operation fails, the file is closed, and an error is returned.
//
// Example:
//
//	file := fs.NewFile("example.txt")
//	result := file.Seek(100, io.SeekStart)
//	if result.Err() != nil {
//	    log.Fatal(result.Err())
//	}
//
// Parameters:
//   - offset: The new offset in bytes.
//   - whence: The reference point for the offset (io.SeekStart, io.SeekCurrent, or io.SeekEnd).
//
// Don't forget to close the file!
func (f *File) Seek(offset int64, whence int) g.Result[*File] {
	if f.file == nil {
		if r := f.Open(); r.IsErr() {
			return r
		}
	}

	if _, err := f.file.Seek(offset, whence); err != nil {
		f.Close()
		return g.Err[*File](err)
	}

	return g.Ok(f)
}

// Close closes the File and unlocks its underlying file, if it is not already closed.
func (f *File) Close() error {
	if f.file == nil {
		return &ErrFileClosed{f.name.Std()}
	}

	var err error

	if f.guard {
		err = filelock.Unlock(f.file)
	}

	if closeErr := f.file.Close(); closeErr != nil {
		err = errors.Join(err, closeErr)
	}

	f.file = nil

	return err
}

// Copy copies the file to the specified destination, with the specified mode (optional).
// If no mode is provided, the default FileMode (0644) is used.
func (f *File) Copy(dest g.String, mode ...os.FileMode) g.Result[*File] {
	if r := f.Open(); r.IsErr() {
		return r
	}

	defer f.Close()

	nf := NewFile(dest)
	if f.guard {
		nf.guard = true
	}

	return nf.WriteFromReader(f.file, mode...)
}

// Create is similar to os.Create; if the file is guarded, the returned file is write-locked.
// Don't forget to close the file!
func (f *File) Create() g.Result[*File] {
	return f.OpenFile(os.O_RDWR|os.O_CREATE|os.O_TRUNC, g.FileCreate)
}

// Dir returns the directory the file is in as a Dir instance.
func (f *File) Dir() g.Result[*Dir] {
	dirPath := f.dirPath()
	if dirPath.IsErr() {
		return g.Err[*Dir](dirPath.Err())
	}

	return g.Ok(NewDir(dirPath.Ok()))
}

// Exists checks if the file exists.
func (f *File) Exists() bool {
	_, err := os.Stat(f.name.Std())
	return err == nil
}

// Ext returns the file extension.
func (f *File) Ext() g.String { return g.String(filepath.Ext(f.name.Std())) }

// Guard sets a lock on the file to protect it from concurrent access.
// It returns the File instance with the guard enabled.
func (f *File) Guard() *File {
	f.guard = true
	return f
}

// MimeType returns the MIME type of the file as Result[String].
func (f *File) MimeType() g.Result[g.String] {
	if r := f.Open(); r.IsErr() {
		return g.Err[g.String](r.Err())
	}

	defer f.Close()

	buff := make([]byte, mimesniff.SniffLen)

	bytesRead, err := f.file.ReadAt(buff, 0)
	if err != nil && err != io.EOF {
		return g.Err[g.String](err)
	}

	buff = buff[:bytesRead]

	return g.Ok(g.String(mimesniff.DetectContentType(buff)))
}

// Name returns the name of the file.
func (f *File) Name() g.String {
	if f.file != nil {
		return g.String(filepath.Base(f.file.Name()))
	}

	return g.String(filepath.Base(f.name.Std()))
}

// Open is like os.Open; if the file is guarded, the returned file is read-locked.
// Don't forget to close the file!
func (f *File) Open() g.Result[*File] { return f.OpenFile(os.O_RDONLY, 0) }

// OpenFile is like os.OpenFile; if the file is guarded, the returned file is locked.
// If flag includes os.O_WRONLY or os.O_RDWR, the file is write-locked;
// otherwise, it is read-locked.
// Don't forget to close the file!
func (f *File) OpenFile(flag int, perm fs.FileMode) g.Result[*File] {
	file, err := os.OpenFile(f.name.Std(), flag&^os.O_TRUNC, perm)
	if err != nil {
		return g.Err[*File](err)
	}

	if f.guard {
		switch flag & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR) {
		case os.O_WRONLY, os.O_RDWR:
			err = filelock.Lock(file)
		default:
			err = filelock.RLock(file)
		}

		if err != nil {
			file.Close()
			return g.Err[*File](err)
		}
	}

	if flag&os.O_TRUNC == os.O_TRUNC {
		if err := file.Truncate(0); err != nil {
			if fi, statErr := file.Stat(); statErr != nil || fi.Mode().IsRegular() {
				if f.guard {
					filelock.Unlock(file)
				}

				file.Close()

				return g.Err[*File](err)
			}
		}
	}

	// Release any descriptor already held by this File before reassigning,
	// otherwise re-opening (e.g. WriteFromReader on an open handle) leaks the
	// previous fd and orphans its advisory lock.
	if f.file != nil {
		if f.guard {
			filelock.Unlock(f.file)
		}

		f.file.Close()
	}

	f.file = file

	return g.Ok(f)
}

// Path returns the absolute path of the file.
func (f *File) Path() g.Result[g.String] { return f.filePath() }

// Print writes the content of the File to the standard output (console)
// and returns the File unchanged.
func (f *File) Print() *File { fmt.Print(f.Read().UnwrapOrDefault()); return f }

// Println writes the content of the File to the standard output (console) with a newline
// and returns the File unchanged.
func (f *File) Println() *File { fmt.Println(f.Read().UnwrapOrDefault()); return f }

// Read opens the named file (read-locked if the file is guarded) and returns its contents.
func (f *File) Read() g.Result[g.String] {
	if r := f.Open(); r.IsErr() {
		return g.Err[g.String](r.Err())
	}

	defer f.Close()

	content, err := io.ReadAll(f.file)
	if err != nil {
		return g.Err[g.String](err)
	}

	return g.Ok(g.String(content))
}

// Reader returns an io.ReadCloser for reading the file's contents.
// If the file is not already open, it attempts to open it automatically.
// The caller is responsible for closing the returned reader to release system resources.
func (f *File) Reader() g.Result[io.ReadCloser] {
	if f.file == nil {
		if r := f.Open(); r.IsErr() {
			return g.Err[io.ReadCloser](r.Err())
		}
	}

	return g.Ok[io.ReadCloser](fileReader{owner: f, file: f.file})
}

// Remove removes the file.
func (f *File) Remove() g.Result[*File] {
	if err := os.Remove(f.name.Std()); err != nil {
		return g.Err[*File](err)
	}

	return g.Ok(f)
}

// Rename renames the file to the specified new path.
func (f *File) Rename(newpath g.String) g.Result[*File] {
	if !f.Exists() {
		return g.Err[*File](&ErrFileNotExist{f.name.Std()})
	}

	nf := NewFile(newpath)
	if f.guard {
		nf.guard = true
	}

	if r := nf.createAll(); r.IsErr() {
		return r
	}

	if err := os.Rename(f.name.Std(), newpath.Std()); err != nil {
		return g.Err[*File](err)
	}

	return g.Ok(nf)
}

// Split splits the file path into its directory and file components.
func (f *File) Split() (*Dir, *File) {
	path := f.Path()
	if path.IsErr() {
		return nil, nil
	}

	dir, file := filepath.Split(path.Ok().Std())

	return NewDir(g.String(dir)), NewFile(g.String(file))
}

// Stat returns the fs.FileInfo of the file.
// It calls the file's Stat method if the file is open, or os.Stat otherwise.
func (f *File) Stat() g.Result[fs.FileInfo] {
	if f.file != nil {
		return g.ResultOf(f.file.Stat())
	}

	return g.ResultOf(os.Stat(f.name.Std()))
}

// Lstat retrieves information about the symbolic link represented by the *File instance.
// It returns a Result[fs.FileInfo] containing details about the symbolic link's metadata.
// Unlike Stat, Lstat does not follow the link and provides information about the link itself.
func (f *File) Lstat() g.Result[fs.FileInfo] {
	return g.ResultOf(os.Lstat(f.name.Std()))
}

// IsDir checks if the file is a directory.
func (f *File) IsDir() bool {
	stat := f.Stat()
	return stat.IsOk() && stat.Ok().IsDir()
}

// IsLink checks if the file is a symbolic link.
func (f *File) IsLink() bool {
	stat := f.Lstat()
	return stat.IsOk() && stat.Ok().Mode()&os.ModeSymlink != 0
}

// Std returns the underlying *os.File instance.
// Don't forget to close the file with Close!
func (f *File) Std() *os.File { return f.file }

// CreateTemp creates a new temporary file in the specified directory with the
// specified name pattern and returns a Result, which contains a pointer to the File
// or an error if the operation fails.
// If no directory is specified, the default directory for temporary files is used.
// If no name pattern is specified, the default pattern "*" is used.
//
// Parameters:
//
// - args ...String: A variadic parameter specifying the directory and/or name
// pattern for the temporary file.
//
// Returns:
//
// - *File: A pointer to the File representing the temporary file.
//
// Example usage:
//
//	tmpfile := fs.CreateTempFile()                     // Creates a temporary file with default settings
//	tmpfileWithDir := fs.CreateTempFile("mydir")       // Creates a temporary file in "mydir" directory
//	tmpfileWithPattern := fs.CreateTempFile("", "tmp") // Creates a temporary file with "tmp" pattern
//
// Call .Guard() on the result to lock it; no guard state is set implicitly.
func CreateTempFile(args ...g.String) g.Result[*File] {
	dir := ""
	pattern := "*"

	if len(args) != 0 {
		if len(args) > 1 {
			pattern = args[1].Std()
		}

		dir = args[0].Std()
	}

	tmpfile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return g.Err[*File](err)
	}

	ntmpfile := NewFile(g.String(tmpfile.Name()))
	ntmpfile.file = tmpfile

	defer ntmpfile.Close()

	return g.Ok(ntmpfile)
}

// Write opens the named file (creating it with the given permissions if needed)
// and overwrites it with the given content; if the file is guarded, it is
// write-locked while writing.
func (f *File) Write(content g.String, mode ...os.FileMode) g.Result[*File] {
	return f.WriteFromReader(content.Reader(), mode...)
}

// WriteFromReader takes an io.Reader (scr) as input and writes the data from the reader into the file.
// If no FileMode is provided, the default FileMode (0644) is used.
func (f *File) WriteFromReader(scr io.Reader, mode ...os.FileMode) g.Result[*File] {
	if f.file == nil {
		if r := f.createAll(); r.IsErr() {
			return r
		}
	}

	filePath := f.filePath()
	if filePath.IsErr() {
		return g.Err[*File](filePath.Err())
	}

	fmode := os.FileMode(g.FileDefault)
	if len(mode) > 0 {
		fmode = mode[0]
	}

	if r := f.OpenFile(os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fmode); r.IsErr() {
		return g.Err[*File](r.Err())
	}

	defer f.Close()

	_, err := io.Copy(f.file, scr)
	if err != nil {
		return g.Err[*File](err)
	}

	err = f.file.Sync()
	if err != nil {
		return g.Err[*File](err)
	}

	return g.Ok(f)
}

// dirPath returns the absolute path of the directory containing the file.
func (f *File) dirPath() g.Result[g.String] {
	name := f.name.Std()
	if info, err := os.Stat(name); err == nil && info.IsDir() {
		path, err := filepath.Abs(name)
		if err != nil {
			return g.Err[g.String](err)
		}

		return g.Ok(g.String(path))
	}

	path, err := filepath.Abs(filepath.Dir(name))
	if err != nil {
		return g.Err[g.String](err)
	}

	return g.Ok(g.String(path))
}

// filePath returns the full file path, including the directory and file name.
func (f *File) filePath() g.Result[g.String] {
	path, err := filepath.Abs(f.name.Std())
	if err != nil {
		return g.Err[g.String](err)
	}

	return g.Ok(g.String(path))
}

func (f *File) createAll() g.Result[*File] {
	dirPath := f.dirPath()
	if dirPath.IsErr() {
		return g.Err[*File](dirPath.Err())
	}

	if !f.Exists() {
		if err := os.MkdirAll(dirPath.Ok().Std(), g.DirDefault); err != nil {
			return g.Err[*File](err)
		}
	}

	return g.Ok(f)
}
