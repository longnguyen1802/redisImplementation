package datastructure

import "sync"
type KeyValueStore struct {
    data     map[string]string
    mu       sync.RWMutex
    tasks    chan Request
    workers  int
    response chan Response
}

type Request struct {
    Command string
    Key     string
    Value   string
    Result  chan Response
}

type Response struct {
    Message string
	Successful bool
}

func NewKeyValueStoreWithWorker(workerCount int) *KeyValueStore {
    kvs := &KeyValueStore{
        data:     make(map[string]string),
        tasks:    make(chan Request),
        workers:  workerCount,
        response: make(chan Response),
    }

    // Start worker Goroutines
    for i := 0; i < workerCount; i++ {
        go kvs.worker()
    }

    return kvs
}

func (kvs *KeyValueStore) Set(key, value string) {
    kvs.tasks <- Request{Command: "SET", Key: key, Value: value}
}

func (kvs *KeyValueStore) Get(key string) (string, bool) {
    resultCh := make(chan Response)
    kvs.tasks <- Request{Command: "GET", Key: key, Result: resultCh}
    response := <-resultCh
    return response.Message, response.Message != ""
}

func (kvs *KeyValueStore) Del(key string) {
    kvs.tasks <- Request{Command: "DEL", Key: key}
}

func (kvs *KeyValueStore) ProcessRequest(request Request) Response {
    switch request.Command {
    case "SET":
        kvs.mu.Lock()
        kvs.data[request.Key] = request.Value
        kvs.mu.Unlock()
        return Response{Message: "OK"}

    case "GET":
        kvs.mu.RLock()
        value, ok := kvs.data[request.Key]
        kvs.mu.RUnlock()
        if ok {
            return Response{Message: value}
        }
        return Response{Message: "Key not found"}

    case "DEL":
        kvs.mu.Lock()
        delete(kvs.data, request.Key)
        kvs.mu.Unlock()
        return Response{Message: "OK"}

    default:
        return Response{Message: "Unknown command"}
    }
}

func (kvs *KeyValueStore) worker() {
    for req := range kvs.tasks {
        response := kvs.ProcessRequest(req)
        if req.Result != nil {
            req.Result <- response
        } else {
            kvs.response <- response
        }
    }
}