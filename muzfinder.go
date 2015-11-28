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

func GetMp3Data(filename string) (Mp3Song, error) {
	mp3File, err := id3.Open(filename)
	if err != nil {
		fmt.Println("Open: unable to open file: ", err)
		return Mp3Song{}, err
	}
	defer mp3File.Close()

	// fmt.Println("Filename:", filename)
	// fmt.Println("Artist:", mp3File.Artist())
	// fmt.Println("Title:", mp3File.Title())
	// fmt.Println("Genre:", mp3File.Genre())
	// fmt.Println("")

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
		//fmt.Println("it is dir")
		return nil
	}

	if filepath.Ext(path) != ".mp3" {
		return nil
	}

	data, err := GetMp3Data(path)
	data.Size = fi.Size()

	if err == nil {
		mp3List = append(mp3List, data)
	}

	//fmt.Println("fi: ", fi)
	return nil
}

func main() {
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

	rxp, _ := regexp.Compile(`(.*) [-—]{1,2} (.*)`)

	for err3 == nil {
		//fmt.Printf("read[%d]: %s", i, s)
		fileList = append(fileList, s)

		fmt.Println(rxp.FindString(s))

		f := rxp.FindStringSubmatch(s)

		for k, v := range f {
			fmt.Printf("%d. %s\n", k, v)
		}
		// if len(f) > 0 {
		// 	fmt.Printf("%d. %s\n", 0, f[0])
		// }

		fmt.Println()

		s, err3 = r.ReadString('\n')
		i++
	}

	//Scanner

	// Walk in dirs
	err2 := filepath.Walk(*inputdir, DirWalk)
	if err2 != nil {
		fmt.Errorf("Dir Walk error: %v", err2)
	}

	fmt.Printf("Result: add %d songs from directory %s\n", len(mp3List), *inputdir)
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

// /media/disk/home/music/latin music
// /home/vova/Музыка/future/__flash

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
