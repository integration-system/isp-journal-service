package service

import (
	"bufio"
	"compress/gzip"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-lib/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"isp-journal-service/conf"
	"isp-journal-service/shared"
	"os"
	"path"
	"strings"
	"time"
)

const (
	dirLayout  = "2006-01-02"
	fileLayout = "2006-01-02T15-04-05.000"

	bufSize   = 8 * 1024
	fileSplit = "__"
	fileEnd   = ".log"
)

type searchService struct {
	count    int
	offset   int
	response []shared.SearchResponse

	hostByExist  map[string]bool
	eventByExist map[string]bool
	levelByExist map[string]bool
}

func NewSearchService() *searchService {
	return &searchService{
		count:        0,
		offset:       0,
		hostByExist:  make(map[string]bool),
		eventByExist: make(map[string]bool),
		levelByExist: make(map[string]bool),
	}
}

func (s *searchService) Search(req shared.SearchRequest) ([]shared.SearchResponse, error) {
	s.response = make([]shared.SearchResponse, 0, req.Limit)
	s.initMapOfExist(req.Host, req.Event, req.Level)

	if err := s.defineTimeForSearch(&req); err != nil {
		return nil, err
	}

	if arrayOfPath, err := s.getFilesPath(req); err != nil {
		return nil, err
	} else if arrayOfPath != nil {
		if err := s.readFiles(req.Offset, req.Limit, arrayOfPath); err != nil {
			return nil, err
		}
	}
	return s.response, nil
}

func (s *searchService) initMapOfExist(host, event, level []string) {
	for _, value := range host {
		s.hostByExist[value] = true
	}
	for _, value := range event {
		s.eventByExist[value] = true
	}
	for _, value := range level {
		s.levelByExist[value] = true
	}
}

func (s *searchService) defineTimeForSearch(request *shared.SearchRequest) error {
	if request.From.IsZero() {
		request.From = time.Now().UTC().AddDate(0, 0, -1)
	} else {
		request.From = request.From.UTC()
	}

	if request.To.IsZero() {
		request.To = time.Now().UTC()
	} else {
		request.To = request.To.UTC()
		if request.To.Before(request.From) {
			return status.Error(codes.InvalidArgument, "expected FROM will before TO")
		}
	}
	return nil
}

func (s *searchService) getFilesPath(req shared.SearchRequest) ([]string, error) {
	dirs := s.findDirs(req.From, req.To)
	if len(dirs) == 0 {
		return nil, nil
	}
	return s.findFiles(req.From, req.To, dirs, req.ModuleName)
}

func (s *searchService) findDirs(from, to time.Time) []string {
	f := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	t := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	dirs := make([]string, 0)
	for {
		if from.Before(t) {
			dirs = append(dirs, f.Format(dirLayout))
			f = f.AddDate(0, 0, 1)
		} else {
			dirs = append(dirs, f.Format(dirLayout))
			break
		}
	}
	return dirs
}

func (s *searchService) findFiles(from, to time.Time, dirs []string, middleFile string) ([]string, error) {
	response := make([]string, 0)
	baseDir := config.GetRemote().(*conf.RemoteConfig).BaseLogDirectory
	for _, dir := range dirs {
		dir := path.Join(baseDir, dir, middleFile)
		filesInfo, err := ioutil.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return nil, err
			}
		}
		for _, fileInfo := range filesInfo {
			fileName := strings.Split(fileInfo.Name(), fileSplit)
			if !s.checkEntryField(s.hostByExist, fileName[0]) {
				continue
			}
			fileTimePartName := strings.Split(fileName[1], fileEnd)
			if ok, err := s.checkFileTimePart(from, to, fileTimePartName[0]); err != nil {
				return nil, err
			} else if !ok {
				continue
			}
			response = append(response, path.Join(dir, fileInfo.Name()))
		}
	}
	return response, nil
}

func (s *searchService) readFiles(offset, limit int, files []string) error {
	for _, filePath := range files {
		if err := s.extractData(offset, limit, filePath); err != nil {
			return err
		}
	}
	return nil
}

func (s *searchService) extractData(offset, limit int, path string) error {
	file, err := os.Open(path)
	defer func() { _ = file.Close() }()
	if err != nil {
		return err
	}
	bufReader := bufio.NewReaderSize(file, bufSize)
	gzipReader, err := gzip.NewReader(bufReader)
	defer func() { _ = gzipReader.Close() }()
	if err != nil {
		return err
	}
	return s.unmarshalFile(gzipReader, limit, offset)
}

func (s *searchService) unmarshalFile(reader io.Reader, limit, offset int) error {
	for {
		if s.count >= limit {
			return nil
		}
		response, err := entry.UnmarshalNext(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if s.checkEntry(response) {
			if s.offset == offset {
				s.count++
				s.response = append(s.response, s.convertResponse(response))
			} else {
				s.offset++
			}
		}
	}
}

func (s *searchService) checkEntry(entries *entry.Entry) bool {
	if !s.checkEntryField(s.levelByExist, entries.Level) {
		return false
	}
	if !s.checkEntryField(s.hostByExist, entries.Host) {
		return false
	}
	if !s.checkEntryField(s.eventByExist, entries.Event) {
		return false
	}
	return true
}

func (s *searchService) checkEntryField(expected map[string]bool, field string) bool {
	if len(expected) == 0 {
		return true
	} else if expected[field] {
		return true
	}
	return false
}

func (s *searchService) checkFileTimePart(from, to time.Time, name string) (bool, error) {
	test, err := time.Parse(fileLayout, name)
	if err != nil {
		return false, err
	}
	if test.Before(to) && test.After(from) {
		return true, nil
	}
	return false, nil
}

func (s *searchService) convertResponse(entries *entry.Entry) shared.SearchResponse {
	return shared.SearchResponse{
		ModuleName: entries.ModuleName,
		Host:       entries.Host,
		Event:      entries.Event,
		Level:      entries.Level,
		Time:       entries.Time,
		Request:    string(entries.Request),
		Response:   string(entries.Response),
		ErrorText:  entries.ErrorText,
	}
}
