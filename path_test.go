package Unipath

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var (
	testDir, _ = filepath.Abs("./TEST_DIR")
	testFileSuffix = "test.test"
	testFile, _  = filepath.Abs(testDir + "/" + testFileSuffix)
	fileConent = []byte{116,104,105,115,32,105,115,32,116,101,115,116,32,102,105,108,101}
)


func TestMain(m *testing.M){
	// make test file and directory
	if err := os.Mkdir(testDir, 0777); err != nil{
		log.Panic(err)
	}
	f, err := os.OpenFile(testFile, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil{
		log.Panic(err)
	}
	if _, err := f.Write(fileConent); err != nil{
		log.Panic(err)

	}
	if err := f.Close(); err != nil{
		log.Panic(err)
	}
	f.Close()
	m.Run()
	if err := os.RemoveAll(testDir); err != nil{
		log.Panic(err)
	}
}

func TestPath_Copy(t *testing.T) {
	// test directory copy
	err, d := NewPathObj(testDir)
	if err != nil{
		t.Fatal(err)
	}
	if err, p := d.Copy("./test_copy"); err != nil{
		t.Fatal(err)
	}else{
		defer os.RemoveAll("./test_copy")
		if err , p := p.List(); err != nil{
			t.Fatal(err)
		}else if len(p) == 0{
			t.Fatal("copy directory error: not copy file, only copy directory")
		}else{
			if p[0].isEmpty(){
				t.Fatal("copy directory error: file name is empty")
			}
			c, err := ioutil.ReadFile(p[0].PathName)
			if err != nil{
				t.Fatal(err)
			}
			if !bytes.Equal(c, fileConent){
				t.Fatal("copy directory error: copy file content error")
			}
		}
	}
	err , p := NewPathObj(testFile)
	if err != nil{
		t.Fatal(err)
	}
	if err, p := p.Copy("./test_copy1"); err != nil{
		t.Fatal(err)
	}else{
		defer os.RemoveAll("./test_copy1")
		c, err := ioutil.ReadFile(p.PathName)
		if err != nil{
			t.Fatal(err)
		}
		if !bytes.Equal(c, fileConent){
			t.Fatal("copy directory error: copy file content error")
		}
	}
}

func TestNewPathObj(t *testing.T) {
	err, p := NewPathObj(testDir)
	if err != nil{
		t.Fatal(err)
	}
	if p.PathName != testDir{
		t.Fatal("p.PathName != testDir")
	}
	if p.File == nil{
		t.Fatal("p.File == nil")
	}
}

func TestPath_Delete_List(t *testing.T) {
	if err := os.Mkdir("./test_delete", 0777); err != nil{
		t.Fatal(err)
	}
	defer os.RemoveAll("./test_delete")
	if err := ioutil.WriteFile("./test_delete/test.test", []byte(""), 0777); err != nil{
		t.Fatal(err)
	}
	err, p := NewPathObj("./test_delete")
	if err != nil{
		t.Fatal(err)
	}
	err, files := p.List()
	if err != nil{
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("List function test error: p.list return file number error")
	}
	oldPath, _ := filepath.Abs("./test_delete/test.test")
	if files[0].PathName !=  oldPath{
		fmt.Println(files[0].PathName)
		t.Fatal("List  function test error: p.list return file name error")
	}
	if err := p.Delete(); err != nil{
		t.Fatal(err)
	}
	if _, err := os.Stat("./test_delete"); err == nil{
		t.Fatal("Delete function test error: p.Delete remove direction fail")
	}
}

func TestPath_Move(t *testing.T) {
	testMoveDir := "./test_move_dir"
	if err := os.Mkdir(testMoveDir, 0777); err != nil{
		t.Fatal(err)
	}
	defer os.RemoveAll(testMoveDir)
	if err := ioutil.WriteFile(testMoveDir + "/test.test", fileConent, 0777); err != nil{
		t.Fatal(err)
	}
	err, p := NewPathObj(testMoveDir)
	if err != nil{
		t.Fatal(err)
	}

	if err := p.Move("./test_to_move_dir"); err != nil{
		t.Fatal(err)
	}
	defer os.RemoveAll(p.PathName)
	path, _ := filepath.Abs("./test_to_move_dir")
	if p.PathName != path{
		t.Fatal("file move error: p.PathName error move after")
	}
}

func TestPath_ReadContent(t *testing.T) {
	_, p := NewPathObj(testFile)
	err, c := p.ReadContent()
	if err != nil{
		t.Fatal(err)
	}
	data := make([]byte, 0)
	for{
		content := <- c
		if len(content) == 0 || bytes.Equal(content, FILE_END){
			break
		}
		data = append(data, content...)
	}
	if !bytes.Equal(data, fileConent){
		t.Fatal("Readcontent function error: file content unequal read after")
	}
}

func TestPath_Base(t *testing.T) {
	_, p := NewPathObj(testFile)
	if p.Base() != testFileSuffix{
		t.Fatal("p.Base test error!")
	}
}

func TestPath_Equal(t *testing.T) {
	_, p := NewPathObj(testFile)
	_, p1 := NewPathObj(testFile)
	if !p.Equal(p1){
		t.Fatal("P.Equal test error")
	}
}

func TestPath_Read(t *testing.T) {
	_, p := NewPathObj(testFile)
	err, read := p.Read()
	if err != nil{
		t.Fatal(err)
	}
	c, err := ioutil.ReadAll(read)
	if err != nil{
		t.Fatal(err)
	}
	if !bytes.Equal(c, fileConent){
		t.Fatal("p.Read test error: read content unequal content before")
	}
}

func TestPath_Write(t *testing.T) {
	if err := os.Mkdir("./test_write", 0777); err != nil{
		t.Fatal(err)
	}
	defer os.RemoveAll("./test_write")
	_, p := NewPathObj("./test_write")
	if err := p.Write(bytes.NewReader(fileConent), "test.test"); err != nil{
		t.Fatal(err)
	}
	c, err := ioutil.ReadFile("./test_write/test.test")
	if err != nil{
		t.Fatal(err)
	}
	if !bytes.Equal(c, fileConent){
		t.Fatal("p.Read test error: write content unequal content before")
	}
	_, p = NewPathObj("./test_write/test.test")
	if err := p.Write(bytes.NewReader([]byte("this is new test file"))); err != nil{
		t.Fatal(err)
	}
	c, err = ioutil.ReadFile("./test_write/test.test")
	if err != nil{
		t.Fatal(err)
	}
	if !bytes.Equal(c, []byte("this is new test file")){
		t.Fatal("p.Read test error: write content unequal content before")
	}
}

