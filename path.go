package Unipath

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type PathType uint8
type FileType uint8
var FILE_END = []byte("0x00\n\t0x01")

type Path struct {
	PathName string  // whole path
	File     os.FileInfo // path info
	Parent string // if PathName if file, parent is parent directory else parent is empty
	Name string // if PathName is  file, Name is whole file name
	Suffix string // if PathName is file, Suffix is file name suffix
	Stem string // if PathName is file, Stem is file name prefix
}

func (p *Path)String()string{
	return fmt.Sprintf("path name:%s", p.PathName)
}

// open file. if open file error return nil
// read-only
func (p *Path)Open()*os.File{
	f, _ := os.Open(p.PathName)
	return f
}

// current directory create file and write reader content to p.PathName file
// if p.PathName is directory filename must not empty
func (p *Path)Write(reader io.Reader, filename...string)error{
	if p.File.IsDir(){ // create file on current directory
		if len(filename) == 0{
			return errors.New("p.PathName is directory filename must not empty")
		}
		f, err := ioutil.ReadAll(reader)
		if err != nil{
			return err
		}
		return ioutil.WriteFile(filepath.Join(p.PathName, filename[0]), f, 0777)
	}else{ // write
		f, err := ioutil.ReadAll(reader)
		if err != nil{
			return err
		}
		return ioutil.WriteFile(p.PathName, f, 0777)
	}
}

// read current file content
// disposable read over
func (p *Path)Read()(error, io.Reader){
	if p.File.IsDir(){
		return errors.New("this function only support file but current object is directory"), nil
	}
	content, err := ioutil.ReadFile(p.PathName)
	if err != nil{
		return err, nil
	}
	return nil, bytes.NewReader(content)
}

// New Path object
// if path not exist return error
func NewPathObj(path string)(err error, p *Path){
	if path == ""{
		return errors.New("Param 'path' is empty"), nil
	}
	path, _ = filepath.Abs(path)
	p = &Path{PathName:path}
	info, err := os.Stat(path)
	if err != nil{
		return
	}
	p.File = info
	if !p.File.IsDir(){
		p.Parent = filepath.Dir(path)
		p.Name = filepath.Base(path)
		t := strings.SplitN(p.Name, ".", 2)
		if len(t) >= 2{
			p.Suffix = t[len(t) - 1]
		}
		p.Stem = t[0]
	}
	return
}

// contrast p and path is equal
// if p and path is directory, contrast any file or directory is equal
// if p and path is file, contrast file name if equal and file size is equal
func (p *Path)Equal(path *Path)bool{
	if !p.File.IsDir(){
		if p.PathName == path.PathName && p.File.Size() == path.File.Size() && p.File.ModTime().Equal(path.File.ModTime()){
			return true
		}
		return false
	}
	if p.File.Size() != path.File.Size(){
		return false
	}
	if p.File.ModTime().Equal(path.File.ModTime()){
		return false
	}
	err, paths := p.Walk()
	if err != nil{
		return false
	}
	err, paths1 := path.Walk()
	if err != nil{
		return false
	}
	var count int
	data1 := make(map[string]*Path, 0)
	data2 := make(map[string]*Path, 0)
	for count < 2{
		select {
		case p1 := <- paths1:
			if p1 == nil{
				count++
				break
			}
			data1[p1.PathName] = p1
		case p2 := <- paths:
			if p2 == nil{
				count++
				break
			}
			data2[p2.PathName] = p2
		}
	}
	for k, v := range data1{
		if v1, ok := data2[k]; !ok{
			return false
		}else{
			if v.PathName != v.PathName{
				return false
			}
			if v.File.Size() != v1.File.Size(){
				return false
			}
			if !v.File.ModTime().Equal(v1.File.ModTime()){
				return false
			}
		}
	}
	return true
}


// call filepath.Walk function
// return error and chan
func (p *Path)Walk()(error, chan *Path){
	if !p.File.IsDir(){
		return errors.New("this function only support directory"), nil
	}
	paths := make(chan *Path, 0)
	go filepath.Walk(p.PathName, func(path string, info os.FileInfo, err error) error {
		if err != nil{
			return nil
		}
		if info == nil{
			return nil
		}
		err, _p := NewPathObj(path)
		if err != nil{
			return nil
		}
		paths <- _p
		return nil
	})
	return nil, paths
}

// if p.PathName is empty return false else true
func (p *Path)isEmpty()(b bool){
	if p.PathName == ""{
		b = true
		return
	}
	return
}

// get p.PathName all file and folder
// only return first level directory
// if p.PathName is file return empty array and error == nil
func (p *Path)List()(err error, paths []*Path){
	if !p.File.IsDir(){
		err = errors.New("this function only support directory")
		return
	}
	paths = make([]*Path, 0)
	filepath.Walk(p.PathName, func(path string, info os.FileInfo, err error) error {


		if err != nil{
			return nil
		}
		if info == nil{
			return nil
		}
		if strings.Contains(info.Name(), "/"){ // second level dir not in paths
			return nil
		}
		if path == p.PathName{
			return nil
		}
		path, _ = filepath.Abs(path)
		_p := &Path{PathName:path, File:info}
		paths = append(paths, _p)
		return nil
	})
	return
}

// delete p.PathName
func (p * Path)Delete()(err error){
	return os.RemoveAll(p.PathName)
}

// move p.PathName to
// call os.rename function
// p.PathName reset to path
func (p *Path)Move(path string)(err error){
	path, _ = filepath.Abs(path)
	err = os.Rename(p.PathName, path)
	if err != nil{
		return err
	}
	p.PathName = path
	return err
}


// read big file
func (p *Path)ReadContent()(error, chan []byte ){
	if p.File.IsDir(){
		return errors.New("this function only suppprt p.PathName type is file "), nil
	}
	content := make(chan []byte, 0)
	go func() { // todo: is true?
		f, err := os.Open(p.PathName)
		if err != nil{
			return
		}
		r := bufio.NewReader(f)
		line := make([]byte, 0)
		for {
			fileByte, err := r.ReadByte()
			if err != nil{
				if len(line) != 0{
					content <- line
					break
				}
			}
			line = append(line, fileByte)
			if len(line) >= 10485760{
				content <- line
				line = make([]byte, 0)
			}
		}
		content <- FILE_END
	}()
	return nil, content
}

// copy p.PathName to path
// call system api
// path: this param copy directory
func (p *Path)Copy(path string)(err error, newPath *Path){
	if p.isEmpty(){
		err = errors.New("Path is empty")
		return
	}
	path, _ = filepath.Abs(path)
	_, err = os.Stat(path)
	if err != nil{ // make new directory
		err = os.MkdirAll(path, 0777)
		if err != nil{
			return
		}
	}

	if !p.File.IsDir(){
		err1, content := p.ReadContent()
		if err1 != nil{
			err = err1
			return
		}
		f1, err1 := os.OpenFile(filepath.Join(path, filepath.Base(p.PathName)), os.O_CREATE|os.O_RDWR, 0777)
		if err1 != nil{
			err = err1
			return
		}
		defer f1.Close()
		for {
			line := <- content
			if bytes.Equal(line, FILE_END){
				break
			}
			f1.Write(line)
		}
		err , newPath = NewPathObj(filepath.Join(path, filepath.Base(p.PathName)))
	}else{
		args := make([]string, 0)
		switch runtime.GOOS {
		case `windows`:
			args = append(args, "copy", p.PathName, path)
		default:
			args = append(args, "cp", "-r", p.PathName + "/", path + "/")
		}
		c := exec.Command(args[0], args[1:]...)
		msg, err1 := c.CombinedOutput()
		if err1 != nil{
			err = errors.New(string(msg))
			return
		}
		err , newPath = NewPathObj(path)
		return
	}

	return
}

// call filepath.Base
func (p *Path)Base()string{
	return filepath.Base(p.PathName)
}
