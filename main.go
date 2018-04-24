package main // import "github.com/cam-a/storj-client"

import (
  "fmt"
  "io"
  "log"
  "net/http"
  "net/url"
  "os"
  "crypto/md5"

  "github.com/urfave/cli"
  "github.com/zeebo/errs"
)

var ArgError = errs.Class("argError")

func getAvailableFarmers() []string{
  farmers := []string {
          "http://127.0.0.1:8080", "http://127.0.0.1:8081", "http://127.0.0.1:8082",
          "http://127.0.0.1:8083", "http://127.0.0.1:8084", "http://127.0.0.1:8085",
          "http://127.0.0.1:8086", "http://127.0.0.1:8087", "http://127.0.0.1:8088",
          "http://127.0.0.1:8089",
        }

  return farmers
}

func determineHash(f *os.File, offset int64, length int64) (string, error){
	h := md5.New()
  f.Seek(offset, 0)
	if _, err := io.CopyN(h, f, length); err != nil {
	   return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func determineShardCount(size int64) (int64, int64, int64) {
  var minSize int64 = 1048576
  shardCount := size / minSize
  remainder := size % minSize
  if remainder > 0 {
    shardCount += 1
  }

  return shardCount, minSize, remainder
}

func main() {
  app := cli.NewApp()
  app.Name = "storj-client"
  app.Usage = ""
  app.Version = "1.0.0"

	app.Flags = []cli.Flag{}

	app.Commands = []cli.Command{
		{
			Name:      "upload",
			Aliases:   []string{"u"},
			Usage:     "Upload data",
			ArgsUsage: "[path]",
			Action: func(c *cli.Context) error {
				if c.Args().Get(0) == "" {
					return ArgError.New("No path provided")
				}

        file, err := os.Open(c.Args().Get(0))
        if err != nil {
          return err
        }
        defer file.Close()

        fileInfo, err := file.Stat()
        if err != nil {
          return err
        }

        hash, err := determineHash(file, 0, fileInfo.Size())
        if err != nil {
          return err
        }
        fmt.Println(hash)

        shardCount, avgShardSize, tailShardSize := determineShardCount(fileInfo.Size())
        fmt.Println(shardCount, avgShardSize, tailShardSize)

        farmers := getAvailableFarmers()

        for i := 1; int64(i) <= shardCount; i++ {
          var shardSize int64
          if int64(i) == shardCount && tailShardSize > 0 {
            shardSize = tailShardSize
          } else {
            shardSize = avgShardSize
          }

          farmer := farmers[0]
          farmers = farmers[1:]

          fmt.Println(i, shardSize, farmer)
        req, err := http.PostForm(farmer+"/upload", url.Values{hash: {hash}, offset: {"0"}, size: {shardSize}})
        }

				return nil
			},
		},
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
