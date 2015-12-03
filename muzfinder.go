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

	id3 "github.com/mikkyang/id3-go"
)

type Mp3Song struct {
	Filename string
	Artist   string
	Title    string
	Genre    string
	Size     int64
}

var mp3List []Mp3Song
var songlist map[string][]string
var songFoundList []string

func GetMp3Data(filename string) (Mp3Song, error) {
	mp3File, err := id3.Open(filename)
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
		fmt.Println("Search in ", path)
		fmt.Printf("Process %d files\n", len(mp3List))
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
			fmt.Println()
			fmt.Printf("found Artist: %s, found Title: %s, looking for: %v\n", data.Artist, data.Title, elem)
			for _, v := range elem {
				if strings.Contains(strings.ToLower(data.Title), strings.ToLower(v)) {
					fmt.Printf("found Title: %s\n", data.Title)

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
		// already exist
		defer f.Close()
		return
	}

	err2 := os.Mkdir(dir, 0755)
	if err2 != nil {
		fmt.Errorf("Mkdir %s: %s", dir, err2)
	}
}

func main() {
	songCnt := 0
	songlist = make(map[string][]string)

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

	if *inputdir == "" {
		fmt.Fprintf(os.Stderr, "Not set inputdir\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -inputdir dir\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	if *inputlist == "" {
		fmt.Fprintf(os.Stderr, "Not set inputlist\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -inputlist list\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	if *outdir == "" {
		fmt.Fprintf(os.Stderr, "Not set outdir\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -outdir dir\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	// Read file with list
	//
	f, err := os.Open(*inputlist)
	if err != nil {
		fmt.Println("Open: unable to open file: ", err)
		os.Exit(3)
	}
	defer f.Close()

	var fileList []string

	r := bufio.NewReader(f)
	s, err3 := r.ReadString('\n')
	i := 1

	// pattern to get form filename: "^\d+\. (.*) - (.*)\.mp3$"
	rxp, _ := regexp.Compile(`(.*) [-—]{1,2} (.*)`)

	for err3 == nil {
		//fmt.Printf("read[%d]: %s", i, s)
		fileList = append(fileList, s)
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

			if _, ok := songlist[Artist]; ok {
				songlist[Artist] = append(songlist[Artist], Title)
			} else {
				songlist[Artist] = make([]string, 0)
				songlist[Artist] = append(songlist[Artist], Title)
			}
			songCnt += 1
		} else {
			fmt.Println("couldn't parse:", s)
		}
		//fmt.Println()
		s, err3 = r.ReadString('\n')
		i++
	}

	//fmt.Println(songlist)
	fmt.Printf("Load %d songs from list\n", songCnt)

	// Walk in dirs
	//
	err2 := filepath.Walk(*inputdir, DirWalk)
	if err2 != nil {
		fmt.Errorf("Dir Walk error: %v", err2)
	}

	cntFoundSongs := len(songFoundList)

	fmt.Println("Result:")
	fmt.Println("-------")
	fmt.Printf("Read %d songs from directory %s\n", len(mp3List), *inputdir)
	fmt.Printf("found %d songs:\n", cntFoundSongs)
	for k, filename := range songFoundList {
		i := k + 1
		fmt.Printf("[%d] %s\n", i, filename)
		fmt.Println()
		fmt.Printf("Set %d of %d. choose action: (c)opy, (m)ove, (s)kip, (d)elete: ", i, cntFoundSongs)
		in := ""
		fmt.Scanln(&in)
		switch in {
		case "c":
			mkdir(*outdir)
			// TODO: make result filepath
			fmt.Println("copy file %s to %s\n", filename, *outdir)
		case "m":
			mkdir(*outdir)
			// os.Rename(filename, outdir + (filename - oldDirPrefix))
			fmt.Printf("move file %s to %s\n", filename, *outdir)
		case "s":
			fmt.Println("skip file")
		case "d":
			fmt.Println("delete file")
		default:
			fmt.Println(in)
		}
	}
}

func CopyFile(dst, src string) error {
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
/// 1. parse flags
/// -inputdir [dir1 dir2] - directories for searching files
/// -inputlist file - song list for looking for
/// -outdir - directory to copy/move files
/// -options = copy|remove, add trackid, change trackid as id file

/// 2. get Artist - Title from songlist
/// regexp and see
/// see https://github.com/StefanSchroeder/Golang-Regex-Tutorial/blob/master/01-chapter1.markdown
/// check with Artist, Title as partial match

/// 3. Ask user: copy, move, skip, delete with every found file
///    write list unfound songs
//
// home:
// /media/vova/data/music/playlists
// /media/vova/data/music/playlist_future
// work:
// /media/disk/home/music/latin music
// /home/vova/Музыка/future (on work, fall down)
