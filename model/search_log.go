package model

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

	from = "from"
	to   = "to"
)

type searchLog struct {
	entriesHandler func(*entry.Entry) (bool, error)

	hostByExist   map[string]bool
	eventByExist  map[string]bool
	levelByExist  map[string]bool
	timeCondition map[string]time.Time
}

func NewSearchLog(entriesHandler func(*entry.Entry) (continueRead bool, err error)) *searchLog {
	return &searchLog{
		entriesHandler: entriesHandler,
		hostByExist:    make(map[string]bool),
		eventByExist:   make(map[string]bool),
		levelByExist:   make(map[string]bool),
		timeCondition:  make(map[string]time.Time),
	}
}

func (s *searchLog) Search(req shared.SearchRequest) error {
	s.initMapOfExist(req.Host, req.Event, req.Level)

	if err := s.defineTimeForSearch(req); err != nil {
		return err
	}

	if arrayOfPath, err := s.getFilesPath(req); err != nil {
		return err
	} else if arrayOfPath != nil {
		if err := s.readFiles(req.Offset, req.Limit, arrayOfPath); err != nil {
			return err
		}
	}
	return nil
}

func (s *searchLog) initMapOfExist(host, event, level []string) {
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

func (s *searchLog) defineTimeForSearch(request shared.SearchRequest) error {
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
	s.timeCondition[from] = request.From
	s.timeCondition[to] = request.To
	return nil
}

func (s *searchLog) getFilesPath(req shared.SearchRequest) ([]string, error) {
	dirs := s.findDirs()
	if len(dirs) == 0 {
		return nil, nil
	}
	return s.findFiles(dirs, req.ModuleName)
}

func (s *searchLog) findDirs() []string {
	from := s.timeCondition[from]
	to := s.timeCondition[to]
	f := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	t := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location()).AddDate(0, 0, 1)
	dirs := make([]string, 0)
	for {
		if f.Before(t) {
			dirs = append(dirs, f.Format(dirLayout))
			f = f.AddDate(0, 0, 1)
		} else {
			dirs = append(dirs, f.Format(dirLayout))
			break
		}
	}
	return dirs
}

func (s *searchLog) findFiles(dirs []string, middleFile string) ([]string, error) {
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
			if ok, err := s.checkFileNameTime(s.timeCondition[from], s.timeCondition[to], fileTimePartName[0]); err != nil {
				return nil, err
			} else if !ok {
				continue
			}
			response = append(response, path.Join(dir, fileInfo.Name()))
		}
	}
	return response, nil
}

func (s *searchLog) readFiles(offset, limit int, files []string) error {
	for _, filePath := range files {
		if err := s.extractData(offset, limit, filePath); err != nil {
			return err
		}
	}
	return nil
}

func (s *searchLog) extractData(offset, limit int, path string) error {
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
	return s.unmarshalFile(gzipReader, offset, limit)
}

func (s *searchLog) unmarshalFile(reader io.Reader, offset, limit int) error {
	for {
		entries, err := entry.UnmarshalNext(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if s.checkEntry(entries) {
			if ok, err := s.checkTimeField(s.timeCondition[from], s.timeCondition[to], entries.Time); err != nil {
				return err
			} else if !ok {
				continue
			}
			if continueRead, err := s.entriesHandler(entries); err != nil {
				return err
			} else if !continueRead {
				return nil
			}
		}
	}
}

func (s *searchLog) checkEntry(entries *entry.Entry) bool {
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

func (s *searchLog) checkEntryField(expected map[string]bool, field string) bool {
	if len(expected) == 0 {
		return true
	} else if expected[field] {
		return true
	}
	return false
}

func (s *searchLog) checkFileNameTime(from, to time.Time, timeString string) (bool, error) {
	timeInfo, err := time.Parse(fileLayout, timeString)
	if err != nil {
		return false, err
	}
	if timeInfo.Before(to.AddDate(0, 0, 1)) && timeInfo.After(from) {
		return true, nil
	}
	return false, nil
}

func (s *searchLog) checkTimeField(from, to time.Time, timeString string) (bool, error) {
	timeInfo, err := entry.ParserTime(timeString)
	if err != nil {
		return false, err
	}
	if timeInfo.Before(to) && timeInfo.After(from) {
		return true, nil
	}
	return false, nil
}
