package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tagg "github.com/wtolson/go-taglib"
	gcfg "gopkg.in/gcfg.v1"
)

type Mp3Song struct {
	Filename string
	Artist   string
	Title    string
	Genre    string
	Size     int64
}

type Config struct {
	InputDir  string
	InputList string
	OutDir    string
}

type configFile struct {
	Muzfinder Config
}

var mp3List []Mp3Song
var songlist map[string][]string
var songFoundList []string

func GetMp3Data(filename string) (Mp3Song, error) {
	mp3File, err := tagg.Read(filename)
	if err != nil {
		fmt.Println("Open: unable to open file: ", err)
		return Mp3Song{}, err
	}
	defer mp3File.Close()

	return Mp3Song{
		Filename: filename,
		Artist:   mp3File.Artist(),
		Title:    mp3File.Title(),
		Genre:    mp3File.Genre(),
		Size:     0,
	}, nil
}

func DirWalk(path string, fi os.FileInfo, err error) error {
	//fmt.Println("walk path: ", path)
	if fi.IsDir() {
		//fmt.Println("Search in ", path)
		//fmt.Printf("Process %d files\n", len(mp3List))
		return nil
	}

	if filepath.Ext(path) != ".mp3" {
		return nil
	}

	data, err := GetMp3Data(path)
	if err == nil {
		data.Size = fi.Size()
		mp3List = append(mp3List, data)

		/// TODO: check with Artist, Title as partial match
		/// for this create unique list of Artists and Titles

		if elem, ok := songlist[data.Artist]; ok {
			//fmt.Printf("\nFound Artist: %s, Title: %s, looking for: %v\n", data.Artist, data.Title, elem)
			for _, v := range elem {
				if strings.Contains(strings.ToLower(data.Title), strings.ToLower(v)) {
					fmt.Printf("Found Title: %s\n", data.Title)
					// fmt.Printf("%s == %s\n",
					// 	strings.ToLower(data.Title), strings.ToLower(v))
					fmt.Printf("filepath: %s\n", path)
					fmt.Printf("size: %d bytes\n", data.Size)
					songFoundList = append(songFoundList, path)
				}
			}
		}
	}
	//fmt.Println("fi: ", fi)
	return nil
}

func mkdir(dir string) {
	f, err := os.Open(dir)
	if err == nil {
		// if already exist
		defer f.Close()
		return
	}
	err2 := os.Mkdir(dir, 0755)
	if err2 != nil {
		fmt.Errorf("Mkdir %s: %s", dir, err2)
	}
}

func loadConfig(cfgFile string) Config {
	var cfg configFile
	err := gcfg.ReadFileInto(&cfg, cfgFile)
	if err != nil {
		fmt.Errorf("Error reading config file: %s", err)
	}
	return cfg.Muzfinder
}

// readSongList reads songs with songlist, parse it as Artist - Title and save
func readSongList(cfg Config) int {
	songCnt := 0

	f, err := os.Open(cfg.InputList)
	if err != nil {
		fmt.Println("Error: unable to open file: ", err)
		os.Exit(3)
	}
	defer f.Close()

	//var fileList []string

	r := bufio.NewReader(f)
	s, err3 := r.ReadString('\n')
	i := 1

	// pattern to get form filename: "^\d+\. (.*) - (.*)\.mp3$"
	rxp, _ := regexp.Compile(`(.*) [-—]{1,2} (.*)`)
	//rxp, _ := regexp.Compile(`([0-9]+). (.*) [-—]{1,2} (.*)`)

	for err3 == nil {
		//fmt.Printf("read[%d]: %s", i, s)
		//fileList = append(fileList, s)
		//fmt.Println(rxp.FindString(s))
		f := rxp.FindStringSubmatch(s)

		// for k, v := range f {
		// 	fmt.Printf("%d. %s\n", k, v)
		// }
		// if len(f) > 0 {
		// 	fmt.Printf("%d. %s\n", 0, f[0])
		// }
		if len(f) == 3 {
			Artist := f[1]
			Title := f[2]
			//fmt.Printf("'%v'\n", Title)
			songlist[Artist] = append(songlist[Artist], Title)

			songCnt += 1
		} else {
			fmt.Println("couldn't parse:", s)
		}
		//fmt.Println()
		s, err3 = r.ReadString('\n')
		i++
	}
	return songCnt
}

func main() {
	songlist = make(map[string][]string)

	cfg := loadConfig("muzfinder.conf")

	// Check flags
	//
	inputdir := flag.String("inputdir", "", "Set input dir for searching files")
	inputlist := flag.String("inputlist", "", "Set song list for searching")
	outdir := flag.String("outdir", "", "Set out dir for copying/moving files")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s OPTIONS\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *inputdir == "" && cfg.InputDir == "" {
		fmt.Fprintf(os.Stderr, "Not set inputdir\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -inputdir dir\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *inputdir != "" {
		cfg.InputDir = *inputdir
	}

	if *inputlist == "" && cfg.InputList == "" {
		fmt.Fprintf(os.Stderr, "Not set inputlist\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -inputlist list\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *inputlist != "" {
		cfg.InputList = *inputlist
	}

	if *outdir == "" && cfg.OutDir == "" {
		fmt.Fprintf(os.Stderr, "Not set outdir\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -outdir dir\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *outdir == "" {
		*outdir = cfg.OutDir
	}

	// Read file with list
	//
	songCnt := readSongList(cfg)

	//fmt.Println(songlist)
	fmt.Printf("Load %d songs from list\n", songCnt)

	// Walk in dirs
	//
	dirs := strings.Split(cfg.InputDir, ",")
	for _, dir := range dirs {
		err := filepath.Walk(dir, DirWalk)
		if err != nil {
			fmt.Errorf("DirWalk error: %v", err)
		}
	}

	cntFoundSongs := len(songFoundList)

	fmt.Println("Result:")
	fmt.Println("-------")
	fmt.Printf("Read %d songs from directory %s\n", len(mp3List), cfg.InputDir)
	fmt.Printf("found %d songs:\n", cntFoundSongs)
	in := ""
	for k, filename := range songFoundList {
		i := k + 1
		fmt.Println()
		fmt.Printf("[%d] %s\n", i, filename)
		if in != "sa" && in != "ca" {
			fmt.Printf("Set %d of %d. choose action: (c)opy, (m)ove, (s)kip, (d)elete, (ca)copy all, (sa)skip all: ", i, cntFoundSongs)
			fmt.Scanln(&in)
		}
		switch in {
		case "c":
			mkdir(*outdir)
			_, file := filepath.Split(filename)
			newfilename := filepath.Join(*outdir, file)
			fmt.Printf("copy file %s to %s\n", filename, newfilename)
			cerr := CopyFile(filename, newfilename)
			if cerr != nil {
				fmt.Printf("Error copy file: %v\n", cerr)
			}
		case "ca":
			mkdir(*outdir)
			_, file := filepath.Split(filename)
			newfilename := filepath.Join(*outdir, file)
			fmt.Printf("copy file %s to %s\n", filename, newfilename)
			cerr := CopyFile(filename, newfilename)
			if cerr != nil {
				fmt.Printf("Error copy file: %v\n", cerr)
			}
		case "m":
			fmt.Printf("move file %s to %s\n", filename, *outdir)
			mkdir(*outdir)
			_, file := filepath.Split(filename)
			newfilename := filepath.Join(*outdir, file)
			err := os.Rename(filename, newfilename)
			if err != nil {
				fmt.Printf("Error remove file: %s\n", err)
			}
		case "s":
			fmt.Println("skip file")
		case "sa":
			fmt.Println("skip all files")
		case "d":
			fmt.Printf("delete file %s\n", filename)
			err := os.Remove(filename)
			if err != nil {
				fmt.Printf("Error delete file: %s\n", err)
			}
		default:
			fmt.Println(in)
		}
	}
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

/// TODO:
/// 1. parse flags (done)
/// -inputdir [dir1 dir2] - directories for searching files
/// -inputlist file - song list for looking for
/// -outdir - directory to copy/move files
/// -options = copy|remove, add trackid, change trackid as id file?

/// 2. get Artist - Title from songlist
/// regexp and see
/// see https://github.com/StefanSchroeder/Golang-Regex-Tutorial/blob/master/01-chapter1.markdown
/// check with Artist, Title with filename as partial match (todo)

/// 3. Ask user: copy, move, skip, delete with every found file (done)
/// 3.1. Add copy all, skip all, move all, delete all (todo)
/// 3.2. Write list unfound songs to later search or search in VK (todo)

/// 4. Add config file with arg options  (done)
/// read from config then rewrite from args
/// useful because music dir doesn't change often
/// support various inputdirs, * in songlist (todo)

/// 5. Add logger and some debug levels

/// 6. Support different song names w/ and w/out id
//
// home:
// /media/vova/data/music
// /media/vova/data/music/playlists
// /media/vova/data/music/playlist_future
// /media/vova/new1/music/latin music
// /media/vova/data/vkaudiosaver
// work:
// /media/disk/home/music/latin music
// /home/vova/Музыка
