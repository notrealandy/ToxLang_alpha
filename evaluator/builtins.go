package evaluator

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type BuiltinFunc func(args []interface{}) interface{}

var fileHandles = map[int]*os.File{}
var nextFileHandle = 1
var fileReaders = map[int]*bufio.Reader{}

var Builtins = map[string]BuiltinFunc{
	"go.println": func(args []interface{}) interface{} {
		fmt.Println(args...)
		return nil
	},
	"go.printf": func(args []interface{}) interface{} {
		if len(args) > 0 {
			format, ok := args[0].(string)
			if !ok {
				return nil
			}
			fmt.Printf(format, args[1:]...)
		}
		return nil
	},
	"go.time.now": func(args []interface{}) interface{} {
		return time.Now().Format(time.RFC3339)
	},
	"go.time.sleep": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if ms, ok := args[0].(int64); ok {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			}
		}
		return nil
	},
	"go.file.open": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if fname, ok := args[0].(string); ok {
				flags := os.O_RDWR | os.O_CREATE
				if len(args) > 1 {
					if mode, ok := args[1].(string); ok && mode == "append" {
						flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
					}
				}
				f, err := os.OpenFile(fname, flags, 0644)
				if err != nil {
					fmt.Println("Open error:", err)
					return nil
				}
				handle := nextFileHandle
				fileHandles[handle] = f
				fileReaders[handle] = bufio.NewReader(f) // <-- add this line
				nextFileHandle++
				return handle
			}
		}
		return nil
	},
	"go.file.close": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if handle, ok := args[0].(int); ok {
				if f, ok := fileHandles[handle]; ok {
					f.Close()
					delete(fileHandles, handle)
					delete(fileReaders, handle)
				}
			}
		}
		return nil
	},
	"go.file.read": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if handle, ok := args[0].(int); ok {
				if f, ok := fileHandles[handle]; ok {
					data, err := io.ReadAll(f)
					if err != nil {
						return nil
					}
					return string(data)
				}
			}
		}
		return nil
	},
	"go.file.write": func(args []interface{}) interface{} {
		if len(args) >= 2 {
			handle, ok1 := args[0].(int)
			data, ok2 := args[1].(string)
			if ok1 && ok2 {
				// Unescape escape sequences
				unescaped, err := strconv.Unquote(`"` + data + `"`)
				if err == nil {
					data = unescaped
				}
				if f, ok := fileHandles[handle]; ok {
					_, err := f.WriteString(data)
					return err == nil
				}
			}
		}
		return false
	},
	"go.file.create": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if fname, ok := args[0].(string); ok {
				f, err := os.Create(fname)
				if err != nil {
					fmt.Println("Create error:", err)
					return nil
				}
				handle := nextFileHandle
				fileHandles[handle] = f
				nextFileHandle++
				return handle
			}
		}
		return nil
	},
	"go.file.remove": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if fname, ok := args[0].(string); ok {
				err := os.Remove(fname)
				return err == nil
			}
		}
		return false
	},
	"go.dir.create": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if dirname, ok := args[0].(string); ok {
				err := os.Mkdir(dirname, 0755)
				return err == nil
			}
		}
		return false
	},
	"go.dir.remove": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if dirname, ok := args[0].(string); ok {
				err := os.Remove(dirname) // Only removes empty dirs
				return err == nil
			}
		}
		return false
	},
	"go.dir.removeAll": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if dirname, ok := args[0].(string); ok {
				err := os.RemoveAll(dirname)
				return err == nil
			}
		}
		return false
	},
	"go.path.exists": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if path, ok := args[0].(string); ok {
				_, err := os.Stat(path)
				return err == nil
			}
		}
		return false
	},
	"go.file.stat": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if fname, ok := args[0].(string); ok {
				info, err := os.Stat(fname)
				if err != nil {
					return nil
				}
				m := map[interface{}]interface{}{
					"name":           info.Name(),
					"size":           info.Size(),
					"mode":           info.Mode().String(),
					"modeBits":       uint32(info.Mode()),
					"modTime":        info.ModTime().Format(time.RFC3339),
					"modTimeRaw":     info.ModTime(), // if you want to expose the raw object
					"isDir":          info.IsDir(),
					"isRegular":      info.Mode().IsRegular(),
					"isSymlink":      info.Mode()&os.ModeSymlink != 0,
					"isHidden":       strings.HasPrefix(info.Name(), "."),
					"sys":            info.Sys(), // OS-specific, usually not needed
					"modeDevice":     info.Mode()&os.ModeDevice != 0,
					"modeCharDevice": info.Mode()&os.ModeCharDevice != 0,
					"modeNamedPipe":  info.Mode()&os.ModeNamedPipe != 0,
					"modeSocket":     info.Mode()&os.ModeSocket != 0,
					"modeSetuid":     info.Mode()&os.ModeSetuid != 0,
					"modeSetgid":     info.Mode()&os.ModeSetgid != 0,
					"modeSticky":     info.Mode()&os.ModeSticky != 0,
					"modeTemporary":  info.Mode()&os.ModeTemporary != 0,
					"modeAppend":     info.Mode()&os.ModeAppend != 0,
					"modeExclusive":  info.Mode()&os.ModeExclusive != 0,
					"modeIrregular":  info.Mode()&os.ModeIrregular != 0,
				}
				return m
			}
		}
		return nil
	},
	"go.file.readline": func(args []interface{}) interface{} {
		if len(args) > 0 {
			if handle, ok := args[0].(int); ok {
				if reader, ok := fileReaders[handle]; ok {
					line, err := reader.ReadString('\n')
					if err != nil && err != io.EOF {
						return nil
					}
					return strings.TrimRight(line, "\r\n")
				}
			}
		}
		return nil
	},
	"go.strings.split": func(args []interface{}) interface{} {
		if len(args) == 2 {
			s, ok1 := args[0].(string)
			sep, ok2 := args[1].(string)
			if ok1 && ok2 {
				parts := strings.Split(s, sep)
				result := make([]interface{}, len(parts))
				for i, p := range parts {
					result[i] = p
				}
				return result
			}
		}
		return nil
	},
	"go.strings.trim": func(args []interface{}) interface{} {
		if len(args) == 2 {
			s, ok1 := args[0].(string)
			cutset, ok2 := args[1].(string)
			if ok1 && ok2 {
				return strings.Trim(s, cutset)
			}
		}
		return nil
	},
	"go.strings.toLower": func(args []interface{}) interface{} {
		if len(args) == 1 {
			if s, ok := args[0].(string); ok {
				return strings.ToLower(s)
			}
		}
		return nil
	},
	"go.strings.toUpper": func(args []interface{}) interface{} {
		if len(args) == 1 {
			if s, ok := args[0].(string); ok {
				return strings.ToUpper(s)
			}
		}
		return nil
	},
	"go.bytes.make": func(args []interface{}) interface{} {
		if len(args) == 1 {
			if size, ok := args[0].(int64); ok && size >= 0 {
				buf := make([]byte, size)
				result := make([]interface{}, size)
				for i := range buf {
					result[i] = int64(buf[i])
				}
				return result
			}
		}
		return nil
	},
	"go.bytes.copy": func(args []interface{}) interface{} {
		if len(args) == 2 {
			dst, ok1 := args[0].([]interface{})
			src, ok2 := args[1].([]interface{})
			if ok1 && ok2 {
				n := copy(dst, src)
				return int64(n)
			}
		}
		return int64(0)
	},
	"go.bytes.cap": func(args []interface{}) interface{} {
		if len(args) == 1 {
			if arr, ok := args[0].([]interface{}); ok {
				return int64(cap(arr))
			}
		}
		return int64(0)
	},
}
