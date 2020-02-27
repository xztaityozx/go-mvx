package factory

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/xztaityozx/cpx/cp"
	"golang.org/x/xerrors"
)

type Copyer interface {
	Copy() error
	ExistsDst() bool
	Src() string
	Dst() string
}

func GenerateLocalCopyers(srcList, dstList []string, recursive bool) ([]Copyer, error) {
	var rt []Copyer

	for _, src := range srcList {
		for _, dst := range dstList {
			res, err := dfsDir(src, dst, recursive)
			if err != nil {
				return nil, err
			}
			rt = append(rt, res...)
		}
	}

	return rt, nil
}

func dfsDir(src, dst string, recursive bool) ([]Copyer, error) {
	var rt []Copyer

	src, dst = filepath.Clean(src), filepath.Clean(dst)

	fi, err := os.Stat(src)
	if err != nil {
		return nil, xerrors.Errorf("faild to stat source file/direcory(%s): error:%v", src, err)
	}

	{
		// Dstが無い、もしくはディレクトリのときだけコピー対象にする
		dfi, err := os.Stat(dst)
		if err == nil && !dfi.IsDir() {
			return nil, xerrors.Errorf("%s is not direcotry", dst)
		}
	}

	// Sourceが普通のファイルならそのまま返す
	if fi.Mode().IsRegular() {
		return []Copyer{cp.New(
			src,
			filepath.Join(dst, fi.Name()),
		)}, nil
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return nil, xerrors.Errorf("failed to read source direcotry(%s): error:%v", src, err)
	}

	for _, entry := range entries {
		//logrus.Info(entry.Name())
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())

		// ディレクトリを再帰的に見る
		if entry.IsDir() && recursive {
			res, err := dfsDir(s, d, recursive)
			if err != nil {
				return nil, err
			}
			rt = append(rt, res...)
			// 普通のファイルならコピーの対象にする
		} else if entry.Mode().IsRegular() {
			rt = append(rt, cp.New(s, d))
			// それ以外はSkip
		} else {
			logrus.Warn("skipping... ", src)
		}
	}

	return rt, nil
}