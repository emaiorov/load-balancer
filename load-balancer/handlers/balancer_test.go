package handlers

import (
	"emaiorov/load-balancer/config"
	"testing"
)

func TestLoad(t *testing.T) {

	testCases := []struct {
		name            string
		servers         []Server
		expectedLCM     uint
		expectedServers []Server
	}{
		{
			name: "Case with zero weight interpret as 1 test",
			servers: []Server{
				{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 0}},
				{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 5}},
				{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 7}},
			},
			expectedLCM: 35,
			expectedServers: []Server{
				{LoadCost: 35},
				{LoadCost: 7},
				{LoadCost: 5},
			},
		},
		{
			name: "Case with all zero weights",
			servers: []Server{
				{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 0}},
				{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 0}},
				{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 0}},
			},
			expectedLCM: 1,
			expectedServers: []Server{
				{LoadCost: 1},
				{LoadCost: 1},
				{LoadCost: 1},
			},
		},
		{
			name: "Case with high difference weights",
			servers: []Server{
				{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 5}},
				{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 50000}},
				{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 1000000}},
			},
			expectedLCM: 250000000000,
			expectedServers: []Server{
				{LoadCost: 50000000000},
				{LoadCost: 5000000},
				{LoadCost: 250000},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("NewLeastConnectionsHandlerSuccess", func(t *testing.T) {
			lcHandler := NewLeastConnectionsHandler(tc.servers)

			if lcHandler.LCM != tc.expectedLCM {
				t.Errorf("Wrong LCM: got %d, want %d", lcHandler.LCM, tc.expectedLCM)
			}
			for i, s := range lcHandler.Servers {
				if s.LoadCost != tc.expectedServers[i].LoadCost {
					t.Errorf("Wrong LoadCost: got %d, want %d", s.LoadCost, tc.expectedServers[i].LoadCost)
				}
			}
		})
	}

	testCasesDecrementScore := []struct {
		name              string
		server            Server
		expectedLoadScore uint
	}{
		{
			name:              "Load score deduction logic",
			server:            Server{LoadScore: 5, LoadCost: 1},
			expectedLoadScore: 4,
		},
		{
			name:              "Load score deduction logic",
			server:            Server{LoadScore: 25, LoadCost: 5},
			expectedLoadScore: 20,
		},
		{
			name:              "Load score deduction logic",
			server:            Server{LoadScore: 25, LoadCost: 5},
			expectedLoadScore: 20,
		},
	}

	for _, tc := range testCasesDecrementScore {
		t.Run("DecrementScore", func(t *testing.T) {
			var h Handler
			h.DecrementScore(&tc.server)

			if tc.server.LoadScore != tc.expectedLoadScore {
				t.Errorf("Wrong LoadScore: got %d, want %d", tc.server.LoadScore, tc.expectedLoadScore)
			}
		})
	}

	testCasesGetServer := []struct {
		name           string
		servers        []Server
		expectedServer Server
	}{
		{
			name: "Case with same weights",
			servers: []Server{
				{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 1}, IsAlive: true, LoadScore: 4, LoadCost: 1},
				{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 1}, IsAlive: true, LoadScore: 4, LoadCost: 1},
				{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 1}, IsAlive: true, LoadScore: 3, LoadCost: 1},
			},
			expectedServer: Server{
				ServerConfig: config.ServerConfig{Url: "http://s3"},
			},
		},
		{
			name: "Case with high difference weights",
			servers: []Server{
				{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 5}, IsAlive: true, LoadScore: 40, LoadCost: 20},
				{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 50}, IsAlive: true, LoadScore: 2, LoadCost: 2},
				{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 100}, IsAlive: true, LoadScore: 3, LoadCost: 1},
			},
			expectedServer: Server{
				ServerConfig: config.ServerConfig{Url: "http://s2"},
			},
		},
		{
			name: "Case with one alive server",
			servers: []Server{
				{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 0}, IsAlive: false},
				{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 5}, IsAlive: true},
				{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 7}, IsAlive: false},
			},
			expectedServer: Server{
				ServerConfig: config.ServerConfig{Url: "http://s2"},
			},
		},
	}

	for _, tc := range testCasesGetServer {
		t.Run("GetServerCalculatesProperServer", func(t *testing.T) {
			serversClone := make([]Server, len(tc.servers))
			for i := range tc.servers {
				newServer := tc.servers[i]

				serversClone[i] = newServer
			}

			lcHandler := NewLeastConnectionsHandler(serversClone)

			for i := 0; i < len(lcHandler.Servers); i++ {
				lcHandler.Servers[i].LoadCost = tc.servers[i].LoadCost
				lcHandler.Servers[i].LoadScore = tc.servers[i].LoadScore
			}

			server, err := lcHandler.GetServer()
			if err != nil {
				t.Errorf("Wrong expected server. Error: %s", err)
			}

			if server.Url != tc.expectedServer.Url {
				t.Errorf("Wrong Server detected: got %s, want %s", server.Url, tc.expectedServer.Url)
			}
		})
	}

	t.Run("GetServerReturnsErrorWhenNoServersConfigured", func(t *testing.T) {
		serversClone := make([]Server, 0)

		lcHandler := NewLeastConnectionsHandler(serversClone)

		_, err := lcHandler.GetServer()
		if err == nil {
			t.Errorf("Expected error that no servers found")
		}
	})

	t.Run("GetServerReturnsErrorWhenAllServersDead", func(t *testing.T) {
		servers := []Server{
			{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 0}, IsAlive: false},
			{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 5}, IsAlive: false},
			{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 7}, IsAlive: false},
		}

		lcHandler := NewLeastConnectionsHandler(servers)

		_, err := lcHandler.GetServer()
		if err == nil {
			t.Errorf("Expected error that no servers found")
		}
	})

	testCasesRoundRobin := []struct {
		name        string
		servers     []Server
		extectedUrl string
	}{
		{
			name: "Case with max weight taken as first url",
			servers: []Server{
				{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 5}, IsAlive: true},
				{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 1}, IsAlive: true},
				{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 7}, IsAlive: true},
			},
			extectedUrl: "http://s3",
		},
		{
			name: "Case with all zero weights results in first url taken",
			servers: []Server{
				{ServerConfig: config.ServerConfig{Url: "http://s1", Weight: 0}, IsAlive: true},
				{ServerConfig: config.ServerConfig{Url: "http://s2", Weight: 0}, IsAlive: true},
				{ServerConfig: config.ServerConfig{Url: "http://s3", Weight: 0}, IsAlive: true},
			},
			extectedUrl: "http://s1",
		},
	}

	for _, tc := range testCasesRoundRobin {
		t.Run("RoundRobinGetUrlHandlerSuccess", func(t *testing.T) {
			rrHandler := NewRoundRobinHandler(tc.servers)

			url, err := rrHandler.GetUrl()

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}

			if url != tc.extectedUrl {
				t.Errorf("Wrong Url: got %s, want %s", url, tc.extectedUrl)
			}
		})
	}

	testCasesCounter := []int{1, 5, 0, 50000000}
	for _, tc := range testCasesCounter {
		t.Run("SetLenth", func(t *testing.T) {
			counter := Counter{0, 0}
			counter.SetLenth(tc)

			if counter.length != tc {
				t.Errorf("Unexpected counter length: %d expected length: %d", counter.length, tc)
			}
		})
	}

	testCasesNextCounter := []int{3, 5, 7, 50000000}
	for _, tc := range testCasesNextCounter {
		t.Run("NextCounter", func(t *testing.T) {
			counter := Counter{0, 0}
			counter.SetLenth(tc)
			counter.Next()
			if counter.index != 1 {
				t.Errorf("Unexpected counter index: %d expected index: %d", counter.index, tc)
			}

			counter.Next()
			if counter.index != 2 {
				t.Errorf("Unexpected counter index: %d expected index: %d", counter.index, tc)
			}
		})
	}

	t.Run("NextCounterEdgeCases", func(t *testing.T) {
		counter := Counter{0, 0}
		counter.SetLenth(0)
		counter.Next()
		if counter.index != 0 {
			t.Errorf("Unexpected counter index: %d expected index: %d", counter.index, 0)
		}

		counter.Next()
		if counter.index != 0 {
			t.Errorf("Unexpected counter index: %d expected index: %d", counter.index, 0)
		}

		counter.SetLenth(1)
		counter.Next()
		if counter.index != 0 {
			t.Errorf("Unexpected counter index: %d expected index: %d", counter.index, 0)
		}

		counter.SetLenth(2)
		counter.Next()
		if counter.index != 1 {
			t.Errorf("Unexpected counter index: %d expected index: %d", counter.index, 1)
		}
	})

	t.Run("NextAndWrapCounter", func(t *testing.T) {
		counter := Counter{0, 0}
		counter.SetLenth(2)

		if counter.NextAndWrap() != false {
			t.Errorf("Unexpected NextAndWrap result, expected false")
		}

		if counter.index != 1 {
			t.Errorf("Unexpected counter index: %d expected index: %d", counter.index, 1)
		}

		if counter.NextAndWrap() != true {
			t.Errorf("Unexpected NextAndWrap result, expected true")
		}

		if counter.index != 0 {
			t.Errorf("Unexpected counter index: %d expected index: %d", counter.index, 0)
		}
	})

	t.Run("GetHealthUrlCase", func(t *testing.T) {
		server := Server{ServerConfig: config.ServerConfig{Url: "http://s1", Health: "/health"}}

		if server.GetHealthUrl() != "http://s1/health" {
			t.Errorf("Unexpected health check url: %s", server.GetHealthUrl())
		}
	})
}
