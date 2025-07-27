package main

import (
	"fmt"
	"sync"
)

func main() {
	var serverPool ServerPool

	for i := 0; i < 5; i++ {
		serverUrl := fmt.Sprintf("http://localhost:517%d", i)

		server := NewServer(serverUrl, i, i+6)
		serverPool.AddServer(server)
	}

	// print server details
	//for _, s := range serverPool.Servers {
	//	fmt.Printf(
	//		"URL: %s, Weight: %d, MaxConns: %d\n",
	//		s.URL.String(), s.Weight, s.MaxAllowedConns,
	//	)
	//}
	//fmt.Println()

	var wg sync.WaitGroup

	// test get next server
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			s := serverPool.GetNextServer()

			fmt.Printf(
				"index: %d, url: %s, Weight: %d, MaxConns: %d\n",
				i, s.URL.String(), s.Weight, s.MaxAllowedConns,
			)
		}()
	}

	wg.Wait()

	//http.HandleFunc("/index", func(writer http.ResponseWriter, _ *http.Request) {
	//	cookie := http.Cookie{Name: "barbara"}
	//	writer.Header().Add("Cookie", cookie.Name)
	//
	//	_, err := fmt.Fprintf(writer, "Cookie: %v\n", writer.Header().Get("Cookie"))
	//
	//	if err != nil {
	//		log.Fatal("Could not write to /index route")
	//	}
	//})
	//
	//if err := http.ListenAndServe(":8080", nil); err != nil {
	//	return
	//}
}
