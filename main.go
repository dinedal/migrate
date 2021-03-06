// Package main is the CLI.
// You can use the CLI via Terminal.
// import "github.com/mattes/migrate/migrate" for usage within Go.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dinedal/migrate/migrate"
	"github.com/fatih/color"
	"github.com/mattes/migrate/file"
	"github.com/mattes/migrate/migrate/direction"
	pipep "github.com/mattes/migrate/pipe"
)

var url = flag.String("url", "", "")
var migrationsPath = flag.String("path", "", "")
var version = flag.Bool("version", false, "Show migrate version")

func main() {
	flag.Parse()
	command := flag.Arg(0)

	if *version {
		fmt.Printf("%.2f", Version)
		os.Exit(0)
	}

	if *migrationsPath == "" {
		*migrationsPath, _ = os.Getwd()
	}

	switch command {
	case "create":
		verifyMigrationsPath(*migrationsPath)
		name := flag.Arg(1)
		if name == "" {
			fmt.Println("Please specify name.")
			os.Exit(1)
		}

		migrationFile, err := migrate.Create(*url, *migrationsPath, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Version %v migration files created in %v:\n", migrationFile.Version, migrationsPath)
		fmt.Println(migrationFile.UpFile.FileName)
		fmt.Println(migrationFile.DownFile.FileName)

	case "migrate":
		verifyMigrationsPath(*migrationsPath)
		relativeN := flag.Arg(1)
		relativeNInt, err := strconv.Atoi(relativeN)
		if err != nil {
			fmt.Println("Unable to parse parse param <n>.")
			os.Exit(1)
		}
		timerStart = time.Now()
		pipe := pipep.New()
		go migrate.Migrate(pipe, *url, *migrationsPath, relativeNInt)
		writePipe(pipe)
		printTimer()

	case "up":
		verifyMigrationsPath(*migrationsPath)
		timerStart = time.Now()
		pipe := pipep.New()
		go migrate.Up(pipe, *url, *migrationsPath)
		writePipe(pipe)
		printTimer()

	case "down":
		verifyMigrationsPath(*migrationsPath)
		timerStart = time.Now()
		pipe := pipep.New()
		go migrate.Down(pipe, *url, *migrationsPath)
		writePipe(pipe)
		printTimer()

	case "redo":
		verifyMigrationsPath(*migrationsPath)
		timerStart = time.Now()
		pipe := pipep.New()
		go migrate.Redo(pipe, *url, *migrationsPath)
		writePipe(pipe)
		printTimer()

	case "reset":
		verifyMigrationsPath(*migrationsPath)
		timerStart = time.Now()
		pipe := pipep.New()
		go migrate.Reset(pipe, *url, *migrationsPath)
		writePipe(pipe)
		printTimer()

	case "version":
		verifyMigrationsPath(*migrationsPath)
		version, err := migrate.Version(*url, *migrationsPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(version)

	default:
		fallthrough
	case "help":
		helpCmd()
	}
}

func writePipe(pipe chan interface{}) {
	if pipe != nil {
		for {
			select {
			case item, ok := <-pipe:
				if !ok {
					return
				} else {
					switch item.(type) {

					case string:
						fmt.Println(item.(string))

					case error:
						c := color.New(color.FgRed)
						c.Println(item.(error).Error(), "\n")

					case file.File:
						f := item.(file.File)
						c := color.New(color.FgBlue)
						if f.Direction == direction.Up {
							c.Print(">")
						} else if f.Direction == direction.Down {
							c.Print("<")
						}
						fmt.Printf(" %s\n", f.FileName)

					default:
						text := fmt.Sprint(item)
						fmt.Println(text)
					}
				}
			}
		}
	}
}

func verifyMigrationsPath(path string) {
	if path == "" {
		fmt.Println("Please specify path")
		os.Exit(1)
	}
}

var timerStart time.Time

func printTimer() {
	diff := time.Now().Sub(timerStart).Seconds()
	if diff > 60 {
		fmt.Printf("\n%.4f minutes\n", diff/60)
	} else {
		fmt.Printf("\n%.4f seconds\n", diff)
	}
}

func helpCmd() {
	os.Stderr.WriteString(
		`usage: migrate [-path=<path>] -url=<url> <command> [<args>]

Commands:
   create <name>  Create a new migration
   up             Apply all -up- migrations
   down           Apply all -down- migrations
   reset          Down followed by Up
   redo           Roll back most recent migration, then apply it again
   version        Show current migration version
   migrate <n>    Apply migrations -n|+n
   help           Show this help

'-path' defaults to current working directory.
`)
}
