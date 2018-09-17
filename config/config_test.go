// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//

package config

import (
	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadConfigFromFile(t *testing.T) {
	logger.SetLevel(logger.ErrorLevel)
	tests := []struct {
		name     string
		content  string
		expected *Config
	}{
		{
			name:    "CorrectFileContent",
			content: fakeFileContent(),
			expected: &Config{
				DynastyConfig{
					producers: []string{
						"1ArH9WoB9F7i6qoJiAi7McZMFVQSsBKXZR",
						"1BpXBb3uunLa9PL8MmkMtKNd3jzb5DHFkG",
					},
				},
				ConsensusConfig{
					minerAddr: "1BpXBb3uunLa9PL8MmkMtKNd3jzb5DHFkG",
					privKey: "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa7e",
				},
				NodeConfig{
					port:    5,
					seed:    "/ip4/127.0.0.1/tcp/34836/ipfs/QmPtahvwSvnSHymR5HZiSTpkm9xHymx9QLNkUjJ7mfygGs",
					dbPath:  "dbPath",
					rpcPort: 200,
				},
			},
		},
		{
			name:    "EmptySeed",
			content: noSeedContent(),
			expected: &Config{
				DynastyConfig{
					producers: []string{
						"1BpXBb3uunLa9PL8MmkMtKNd3jzb5DHFkG",
						"121yKAXeG4cw6uaGCBYjWk9yTWmMkhcoDD",
					},
				},
				ConsensusConfig{
					minerAddr: "1BpXBb3uunLa9PL8MmkMtKNd3jzb5DHFkG",
					privKey: "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa7e",
				},
				NodeConfig{
					port: 5,
					seed: "",
				},
			},
		},
		{
			name:     "WrongFileContent",
			content:  "WrongFileContent",
			expected: nil,
		},
		{
			name:    "EmptyFile",
			content: "",
			expected: &Config{
				DynastyConfig{
					producers: []string(nil),
				},
				ConsensusConfig{
					minerAddr: "",
					privKey: "",
				},
				NodeConfig{
					port: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := tt.name + "_config.conf"
			ioutil.WriteFile(filename, []byte(tt.content), 0644)
			defer os.Remove(filename)
			configContent := LoadConfigFromFile(filename)
			assert.Equal(t, tt.expected, configContent)
		})
	}

}

func fakeFileContent() string {
	return `
	dynastyConfig{
		producers: [
			"1ArH9WoB9F7i6qoJiAi7McZMFVQSsBKXZR",
			"1BpXBb3uunLa9PL8MmkMtKNd3jzb5DHFkG"
		]
	}
	consensusConfig{
					minerAddr: "1BpXBb3uunLa9PL8MmkMtKNd3jzb5DHFkG",
					privKey: "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa7e",
	}
	nodeConfig{
		port: 5
		seed: "/ip4/127.0.0.1/tcp/34836/ipfs/QmPtahvwSvnSHymR5HZiSTpkm9xHymx9QLNkUjJ7mfygGs"
		dbPath: "dbPath"
		rpcPort: 200
	}`
}

func noSeedContent() string {
	return `
	dynastyConfig{
		producers: [
			"1BpXBb3uunLa9PL8MmkMtKNd3jzb5DHFkG",
			"121yKAXeG4cw6uaGCBYjWk9yTWmMkhcoDD"
		]
	}
	consensusConfig{
						minerAddr: "1BpXBb3uunLa9PL8MmkMtKNd3jzb5DHFkG",
					privKey: "bb23d2ff19f5b16955e8a24dca34dd520980fe3bddca2b3e1b56663f0ec1aa7e",
	}
	nodeConfig{
		port: 5
	}`
}
