package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"
	carscannertodb "github.com/rootofevil/carscannerlogs"
)

func main() {
	dir := "test"
	validate := ".*\\.csv"
	delim := ";"

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()
	done := make(chan bool)
	point := make(chan carscannertodb.CarData)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					if ValidateFile(path.Base(event.Name), validate) {
						time.Sleep(5000 * time.Millisecond)
						data, err := carscannertodb.ReadCsv(event.Name, delim)
						if err != nil {
							log.Println(err)
							return
						}
						for _, p := range data {
							point <- p
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case p := <-point:
				fmt.Printf("%+v", p)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(InitDir(dir, validate))
	<-done
}

func InitDir(inputdir, validate string) error {

	// ch := make(chan string)

	if _, err := os.Stat(inputdir); os.IsNotExist(err) {
		log.Printf("Check dir existance: %s\n", err)
		time.Sleep(5000 * time.Millisecond)
		return err
	}
	files, err := ioutil.ReadDir(inputdir)
	if err != nil {
		log.Printf("Read dir: %s\n", err)
		time.Sleep(5000 * time.Millisecond)
		return err
	}

	for _, f := range files {
		if !ValidateFile(f.Name(), validate) {
			log.Println("Wrong file:", f.Name())
			continue
		}
		log.Println("Processing file:", f.Name())
		inputfile := path.Join(inputdir, f.Name())
		fmt.Println(inputfile)
	}

	return nil
}

func ValidateFile(file, pattern string) bool {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	if !r.MatchString(file) {
		return false
	}
	return true
}
