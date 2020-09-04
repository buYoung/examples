package global

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Compresslevel int

const (
	NoCompression      Compresslevel = 0
	BestSpeed          Compresslevel = 1
	BestCompression    Compresslevel = 9
	DefaultCompression Compresslevel = -1
)

type Zips struct {
	Level Compresslevel
}

func (z *Zips) initzip(zipfilename string) (*os.File, error) {
	makezip, err := os.Create(zipfilename)
	if err != nil {
		return nil, errors.Errorf("zip init err : %v", err)
	} else {
		return makezip, nil
	}
}
func (z *Zips) log(str string, debug bool) {
	if debug {
		fmt.Println(str)
	}
}

func (z *Zips) Makezip(zipfilename string, destfiles []string) error {
	zipfile, err := z.initzip(zipfilename) // 지정된 경로에 빈파일생성
	defer zipfile.Close()
	if err != nil {
		return err
	} else {
		zw := zip.NewWriter(zipfile) // 생성된 빈파일을 zip writer로 가공
		defer zw.Close()
		for _, data := range destfiles {
			destfilepath := filepath.Dir(data)
			err := filepath.Walk(data, func(paths string, info os.FileInfo, err error) error { // 대상폴더 혹은 파일을 전체 루프
				if err != nil || info.IsDir() {
					return err
				}
				Relpath, err := filepath.Rel(destfilepath, paths)
				if err != nil {
					return err
				}

				savefilepath := path.Join(filepath.SplitList(Relpath)...)

				file, err := os.Open(paths)
				defer file.Close()
				if err != nil {
					return err
				}
				zw.RegisterCompressor(zip.Deflate, func(w io.Writer) (closer io.WriteCloser, e error) { // Zip파일 압축 [파일을 zip으로 바꾼다고 압축되는게아님 별도의 compressor가있어야함]
					return flate.NewWriter(w, flate.BestCompression)
				})

				zfw, err := zw.CreateHeader(&zip.FileHeader{ // zip파일에 저장되는 파일들의 기본정보
					Name:     savefilepath,   // 폴더 + 파일명
					Method:   zip.Deflate,    // 파일이 압축을사용할경우 zip.deflate 안할경우 zip.Store
					Modified: info.ModTime(), // 파일의 수정시간
					NonUTF8:  false,          // 파일이 utf8이 아닌경우 true  보통 utf8이므로 false
				})

				if err != nil {
					return err
				}
				_, err = io.Copy(zfw, file) // zipwriter에 위에서 가공한 파일의 정보와 파일데이터를 복사

				return err
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (z *Zips) Decompresszipfile(zipfilepath string, savefolder string, debug bool) error {
	zipbyte, err := ioutil.ReadFile(zipfilepath) // ioutil library를 활용해서 file을 읽습니다.
	if err != nil {
		return err
	}
	return z.DecompresszipByte(zipbyte, savefolder, debug)
}

func (z *Zips) DecompresszipByte(file []byte, savefolder string, debug bool) error {
	zipbuf := bytes.NewReader(file)                      // zip파일의 byte정보를 io.reader로 만들어줍니다.
	archive, err := zip.NewReader(zipbuf, zipbuf.Size()) // io.reader를 zip.reader로 만들어줍니다.
	if err != nil {
		return err
	}
	archive.RegisterDecompressor(zip.Deflate, func(r io.Reader) io.ReadCloser { // zip이 압축이된걸 해제하는 compressor를 등록합니다.
		return flate.NewReader(r)
	})

	for _, ifile := range archive.File { // zip안에있는 모든파일을 루프합니다.
		savefilename := ifile.Name
		modifytime := ifile.Modified
		checkbackslash := strings.HasSuffix(savefilename, "\\")     // 백슬래쉬 "\" 찾습니다.
		savefilename = strings.Replace(savefilename, "\\", "/", -1) // 백슬래쉬 "\" 가있다면 슬래쉬로 바꿉니다.

		info := ifile.FileInfo()
		location := path.Join(savefolder, savefilename)
		switch {
		case info.IsDir() || checkbackslash: // 파일이 폴더인지 백슬래쉬가있는지
			if err := os.MkdirAll(location, info.Mode()|os.ModeDir|100); err != nil {
				return errors.Errorf("MKdir error Detail :", err)
			} else {
				z.log(fmt.Sprintln("folder create", location), debug)
			}
		case info.Mode()&os.ModeSymlink != 0: //파일이 Symlink인지
			f, err := ifile.Open()
			if err != nil {
				z.log(fmt.Sprintln("sym err", err), debug)
				return errors.Errorf("Symlink open error to zip. Detail :", err)
			} else {
				name, err := ioutil.ReadAll(f)
				if err != nil {
					z.log(fmt.Sprintln("sym read", name), debug)
					return errors.Errorf("Symlink file read error Detail :", err)
				} else {
					if err = os.Symlink(location, string(name)); err != nil {
						z.log(fmt.Sprintln("create sym", err), debug)
						return errors.Errorf("Symlink create error Detail :", err)
					} else {
						z.log(fmt.Sprintln("symlink create", location), debug)
					}
				}
			}
		default: // 일반적인 파일일경우
			f, err := ifile.Open()
			if err != nil {
				z.log(fmt.Sprintln("openfile", err), debug)
				return errors.Errorf("zip inside file open error Detail :", err)
			} else {
				filebufs, err := ioutil.ReadAll(f)
				if err != nil {
					z.log(fmt.Sprintln("fileread", err), debug)
					return errors.Errorf("file read error Detail :", err)
				} else {
					dirpath, _ := path.Split(location)
					_, err := os.Stat(dirpath)
					if os.IsNotExist(err) {
						err = os.MkdirAll(dirpath, info.Mode()|os.ModeDir|100)
						if err != nil {
							z.log(fmt.Sprintln("mkdir all", err), debug)
							return errors.Errorf("makedir error Detail :", err)

						} else {
							z.log(fmt.Sprintln("folder create", dirpath), debug)
						}
					}
					err = ioutil.WriteFile(location, filebufs, ifile.Mode()) // zip에서 읽어온 파일을 지정경로에 폴더 + 파일이름으로 저장
					if err != nil {
						z.log(fmt.Sprintln("file create ", err), debug)
						return errors.Errorf("file create error Detail :", err)
					} else {
						err = os.Chtimes(location, modifytime, modifytime) // zip에서 읽어온 파일의 수정시간을 원본상태로 변경
						if err != nil {
							z.log(fmt.Sprintln("file modifytime ", err), debug)
							return errors.Errorf("file edit modify time error Detail :", err)
						} else {
							z.log(fmt.Sprintln("file create", location), debug)
						}
					}
				}
			}
		}
		if savefilename == "" {
			continue
		}
	}
	return nil
}
