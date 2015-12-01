package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

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
var songFoundCnt int

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
		fmt.Println(path)
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

		if elem, ok := songlist[data.Artist]; ok {
			fmt.Printf("found Artist: %s\n", data.Artist)
			for _, v := range elem {
				if v == data.Title {
					fmt.Printf("found Title: %s\n", data.Title)
					fmt.Printf("filepath: %s\n", path)
					songFoundCnt += 1
				}
			}
		}
	}
	//fmt.Println("fi: ", fi)
	return nil
}

func main() {
	songCnt := 0
	songlist = make(map[string][]string)

	// Check flags
	inputdir := flag.String("inputdir", "", "Set input dir for searching files")
	inputlist := flag.String("inputlist", "", "Set song list for searching")

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

	// Read file with list

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
				songlist[Artist] = make([]string, 1)
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

	//Scanner

	// Walk in dirs
	err2 := filepath.Walk(*inputdir, DirWalk)
	if err2 != nil {
		fmt.Errorf("Dir Walk error: %v", err2)
	}

	fmt.Printf("Read %d songs from directory %s\n", len(mp3List), *inputdir)
	fmt.Printf("Result: found %d songs\n", songFoundCnt)
}

/// TODO:
/// 1. parse flags
/// -inputdir [dir1 dir2] - directories for searching files
/// -inputlist file - song list for looking for
/// -outputdir - directory to copy/remove files
/// -options = copy|remove, add trackid, change trackid as id file

/// 2. get Artist - Title from songlist
/// regexp and see
/// see https://github.com/StefanSchroeder/Golang-Regex-Tutorial/blob/master/01-chapter1.markdown

/// 3. copy found songs somewhere
///    write list unfound songs
//
// home:
// /media/vova/data/music/playlists
// /media/vova/data/music/playlist_future
// work:
// /media/disk/home/music/latin music
// /home/vova/Музыка/future (on work, fall down)

// old code
/*
func readDir(dirname string) error {
	dir, err := os.Open(dirname)
	if err != nil {
		return err
	}
	files, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}

	fmt.Println("Dir: ", dirname)

	for _, filename := range files {
		// check that filename is not directory
		getMp3Data(dirname, filename)
		// if filepath.Ext(f) != ".article" {
		// 	continue
		// }
		// content, err := parseLesson(tmpl, filepath.Join(content, f))
		// if err != nil {
		// 	return fmt.Errorf("parsing %v: %v", f, err)
		// }
		// name := strings.TrimSuffix(f, ".article")
		// lessons[name] = content
		//fmt.Println("file: ", filename)
	}
	return nil
}
*/

// old code
/*
func getMp3Data(dirname string, filename string) {
	mp3File, err := id3.Open(filepath.Join(dirname, filename))
	if err != nil {
		fmt.Println("Open: unable to open file: ", err)
		return
	}
	defer mp3File.Close()

	fmt.Println("Filename:", filename)
	fmt.Println("Artist:", mp3File.Artist())
	fmt.Println("Title:", mp3File.Title())

	fmt.Println("Genre:", mp3File.Genre())
	fmt.Println("")
}
*/
