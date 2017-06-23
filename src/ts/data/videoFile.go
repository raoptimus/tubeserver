package data

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/raoptimus/gserv/config"
	"strconv"
	"strings"
	"time"
)

type (
	VideoFile struct {
		W    int    `bson:"W"`
		H    int    `bson:"H"`
		Path string `bson:"Path"`
		Name string `bson:"Name"`
		Hash string `bson:"Hash"`
		Ext  string `bson:"Ext"`
		Size int    `bson:"Size"`
	}
)

const TTL = 7200

func NewVideoFile(fileName string, w, h, size int) *VideoFile {
	hash := fileName[:32]
	path := fmt.Sprintf("/%s/%s/%s/", hash[:2], hash[2:7], hash[5:25])

	return &VideoFile{
		Path: path,
		Name: hash + ".mp4",
		Hash: hash,
		Ext:  "mp4",
		W:    w,
		H:    h,
		Size: size,
	}
}

func (s *VideoFile) GetViewUrl(ip string) string {
	host := config.String("ViewVideoHost", "")
	return s.getUrl(host, ip)
}

func (s *VideoFile) GetDownloadUrl(ip string) string {
	host := config.String("DownloadVideoHost", "")
	return s.getUrl(host, ip)
}

func (s *VideoFile) getUrl(host, ip string) string {
	mod := config.String("AntihotlinkMod", "")
	switch mod {
	case "modsec":
		return s.getModSecUrl(host, ip)
	default:
		return s.getUcdnUrl(host)
	}
}

func (s *VideoFile) getUcdnUrl(host string) string {
	uri := s.Path + s.Name
	creation := time.Now().UTC().Unix()
	key := config.String("AntihotlinkKey", "")
	crypt := md5.New()
	crypt.Write([]byte(uri + key + strconv.Itoa(int(creation)) + strconv.Itoa(TTL)))
	hash := hex.EncodeToString(crypt.Sum(nil))
	url := fmt.Sprintf("http://%s%s?cdn_hash=%s&cdn_creation_time=%d&cdn_ttl=%d", host,
		uri, hash, creation, TTL)

	return url
}

func (s *VideoFile) getModSecUrl(host, ip string) string {
	uri := s.Path + s.Name
	key := config.String("AntihotlinkKey", "")
	expires := strconv.FormatInt(time.Now().Unix()+TTL, 10)
	crypt := md5.New()
	crypt.Write([]byte(key + uri + expires + ip))
	hash := base64.URLEncoding.EncodeToString(crypt.Sum(nil))
	hash = strings.Replace(hash, "=", "", -1)

	return "http://" + host + "/" + hash + "/" + expires + uri + "?ip=" + ip
}
