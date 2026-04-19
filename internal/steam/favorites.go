package steam

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

type FavoriteServer struct {
	Address    string
	IP         string
	Port       int
	SourceFile string
}

// FindFavoriteServers discovers Steam server browser favorites from local
// userdata and returns unique server addresses that can be imported into the
// app config.
func FindFavoriteServers() ([]FavoriteServer, error) {
	candidateFiles, err := favoriteFiles()
	if err != nil {
		return nil, err
	}
	if len(candidateFiles) == 0 {
		return nil, fmt.Errorf("no Steam favorites file was found")
	}

	seen := make(map[string]bool)
	var favorites []FavoriteServer

	for _, path := range candidateFiles {
		parsed, err := parseFavoriteServers(path)
		if err != nil {
			continue
		}

		for _, favorite := range parsed {
			if seen[favorite.Address] {
				continue
			}
			seen[favorite.Address] = true
			favorites = append(favorites, favorite)
		}
	}

	if len(favorites) == 0 {
		return nil, fmt.Errorf("no favorite servers were found in Steam userdata")
	}

	return favorites, nil
}

func favoriteFiles() ([]string, error) {
	var files []string
	for _, root := range steamRoots() {
		userdataPath := filepath.Join(root, "userdata")
		entries, err := os.ReadDir(userdataPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			candidate := filepath.Join(userdataPath, entry.Name(), "7", "remote", "serverbrowser_hist.vdf")
			if _, err := os.Stat(candidate); err == nil {
				files = append(files, candidate)
			}
		}
	}

	slices.Sort(files)
	return slices.Compact(files), nil
}

func steamRoots() []string {
	var roots []string

	switch runtime.GOOS {
	case "windows":
		roots = append(roots,
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Steam"),
			filepath.Join(os.Getenv("ProgramFiles"), "Steam"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Steam"),
			filepath.Join(os.Getenv("APPDATA"), "Steam"),
		)
	case "linux":
		home, _ := os.UserHomeDir()
		roots = append(roots,
			filepath.Join(home, ".local", "share", "Steam"),
			filepath.Join(home, ".steam", "steam"),
			filepath.Join(home, ".steam", "root"),
			filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".steam", "steam"),
		)
	}

	var filtered []string
	for _, root := range roots {
		if root == "" {
			continue
		}
		filtered = append(filtered, root)
	}
	return filtered
}

func parseFavoriteServers(path string) ([]FavoriteServer, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	root, err := parseVDF(string(content))
	if err != nil {
		return nil, err
	}

	filters, ok := root["Filters"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing Filters block")
	}

	favoritesBlock, ok := filters["favorites"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing favorites block")
	}

	keys := make([]int, 0, len(favoritesBlock))
	indexToKey := make(map[int]string, len(favoritesBlock))
	for key := range favoritesBlock {
		index, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		keys = append(keys, index)
		indexToKey[index] = key
	}
	slices.Sort(keys)

	var favorites []FavoriteServer
	for _, index := range keys {
		entry, ok := favoritesBlock[indexToKey[index]].(map[string]any)
		if !ok {
			continue
		}

		address, ok := entry["address"].(string)
		if !ok || strings.TrimSpace(address) == "" {
			continue
		}

		host, port, err := splitAddress(address)
		if err != nil {
			continue
		}

		favorites = append(favorites, FavoriteServer{
			Address:    address,
			IP:         host,
			Port:       port,
			SourceFile: path,
		})
	}

	return favorites, nil
}

func splitAddress(address string) (string, int, error) {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portText)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}

func parseVDF(input string) (map[string]any, error) {
	tokens, err := tokenizeVDF(input)
	if err != nil {
		return nil, err
	}

	index := 0
	return parseVDFObject(tokens, &index)
}

func parseVDFObject(tokens []string, index *int) (map[string]any, error) {
	result := make(map[string]any)

	for *index < len(tokens) {
		token := tokens[*index]
		if token == "}" {
			*index++
			return result, nil
		}

		key := token
		*index++
		if *index >= len(tokens) {
			return nil, fmt.Errorf("unexpected end of VDF after key %q", key)
		}

		next := tokens[*index]
		if next == "{" {
			*index++
			value, err := parseVDFObject(tokens, index)
			if err != nil {
				return nil, err
			}
			result[key] = value
			continue
		}

		if next == "}" {
			return nil, fmt.Errorf("unexpected closing brace after key %q", key)
		}

		result[key] = next
		*index++
	}

	return result, nil
}

func tokenizeVDF(input string) ([]string, error) {
	var tokens []string

	for index := 0; index < len(input); {
		switch input[index] {
		case ' ', '\t', '\r', '\n':
			index++
		case '/':
			if index+1 < len(input) && input[index+1] == '/' {
				for index < len(input) && input[index] != '\n' {
					index++
				}
				continue
			}
			index++
		case '{', '}':
			tokens = append(tokens, string(input[index]))
			index++
		case '"':
			index++
			start := index
			var builder strings.Builder
			closed := false

			for index < len(input) {
				if input[index] == '\\' && index+1 < len(input) {
					builder.WriteString(input[start:index])
					index++
					builder.WriteByte(input[index])
					index++
					start = index
					continue
				}
				if input[index] == '"' {
					builder.WriteString(input[start:index])
					tokens = append(tokens, builder.String())
					index++
					closed = true
					break
				}
				index++
			}

			if !closed {
				return nil, fmt.Errorf("unterminated string in VDF")
			}
		default:
			index++
		}
	}

	return tokens, nil
}
