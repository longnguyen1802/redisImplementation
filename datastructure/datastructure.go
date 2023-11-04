package datastructure

type DataStore interface {
    Set(key, value string)
    Get(key string) (string, bool)
    Del(key string)
    ProcessRequest(request Request) Response
}


