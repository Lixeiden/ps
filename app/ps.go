package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// replaces all non-ASCII characters and path separators with "_"
func sanitizeFileName(name string) string {
	return strings.Map(func(r rune) rune {
		if r > 127 || r == '/' || r == '\\' {
			return '_'
		}
		return r
	}, name)
}

// periodically checks and removes expired files
func deleteExpiredFiles() {
	for {
		// Checks every minute
		time.Sleep(time.Minute)

		// Reads all files in the directory
		files, err := ioutil.ReadDir("./pst")
		// If an error occurs while reading the directory, log the error and continue the loop
		if err != nil {
			log.Printf("Could not read directory: %v", err)
			continue
		}
		// Iterates over each file
		for _, f := range files {
			// Checks if the file is a metadata file
			if strings.HasSuffix(f.Name(), ".meta") {
				metaData, err := ioutil.ReadFile("./pst/" + f.Name())
				// If an error occurs while reading the file, continue with the next file
				if err != nil {
					continue
				}

				// Splits the metadata into creation time and TTL
				parts := strings.Split(string(metaData), ",")
				if len(parts) != 2 {
					continue
				}

				// Tries to parse the creation time and TTL as integers
				createTime, err := strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					continue
				}

				TTL, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					continue
				}

				// If the file is expired, removes it and the corresponding metadata file
				if time.Now().Unix()-createTime > TTL {
					// Logs the files to be removed
					log.Printf("Removing expired paste: %s and %s", strings.TrimSuffix(f.Name(), ".meta")+".txt", f.Name())

					// Removes the paste and metadata files
					os.Remove("./pst/" + f.Name())
					os.Remove("./pst/" + strings.TrimSuffix(f.Name(), ".meta") + ".txt")
				}
			}
		}
	}
}

func main() {
	// Starts the file expiration check in a separate goroutine
	go deleteExpiredFiles()

	// Creates a new Gin server instance
	r := gin.Default()
	// Sets up a route for serving static files
	r.Static("/static", "./static")

	// Sets up a route for the main page
	r.GET("/", func(c *gin.Context) {
		// Serves the HTML file with the form
		http.ServeFile(c.Writer, c.Request, "./templates/index.html")
	})

	// Sets up a route for creating a new paste
	r.POST("/add", func(c *gin.Context) {
		// Retrieves the form data and sanitizes the file name
		name := sanitizeFileName(c.PostForm("name"))
		content := c.PostForm("content")
		TTL := c.PostForm("TTL")

		// If the file name is empty, returns an error
		if len(name) == 0 {
			log.Printf("File name cannot be empty")
			c.String(http.StatusBadRequest, "File name cannot be empty")
			return
		}

		// Tries to create the paste file and the metadata file
		if err := ioutil.WriteFile("./pst/"+name+".txt", []byte(content), 0644); err != nil {
			log.Printf("Could not write file: %v", err)
			if gin.Mode() == gin.DebugMode {
				c.String(http.StatusInternalServerError, "Could not write file: %v", err)
			} else {
				c.String(http.StatusInternalServerError, "Internal Server Error")
			}
			return
		}

		if err := ioutil.WriteFile("./pst/"+name+".meta", []byte(strconv.FormatInt(time.Now().Unix(), 10)+","+TTL), 0644); err != nil {
			log.Printf("Could not write metadata file: %v", err)
			if gin.Mode() == gin.DebugMode {
				c.String(http.StatusInternalServerError, "Could not write metadata file: %v", err)
			} else {
				c.String(http.StatusInternalServerError, "Internal Server Error")
			}
			return
		}

		// Redirects the client to the newly created paste
		c.Redirect(http.StatusSeeOther, "/v/"+name)
	})

	// Sets up a route for viewing a paste
	r.GET("/v/:name", func(c *gin.Context) {
		// Retrieves the file name from the URL parameters and sanitizes it
		name := sanitizeFileName(c.Param("name"))
		// Tries to read the paste file
		content, err := ioutil.ReadFile("./pst/" + name + ".txt")
		// If an error occurs, returns a "Not Found" error
		if err != nil {
			log.Printf("Could not find paste: %v", err)
			c.String(http.StatusNotFound, "Could not find paste: %v", err)
			return
		}

		// If the file is found, sends it as a response
        c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
	})

	// Listen on 0.0.0.0:8080 by default
	r.Run()
}
