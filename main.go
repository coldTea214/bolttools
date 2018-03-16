package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/boltdb/bolt"
)

var (
	ErrUsage          = errors.New("usage")
	ErrUnknownCommand = errors.New("unknown command")

	ErrPathRequired   = errors.New("path required")
	ErrBucketRequired = errors.New("bucket required")
	ErrKeyRequired    = errors.New("key required")
	ErrValueRequired  = errors.New("value required")

	ErrFileNotFound   = errors.New("file not found")
	ErrBucketNotFound = errors.New("bucket not found")
)

func main() {
	m := NewMain()
	if err := m.Run(os.Args[1:]...); err == ErrUsage {
		os.Exit(2)
	} else if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// Main represents the main program execution.
type Main struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// NewMain returns a new instance of Main connect to the standard input/output.
func NewMain() *Main {
	return &Main{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Run executes the program.
func (m *Main) Run(args ...string) error {
	// Require a command at the beginning.
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fmt.Fprintln(m.Stderr, m.Usage())
		return ErrUsage
	}

	// Execute command.
	switch args[0] {
	case "help":
		fmt.Fprintln(m.Stderr, m.Usage())
		return ErrUsage
	case "buckets":
		return newBucketsCommand(m).Run(args[1:]...)
	case "list":
		return newListCommand(m).Run(args[1:]...)
	case "delete":
		return newDeleteCommand(m).Run(args[1:]...)
	case "insert":
		return newInsertCommand(m).Run(args[1:]...)
	default:
		return ErrUnknownCommand
	}
}

// Usage returns the help message.
func (m *Main) Usage() string {
	return strings.TrimLeft(`
BoltView is a tool for reading/writting bolt databases.

Usage:

    boltview command [arguments]

The commands are:

    buckets       list buckets in bolt database
    list          list key-value pairs in bucket
    insert        insert a key-value pair into bucket
    delete        delete a key-value pair from bucket

Use "bolt [command] -h" for more information about a command.
`, "\n")
}

type CommonCommand struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type BucketsCommand struct {
	CommonCommand
}

func newBucketsCommand(m *Main) *BucketsCommand {
	return &BucketsCommand{
		CommonCommand: CommonCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		},
	}
}

// Run executes the command.
func (cmd *BucketsCommand) Run(args ...string) error {
	// Parse flags.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	help := fs.Bool("h", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	} else if *help {
		fmt.Fprintln(cmd.Stderr, cmd.Usage())
		return ErrUsage
	}

	// Require database path.
	path := fs.Arg(0)
	if path == "" {
		return ErrPathRequired
	} else if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// Open database.
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Write header.
	fmt.Fprintln(cmd.Stdout, "NAME     ITEMS")
	fmt.Fprintln(cmd.Stdout, "======== ========")

	return db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			fmt.Fprintf(cmd.Stdout, "%-8s %-8d\n", string(name), bucket.Stats().KeyN)
			return nil
		})
	})
}

func (cmd *BucketsCommand) Usage() string {
	return strings.TrimLeft(`
usage: bolt buckets PATH

Buckets prints a table of buckets in bolt database
`, "\n")
}

type ListCommand struct {
	CommonCommand
}

func newListCommand(m *Main) *ListCommand {
	return &ListCommand{
		CommonCommand: CommonCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		},
	}
}

// Run executes the command.
func (cmd *ListCommand) Run(args ...string) error {
	// Parse flags.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	help := fs.Bool("h", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	} else if *help {
		fmt.Fprintln(cmd.Stderr, cmd.Usage())
		return ErrUsage
	}

	// Require database path.
	path := fs.Arg(0)
	if path == "" {
		return ErrPathRequired
	} else if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// Open database.
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	bucketName := fs.Arg(1)
	if bucketName == "" {
		return ErrBucketRequired
	}

	// Write header.
	fmt.Fprintln(cmd.Stdout, "KEY          VALUE")
	fmt.Fprintln(cmd.Stdout, "============ ============")

	return db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return ErrBucketNotFound
		}

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			if len(k) > 12 {
				k = k[0:12]
			}
			fmt.Fprintf(cmd.Stdout, "%-12s %-12s\n", string(k), string(v))
		}
		return nil
	})
}

func (cmd *ListCommand) Usage() string {
	return strings.TrimLeft(`
usage: bolt list PATH BUCKET_NAME

List prints a table of key-value pairs in that bucket
`, "\n")
}

type InsertCommand struct {
	CommonCommand
}

func newInsertCommand(m *Main) *InsertCommand {
	return &InsertCommand{
		CommonCommand: CommonCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		},
	}
}

// Run executes the command.
func (cmd *InsertCommand) Run(args ...string) error {
	// Parse flags.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	help := fs.Bool("h", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	} else if *help {
		fmt.Fprintln(cmd.Stderr, cmd.Usage())
		return ErrUsage
	}

	// Require database path.
	path := fs.Arg(0)
	if path == "" {
		return ErrPathRequired
	} else if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// Open database.
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	bucketName := fs.Arg(1)
	if bucketName == "" {
		return ErrBucketRequired
	}
	key := fs.Arg(2)
	if key == "" {
		return ErrKeyRequired
	}
	value := fs.Arg(3)
	if value == "" {
		return ErrValueRequired
	}

	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return ErrBucketNotFound
		}
		return bucket.Put([]byte(key), []byte(value))
	})
}

func (cmd *InsertCommand) Usage() string {
	return strings.TrimLeft(`
usage: bolt insert PATH BUCKET_NAME KEY VALUE

Insert add a pair of key-value into the bucket
`, "\n")
}

type DeleteCommand struct {
	CommonCommand
}

func newDeleteCommand(m *Main) *DeleteCommand {
	return &DeleteCommand{
		CommonCommand: CommonCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		},
	}
}

// Run executes the command.
func (cmd *DeleteCommand) Run(args ...string) error {
	// Parse flags.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	help := fs.Bool("h", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	} else if *help {
		fmt.Fprintln(cmd.Stderr, cmd.Usage())
		return ErrUsage
	}

	// Require database path.
	path := fs.Arg(0)
	if path == "" {
		return ErrPathRequired
	} else if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// Open database.
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	bucketName := fs.Arg(1)
	if bucketName == "" {
		return ErrBucketRequired
	}
	key := fs.Arg(2)
	if key == "" {
		return ErrKeyRequired
	}

	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return ErrBucketNotFound
		}
		return bucket.Delete([]byte(key))
	})
}

func (cmd *DeleteCommand) Usage() string {
	return strings.TrimLeft(`
usage: bolt delete PATH BUCKET_NAME KEY

Delete delete a pair of key-value from the bucket
`, "\n")
}
