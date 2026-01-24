package qbit

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/hbollon/go-edlib"
	"github.com/superturkey650/go-qbittorrent/qbt"

	"media-utils/src/getenv"
)

var (
	client           *qbt.Client
	qbitOnce         sync.Once
	qbitError        error
	bookTorrents     []qbt.TorrentInfo
	bookTorrentsOnce sync.Once

	allTorrentPairs   map[string]qbt.TorrentInfo
	torrentPaths      []string
	torrentPairsOnce  sync.Once
	torrentPairsError error
)

func qbit() (*qbt.Client, error) {
	qbitOnce.Do(func() {
		Env, err := getenv.GetEnv()
		if err != nil {
			qbitError = err
			return
		}

		url, ok := Env["QBIT_BASE_URL"]
		if !ok {
			qbitError = errors.New("missing QBIT_BASE_URL from .env")
			return
		}

		user, ok := Env["QBIT_USERNAME"]
		if !ok {
			qbitError = errors.New("missing QBIT_USERNAME from .env")
			return
		}

		pass, ok := Env["QBIT_PASSWORD"]
		if !ok {
			qbitError = errors.New("missing QBIT_PASSWORD from .env")
			return
		}

		client = qbt.NewClient(url)
		err = client.Login(user, pass)
		if err != nil {
			qbitError = err
			return
		}
	})
	return client, nil
}

func emptyFilter() qbt.TorrentsOptions {
	return qbt.TorrentsOptions{
		Filter:   nil,
		Category: nil,
		Sort:     nil,
		Reverse:  nil,
		Limit:    nil,
		Offset:   nil,
		Hashes:   nil,
	}
}

func allBookTorrents() ([]qbt.TorrentInfo, error) {
	bookTorrentsOnce.Do(func() {
		c, err := qbit()
		if err != nil {
			qbitError = err
			return
		}

		filter := emptyFilter()
		filter.Category = &[]string{"books"}[0] // for *string type
		bookTorrents, err = c.Torrents(filter)
		if err != nil {
			qbitError = err
			return
		}
	})
	return bookTorrents, nil
}

func checkEbookSimilarity(path string, title string, extensions map[string]bool) (float32, error) {
	ext := filepath.Ext(filepath.Base(path))
	if !extensions[ext] {
		return 0, nil
	}
	return edlib.StringsSimilarity(title, filepath.Base(path), edlib.Lcs)
}

func MatchTorrent(title string) (*qbt.TorrentInfo, error) {
	allTorrents, err := allBookTorrents()
	if err != nil {
		return nil, err
	}

	ebookExtensions := map[string]bool{
		".epub": true, ".mobi": true, ".azw": true, ".azw3": true,
		".pdf": true, ".cbz": true, ".cbr": true, ".fb2": true,
	}

	for _, torrent := range allTorrents {
		var similarity float32

		info, err := os.Stat(torrent.ContentPath)
		if err != nil {
			//log.Printf("%s\n", err)
			continue
		}

		if info.IsDir() {
			_ = filepath.WalkDir(torrent.ContentPath, func(path string, d os.DirEntry, err error) error {
				if d.IsDir() {
					return nil
				}
				if sim, err := checkEbookSimilarity(path, title, ebookExtensions); err == nil && sim > similarity {
					similarity = sim
				}
				return nil
			})
		} else {
			similarity, _ = checkEbookSimilarity(torrent.ContentPath, title, ebookExtensions)
		}

		if similarity >= 0.85 {
			// fmt.Printf("%s matches %s\n\n", title, torrent.Name)
			return &torrent, nil
		}
	}

	return nil, nil
}

func buildTorrentPairs() error {
	torrentPairsOnce.Do(func() {
		allTorrents, err := allBookTorrents()
		if err != nil {
			torrentPairsError = err
			return
		}

		ebookExtensions := map[string]bool{
			".epub": true, ".mobi": true, ".azw": true, ".azw3": true,
			".pdf": true, ".cbz": true, ".cbr": true, ".fb2": true,
		}

		allTorrentPairs = make(map[string]qbt.TorrentInfo)

		for _, torrent := range allTorrents {
			info, err := os.Stat(torrent.ContentPath)
			if err != nil {
				continue
			}

			if info.IsDir() {
				_ = filepath.WalkDir(torrent.ContentPath, func(path string, d os.DirEntry, err error) error {
					if d.IsDir() {
						return nil
					}
					if !ebookExtensions[filepath.Ext(filepath.Base(path))] {
						return nil
					}
					allTorrentPairs[path] = torrent
					return nil
				})
			} else {
				allTorrentPairs[torrent.ContentPath] = torrent
			}
		}

		for key := range allTorrentPairs {
			torrentPaths = append(torrentPaths, key)
		}
	})
	return torrentPairsError
}

func MatchTorrentMem(title string) (*qbt.TorrentInfo, error) {
	if err := buildTorrentPairs(); err != nil {
		return nil, err
	}

	res, err := edlib.FuzzySearchThreshold(title, torrentPaths, 0.7, edlib.JaroWinkler)
	if err != nil {
		return nil, err
	}

	torrent, ok := allTorrentPairs[res]
	if !ok {
		return nil, nil
	}

	return &torrent, nil
}
