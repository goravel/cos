package cos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
	"github.com/tencentyun/cos-go-sdk-v5"
)

/*
 * Cos COS
 * Document: https://cloud.tencent.com/document/product/436/31215
 */

type Cos struct {
	ctx             context.Context
	config          config.Config
	instance        *cos.Client
	disk            string
	url             string
	accessKeyId     string
	accessKeySecret string
}

func NewCos(ctx context.Context, config config.Config, disk string) (*Cos, error) {
	accessKeyId := config.GetString(fmt.Sprintf("filesystems.disks.%s.key", disk))
	accessKeySecret := config.GetString(fmt.Sprintf("filesystems.disks.%s.secret", disk))
	cosUrl := config.GetString(fmt.Sprintf("filesystems.disks.%s.url", disk))
	if accessKeyId == "" || accessKeySecret == "" || cosUrl == "" {
		return nil, fmt.Errorf("please set %s configuration first", disk)
	}

	u, err := url.Parse(cosUrl)
	if err != nil {
		return nil, fmt.Errorf("init %s disk error: %v", disk, err)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKeyId,
			SecretKey: accessKeySecret,
		},
	})

	return &Cos{
		ctx:             ctx,
		config:          config,
		disk:            disk,
		instance:        client,
		url:             cosUrl,
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
	}, nil
}

func (r *Cos) AllDirectories(path string) ([]string, error) {
	var directories []string
	var marker string
	validPath := validPath(path)
	opt := &cos.BucketGetOptions{
		Prefix:    validPath,
		Delimiter: "/",
		MaxKeys:   1000,
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := r.instance.Bucket.Get(context.Background(), opt)
		if err != nil {
			return nil, err
		}
		for _, commonPrefix := range v.CommonPrefixes {
			directories = append(directories, strings.ReplaceAll(commonPrefix, validPath, ""))
			subDirectories, err := r.AllDirectories(commonPrefix)
			if err != nil {
				return nil, err
			}
			for _, subDirectory := range subDirectories {
				if strings.HasSuffix(subDirectory, "/") {
					directories = append(directories, strings.ReplaceAll(commonPrefix+subDirectory, validPath, ""))
				}
			}
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}

	return directories, nil
}

func (r *Cos) AllFiles(path string) ([]string, error) {
	var files []string
	var marker string
	validPath := validPath(path)
	opt := &cos.BucketGetOptions{
		Prefix:  validPath,
		MaxKeys: 1000,
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := r.instance.Bucket.Get(r.ctx, opt)
		if err != nil {
			return nil, err
		}
		for _, content := range v.Contents {
			if !strings.HasSuffix(content.Key, "/") {
				files = append(files, strings.ReplaceAll(content.Key, validPath, ""))
			}
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}

	return files, nil
}

func (r *Cos) Copy(originFile, targetFile string) error {
	originFile = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSuffix(r.url, "/")+"/"+strings.TrimPrefix(originFile, "/"), "https://", ""), "http://", "")
	if _, _, err := r.instance.Object.Copy(r.ctx, targetFile, originFile, nil); err != nil {
		return err
	}

	return nil
}

func (r *Cos) Delete(files ...string) error {
	var obs []cos.Object
	for _, v := range files {
		obs = append(obs, cos.Object{Key: v})
	}
	opt := &cos.ObjectDeleteMultiOptions{
		Objects: obs,
		Quiet:   true,
	}

	if _, _, err := r.instance.Object.DeleteMulti(r.ctx, opt); err != nil {
		return err
	}

	return nil
}

func (r *Cos) DeleteDirectory(directory string) error {
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	var marker string
	opt := &cos.BucketGetOptions{
		Prefix:  directory,
		MaxKeys: 1000,
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		res, _, err := r.instance.Bucket.Get(r.ctx, opt)
		if err != nil {
			return err
		}
		if len(res.Contents) == 0 {
			return nil
		}

		for _, content := range res.Contents {
			_, err = r.instance.Object.Delete(r.ctx, content.Key)
			if err != nil {
				return err
			}
		}
		isTruncated = res.IsTruncated
		marker = res.NextMarker
	}

	return nil
}

func (r *Cos) Directories(path string) ([]string, error) {
	var directories []string
	var marker string
	validPath := validPath(path)
	opt := &cos.BucketGetOptions{
		Prefix:    validPath,
		Delimiter: "/",
		MaxKeys:   1000,
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := r.instance.Bucket.Get(context.Background(), opt)
		if err != nil {
			return nil, err
		}
		for _, commonPrefix := range v.CommonPrefixes {
			directory := strings.ReplaceAll(commonPrefix, validPath, "")
			if directory != "" {
				directories = append(directories, directory)
			}
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}

	return directories, nil
}

func (r *Cos) Exists(file string) bool {
	ok, err := r.instance.Object.IsExist(r.ctx, file)
	if err != nil {
		return false
	}

	return ok
}

func (r *Cos) Files(path string) ([]string, error) {
	var files []string
	var marker string
	validPath := validPath(path)
	opt := &cos.BucketGetOptions{
		Prefix:    validPath,
		Delimiter: "/",
		MaxKeys:   1000,
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := r.instance.Bucket.Get(r.ctx, opt)
		if err != nil {
			return nil, err
		}
		for _, content := range v.Contents {
			file := strings.ReplaceAll(content.Key, validPath, "")
			if file != "" {
				files = append(files, file)
			}
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}

	return files, nil
}

func (r *Cos) Get(file string) (string, error) {
	data, err := r.GetBytes(file)

	return string(data), err
}

func (r *Cos) GetBytes(file string) ([]byte, error) {
	opt := &cos.ObjectGetOptions{
		ResponseContentType: "text/html",
	}
	resp, err := r.instance.Object.Get(r.ctx, file, opt)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	return data, nil
}

func (r *Cos) LastModified(file string) (time.Time, error) {
	resp, err := r.instance.Object.Head(r.ctx, file, nil)
	if err != nil {
		return time.Time{}, err
	}

	lastModified, err := http.ParseTime(resp.Header.Get("Last-Modified"))
	if err != nil {
		return time.Time{}, err
	}

	l, err := time.LoadLocation(r.config.GetString("app.timezone"))
	if err != nil {
		return time.Time{}, err
	}

	return lastModified.In(l), nil
}

func (r *Cos) MakeDirectory(directory string) error {
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	if _, err := r.instance.Object.Put(r.ctx, directory, strings.NewReader(""), nil); err != nil {
		return err
	}

	return nil
}

func (r *Cos) MimeType(file string) (string, error) {
	resp, err := r.instance.Object.Head(r.ctx, file, nil)
	if err != nil {
		return "", err
	}

	return resp.Header.Get("Content-Type"), nil
}

func (r *Cos) Missing(file string) bool {
	return !r.Exists(file)
}

func (r *Cos) Move(oldFile, newFile string) error {
	if err := r.Copy(oldFile, newFile); err != nil {
		return err
	}

	return r.Delete(oldFile)
}

func (r *Cos) Path(file string) string {
	return file
}

func (r *Cos) Put(file string, content string) error {
	// If the file is created in a folder directly, we can't check if the folder exists.
	// So we need to create the top folder first.
	if err := r.makeDirectories(file); err != nil {
		return err
	}

	tempFile, err := r.tempFile(content)
	defer os.Remove(tempFile.Name())
	if err != nil {
		return err
	}

	_, _, err = r.instance.Object.Upload(
		r.ctx, file, tempFile.Name(), nil,
	)

	return err
}

func (r *Cos) PutFile(filePath string, source filesystem.File) (string, error) {
	return r.PutFileAs(filePath, source, str.Random(40))
}

func (r *Cos) PutFileAs(filePath string, source filesystem.File, name string) (string, error) {
	fullPath, err := fullPathOfFile(filePath, source, name)
	if err != nil {
		return "", err
	}

	// If the file is created in a folder directly, we can't check if the folder exists.
	// So we need to create the top folder first.
	if err := r.makeDirectories(str.Of(filePath).Finish("/").String()); err != nil {
		return "", err
	}

	if _, _, err := r.instance.Object.Upload(
		r.ctx, fullPath, source.File(), nil,
	); err != nil {
		return "", err
	}

	return fullPath, nil
}

func (r *Cos) Size(file string) (int64, error) {
	resp, err := r.instance.Object.Head(r.ctx, file, nil)
	if err != nil {
		return 0, err
	}

	contentLength := resp.Header.Get("Content-Length")
	contentLengthInt, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, err
	}

	return contentLengthInt, nil
}

func (r *Cos) TemporaryUrl(file string, time time.Time) (string, error) {
	// 获取预签名URL
	presignedURL, err := r.instance.Object.GetPresignedURL(r.ctx, http.MethodGet, file, r.accessKeyId, r.accessKeySecret, time.Sub(carbon.Now().StdTime()), nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func (r *Cos) WithContext(ctx context.Context) filesystem.Driver {
	if httpCtx, ok := ctx.(contractshttp.Context); ok {
		ctx = httpCtx.Context()
	}

	driver, err := NewCos(ctx, r.config, r.disk)
	if err != nil {
		color.Redf("init %s disk fail: %v\n", r.disk, err)

		return nil
	}

	return driver
}

func (r *Cos) Url(file string) string {
	objectUrl := r.instance.Object.GetObjectURL(file)

	return objectUrl.String()
}

func (r *Cos) makeDirectories(path string) error {
	folders := strings.Split(path, "/")
	for i := 1; i < len(folders); i++ {
		folder := strings.Join(folders[:i], "/")
		if err := r.MakeDirectory(folder); err != nil {
			return err
		}
	}

	return nil
}

func (r *Cos) tempFile(content string) (*os.File, error) {
	tempFile, err := os.CreateTemp(os.TempDir(), "goravel-")
	if err != nil {
		return nil, err
	}

	if _, err := tempFile.WriteString(content); err != nil {
		return nil, err
	}

	return tempFile, nil
}
