package wechat

import (
	"github.com/ikuiki/wwdk"
	"github.com/liangdas/mqant/gate"
	"github.com/pkg/errors"
	"time"
	"wegate/common"
)

// Uploader 上传者
type Uploader interface {
	// 获取模块名称
	getName() string
	// 获取模块描述
	getDescription() string
	// 上传文件（传入上传文件的content，返回上传好的url）
	upload(file wwdk.MediaFile)
}

// newMediaStorer 新建mediaStorer
func newMediaStorer() *mediaStorer {
	return &mediaStorer{
		uploaders:  make(map[string]Uploader),
		finishChan: make(map[string]chan string),
	}
}

type mediaStorer struct {
	uploaders  map[string]Uploader
	finishChan map[string]chan string
}

func (s *mediaStorer) Storer(file wwdk.MediaFile) (url string, err error) {
	if len(s.uploaders) == 0 {
		return "", errors.New("uploader not found")
	}
	urlChan := make(chan string)
	s.finishChan[file.FileName] = urlChan
	for _, uploader := range s.uploaders {
		go func(uploader Uploader) {
			uploader.upload(file)
		}(uploader)
	}
	// 根据文件大小设置超时时间
	// 以m为单位，再加上2m，然后除以上传带宽得到预计时间作为超时时间
	size := len(file.BinaryContent)/1000/1000 + 2
	timeoutChan := time.After(time.Duration(size) * time.Second * 2)
	select {
	case url := <-urlChan:
		delete(s.finishChan, file.FileName)
		return url, nil
	case <-timeoutChan:
		return "", errors.New("upload timeout")
	}
}

// 上传结束后调用此方法
// @param token 注册时申请到的token
// @param filename 要缓存的文件名
// @param fileurl 缓存后的url
func (s *mediaStorer) mqttUploadFinish(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeUnauthorized,
			Msg: "need login",
		}
		return
	}
	token := common.ForceString(msg["token"])
	// 检查token
	if _, ok := s.uploaders[token]; !ok {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "uploader unregistered",
		}
	}
	filename, fileurl := common.ForceString(msg["filename"]), common.ForceString(msg["fileurl"])
	// 检查mqttUploader是否符合规范
	if filename == "" || fileurl == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "filename or fileurl is empty",
		}
		return
	}
	// 检查完毕，开始给chan发信息
	if c, ok := s.finishChan[filename]; ok {
		c <- fileurl
	}
	result = common.Response{
		Ret: common.RetCodeOK,
	}
	return
}
